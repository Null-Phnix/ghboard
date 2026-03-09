// ui/stars.go
package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/Null-Phnix/ghboard/api"
	"github.com/Null-Phnix/ghboard/store"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type starsLoadedMsg struct {
	repos []api.StarredRepo
	err   error
}

type unstarDoneMsg struct {
	fullName string
	err      error
}

type StarsModel struct {
	rest       *api.RESTClient
	tags       *store.Store
	repos      []api.StarredRepo
	filtered   []api.StarredRepo
	loading    bool
	err        error
	cursor     int
	search     string
	searching  bool
	filterTag  string
	filterLang string
	confirming bool
	tagEditing bool
	tagInput   string
	statusMsg  string
	spinner    spinner.Model
}

func NewStarsModel(rest *api.RESTClient, tags *store.Store) StarsModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C"))
	return StarsModel{rest: rest, tags: tags, spinner: sp}
}

func (m StarsModel) Init() tea.Cmd {
	if m.repos != nil {
		return nil
	}
	m.loading = true
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
		var all []api.StarredRepo
		for page := 1; ; page++ {
			repos, err := m.rest.ListStars(page)
			if err != nil {
				return starsLoadedMsg{err: err}
			}
			all = append(all, repos...)
			if len(repos) < 100 {
				break
			}
		}
		return starsLoadedMsg{repos: all}
		},
	)
}

func (m *StarsModel) allTags() []string {
	seen := make(map[string]struct{})
	for _, r := range m.repos {
		for _, t := range m.tags.Get(r.FullName) {
			if t != "" {
				seen[t] = struct{}{}
			}
		}
	}
	tags := make([]string, 0, len(seen))
	for t := range seen {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	return tags
}

func (m *StarsModel) applyFilter() {
	search := strings.ToLower(m.search)
	m.filtered = nil
	for _, r := range m.repos {
		if m.filterTag != "" {
			tags := m.tags.Get(r.FullName)
			found := false
			for _, t := range tags {
				if t == m.filterTag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if m.filterLang != "" && !strings.EqualFold(r.Language, m.filterLang) {
			continue
		}
		if search != "" {
			haystack := strings.ToLower(r.FullName + " " + r.Description + " " + r.Language)
			if !strings.Contains(haystack, search) {
				continue
			}
		}
		m.filtered = append(m.filtered, r)
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max2(0, len(m.filtered)-1)
	}
}

func (m StarsModel) Update(msg tea.Msg) (StarsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case starsLoadedMsg:
		m.loading = false
		m.repos = msg.repos
		m.err = msg.err
		m.applyFilter()
		return m, nil

	case unstarDoneMsg:
		if msg.err != nil {
			m.statusMsg = "✗ Unstar failed: " + msg.err.Error()
		} else {
			m.statusMsg = "✓ Unstarred " + msg.fullName
		}
		return m, nil

	case tea.KeyMsg:
		// Tag editing mode
		if m.tagEditing {
			switch msg.String() {
			case "enter":
				if m.cursor < len(m.filtered) {
					repo := m.filtered[m.cursor]
					rawTags := strings.Split(m.tagInput, ",")
					var cleaned []string
					for _, t := range rawTags {
						t = strings.TrimSpace(t)
						if t != "" {
							cleaned = append(cleaned, t)
						}
					}
					m.tags.Set(repo.FullName, cleaned)
					m.tags.Save()
					m.statusMsg = fmt.Sprintf("✓ Tags saved for %s", repo.FullName)
				}
				m.tagEditing = false
				m.tagInput = ""
			case "esc":
				m.tagEditing = false
				m.tagInput = ""
			case "backspace":
				if len(m.tagInput) > 0 {
					m.tagInput = m.tagInput[:len(m.tagInput)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.tagInput += msg.String()
				}
			}
			return m, nil
		}

		// Search mode
		if m.searching {
			switch msg.String() {
			case "enter", "esc":
				m.searching = false
			case "backspace":
				if len(m.search) > 0 {
					m.search = m.search[:len(m.search)-1]
					m.applyFilter()
				}
			default:
				if len(msg.String()) == 1 {
					m.search += msg.String()
					m.applyFilter()
				}
			}
			return m, nil
		}

		// Confirm unstar
		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				m.confirming = false
				if m.cursor < len(m.filtered) {
					repo := m.filtered[m.cursor]
					parts := strings.SplitN(repo.FullName, "/", 2)
					m.repos = removeRepo(m.repos, repo.FullName)
					m.tags.Remove(repo.FullName)
					m.applyFilter()
					if len(parts) == 2 {
						owner, name := parts[0], parts[1]
						rest := m.rest
						fullName := repo.FullName
						return m, func() tea.Msg {
							err := rest.Unstar(owner, name)
							return unstarDoneMsg{fullName: fullName, err: err}
						}
					}
				}
			default:
				m.confirming = false
				m.statusMsg = "Unstar cancelled"
			}
			return m, nil
		}

		// Normal mode
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "g":
			m.cursor = 0
		case "G":
			m.cursor = max2(0, len(m.filtered)-1)
		case "/":
			m.searching = true
			m.search = ""
			m.applyFilter()
		case "esc":
			if m.search != "" {
				m.search = ""
				m.applyFilter()
			} else if m.filterTag != "" {
				m.filterTag = ""
				m.applyFilter()
				m.statusMsg = "Filter cleared"
			}
		case "t":
			if m.cursor < len(m.filtered) {
				existing := m.tags.Get(m.filtered[m.cursor].FullName)
				m.tagInput = strings.Join(existing, ", ")
				m.tagEditing = true
			}
		case "u":
			if m.cursor < len(m.filtered) {
				m.confirming = true
			}
		case "o":
			if m.cursor < len(m.filtered) {
				openBrowser(m.filtered[m.cursor].HTMLURL)
			}
		case "f":
			tags := m.allTags()
			if len(tags) == 0 {
				m.statusMsg = "No tags yet — press t to add some"
			} else {
				// Build cycle list: ["", tag1, tag2, ...]
				cycle := append([]string{""}, tags...)
				idx := 0
				for i, v := range cycle {
					if v == m.filterTag {
						idx = i
						break
					}
				}
				idx = (idx + 1) % len(cycle)
				m.filterTag = cycle[idx]
				m.filterLang = ""
				m.applyFilter()
				if m.filterTag == "" {
					m.statusMsg = "Filter cleared"
				} else {
					m.statusMsg = "Filter: #" + m.filterTag
				}
			}
		case "ctrl+r":
			m.repos = nil
			m.statusMsg = ""
			return m, m.Init()
		}
	}
	return m, nil
}

var tagPalette = []lipgloss.Color{
	"#FF79C6", "#8BE9FD", "#50FA7B", "#FFB86C", "#BD93F9", "#FF5555", "#F1FA8C",
}

func renderTag(tag string) string {
	h := 0
	for _, c := range tag {
		h += int(c)
	}
	color := tagPalette[h%len(tagPalette)]
	return lipgloss.NewStyle().
		Background(color).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1).
		Render(tag)
}

func langDot(lang string) string {
	// Common language colors (GitHub-style)
	colors := map[string]string{
		"Go":         "#00ADD8",
		"JavaScript": "#F1E05A",
		"TypeScript": "#3178C6",
		"Python":     "#3572A5",
		"Rust":       "#DEA584",
		"C++":        "#F34B7D",
		"C":          "#555555",
		"Java":       "#B07219",
		"Ruby":       "#701516",
		"Shell":      "#89E051",
		"Swift":      "#FA7343",
		"Kotlin":     "#A97BFF",
		"HTML":       "#E34C26",
		"CSS":        "#563D7C",
		"Zig":        "#EC915C",
		"Lua":        "#000080",
		"C#":         "#178600",
		"PHP":        "#4F5D95",
		"Dart":       "#00B4AB",
		"Elixir":     "#6E4A7E",
	}
	color, ok := colors[lang]
	if !ok {
		color = "#888888"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("●")
}

func (m StarsModel) View(w, h int) string {
	if m.loading {
		return lipgloss.NewStyle().Padding(2, 4).
			Foreground(lipgloss.Color("#888888")).
			Render(m.spinner.View() + " Loading your stars…")
	}
	if m.err != nil {
		return lipgloss.NewStyle().Padding(2, 4).
			Foreground(lipgloss.Color("#FF5555")).
			Render(fmt.Sprintf("✗  %v\n\nctrl+r to retry", m.err))
	}

	total := len(m.repos)
	showing := len(m.filtered)

	countStr := fmt.Sprintf("⭐  %d starred", total)
	if showing != total {
		countStr += fmt.Sprintf("  (showing %d)", showing)
	}
	headerContent := lipgloss.NewStyle().Bold(true).Render(countStr)
	if m.filterTag != "" {
		tagLabel := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")).
			Bold(true).
			Render("  #" + m.filterTag)
		headerContent += tagLabel
	}
	header := lipgloss.NewStyle().Padding(1, 4).Render(headerContent)

	// Search bar
	if m.searching || m.search != "" {
		searchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
		cursor := ""
		if m.searching {
			cursor = "█"
		}
		header += "\n" + lipgloss.NewStyle().Padding(0, 4).
			Render(searchStyle.Render("/") + " " + m.search + cursor)
	}

	// Tag editing overlay
	if m.tagEditing && m.cursor < len(m.filtered) {
		repo := m.filtered[m.cursor]
		existing := m.tags.Get(repo.FullName)
		existingStr := ""
		if len(existing) > 0 {
			for _, t := range existing {
				existingStr += " " + renderTag(t)
			}
			existingStr = "\n  Current: " + existingStr
		}
		prompt := lipgloss.NewStyle().Padding(1, 4).Render(
			lipgloss.NewStyle().Bold(true).Render("Tag "+repo.FullName) +
				existingStr +
				"\n\n  Enter comma-separated tags:\n  " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Render(m.tagInput+"█") +
				"\n\n  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).
				Render("Enter to save  •  Esc to cancel"),
		)
		return header + "\n" + prompt
	}

	// Confirm unstar overlay
	if m.confirming && m.cursor < len(m.filtered) {
		repo := m.filtered[m.cursor]
		prompt := lipgloss.NewStyle().Padding(1, 4).
			Foreground(lipgloss.Color("#FF5555")).
			Render(fmt.Sprintf("Unstar %s?\n\n  y to confirm  •  any other key to cancel", repo.FullName))
		return header + "\n" + prompt
	}

	// List
	listHeight := h - 8
	if listHeight < 1 {
		listHeight = 1
	}
	start := 0
	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}

	rows := ""
	for i := start; i < len(m.filtered) && i < start+listHeight; i++ {
		r := m.filtered[i]
		repoTags := m.tags.Get(r.FullName)
		selected := i == m.cursor

		prefix := "  "
		nameStyle := lipgloss.NewStyle()
		if selected {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("#39d353")).Render("▶ ")
			nameStyle = nameStyle.Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
		} else {
			nameStyle = nameStyle.Foreground(lipgloss.Color("#CCCCCC"))
		}

		name := nameStyle.Render(r.FullName)

		lang := ""
		if r.Language != "" {
			lang = " " + langDot(r.Language) + " " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(r.Language)
		}

		stars := lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")).
			Render(fmt.Sprintf(" ⭐%d", r.StargazersCount))

		tagStr := ""
		for _, t := range repoTags {
			tagStr += " " + renderTag(t)
		}

		desc := ""
		if r.Description != "" {
			d := r.Description
			maxDesc := w - 10
			if maxDesc < 20 {
				maxDesc = 20
			}
			if len(d) > maxDesc {
				d = d[:maxDesc-3] + "..."
			}
			desc = "\n    " + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(d)
		}

		row := prefix + name + lang + stars + tagStr + desc + "\n"
		rows += row
	}

	// Scroll indicator
	scrollInfo := ""
	if len(m.filtered) > listHeight {
		pct := 100 * (m.cursor + 1) / len(m.filtered)
		scrollInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).
			Render(fmt.Sprintf(" %d/%d (%d%%)", m.cursor+1, len(m.filtered), pct))
	}

	// Status message
	status := m.statusMsg
	if status == "" {
		status = "/: search  •  t: tags  •  f: cycle tag filter  •  u: unstar  •  o: open  •  ctrl+r: refresh"
	}
	statusBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 4).
		Render(status + scrollInfo)

	return header + "\n" + lipgloss.NewStyle().Padding(0, 2).Render(rows) + "\n" + statusBar
}

func removeRepo(repos []api.StarredRepo, fullName string) []api.StarredRepo {
	out := make([]api.StarredRepo, 0, len(repos))
	for _, r := range repos {
		if r.FullName != fullName {
			out = append(out, r)
		}
	}
	return out
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		cmd = "start"
	}
	exec.Command(cmd, url).Start()
}
