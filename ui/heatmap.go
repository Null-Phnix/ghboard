// ui/heatmap.go
package ui

import (
	"fmt"
	"time"

	"github.com/Null-Phnix/ghboard/api"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type contribLoadedMsg struct {
	data *api.ContributionData
	err  error
}

type HeatmapModel struct {
	gql     *api.GraphQLClient
	year    int
	data    *api.ContributionData
	loading bool
	err     error
	cursorX int // week index
	cursorY int // day index (0=Sun)
}

func NewHeatmapModel(gql *api.GraphQLClient) HeatmapModel {
	return HeatmapModel{
		gql:  gql,
		year: time.Now().Year(),
	}
}

func (m HeatmapModel) Init() tea.Cmd {
	m.loading = true
	year := m.year
	return func() tea.Msg {
		data, err := m.gql.FetchContributions(year)
		return contribLoadedMsg{data: data, err: err}
	}
}

func (m HeatmapModel) Update(msg tea.Msg) (HeatmapModel, tea.Cmd) {
	switch msg := msg.(type) {
	case contribLoadedMsg:
		m.loading = false
		m.data = msg.data
		m.err = msg.err
		// Place cursor at today's position if viewing current year
		if m.data != nil && m.year == time.Now().Year() {
			today := time.Now()
			todayStr := today.Format("2006-01-02")
			for wi, week := range m.data.Weeks {
				for _, d := range week.ContributionDays {
					if d.Date == todayStr {
						m.cursorX = wi
						m.cursorY = d.Weekday
					}
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.cursorX > 0 {
				m.cursorX--
			}
		case "right", "l":
			if m.data != nil && m.cursorX < len(m.data.Weeks)-1 {
				m.cursorX++
			}
		case "up", "k":
			if m.cursorY > 0 {
				m.cursorY--
			}
		case "down", "j":
			if m.cursorY < 6 {
				m.cursorY++
			}
		case "[":
			m.year--
			m.data = nil
			m.loading = true
			return m, m.Init()
		case "]":
			if m.year < time.Now().Year() {
				m.year++
				m.data = nil
				m.loading = true
				return m, m.Init()
			}
		case "ctrl+r":
			m.data = nil
			m.loading = true
			return m, m.Init()
		}
	}
	return m, nil
}

var heatColors = []lipgloss.Color{
	"#161b22", // 0 contributions
	"#0e4429", // 1-3
	"#006d32", // 4-6
	"#26a641", // 7-9
	"#39d353", // 10+
}

func countToColor(count int) lipgloss.Color {
	switch {
	case count == 0:
		return heatColors[0]
	case count <= 3:
		return heatColors[1]
	case count <= 6:
		return heatColors[2]
	case count <= 9:
		return heatColors[3]
	default:
		return heatColors[4]
	}
}

func countToBlock(count int) string {
	switch {
	case count == 0:
		return "░"
	case count <= 3:
		return "▒"
	case count <= 6:
		return "▓"
	default:
		return "█"
	}
}

// monthLabels computes month abbreviations positioned above the week columns.
func monthLabels(weeks []api.ContributionWeek) string {
	if len(weeks) == 0 {
		return ""
	}
	labels := make([]byte, len(weeks)*2)
	for i := range labels {
		labels[i] = ' '
	}
	lastMonth := -1
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	for wi, week := range weeks {
		if len(week.ContributionDays) == 0 {
			continue
		}
		t, err := time.Parse("2006-01-02", week.ContributionDays[0].Date)
		if err != nil {
			continue
		}
		m := int(t.Month()) - 1
		if m != lastMonth {
			lastMonth = m
			pos := wi * 2
			lbl := months[m]
			for j, c := range []byte(lbl) {
				if pos+j < len(labels) {
					labels[pos+j] = c
				}
			}
		}
	}
	return string(labels)
}

func (m HeatmapModel) View(w, h int) string {
	if m.loading {
		return lipgloss.NewStyle().Padding(2, 4).
			Foreground(lipgloss.Color("#888888")).
			Render("⟳  Loading contributions…")
	}
	if m.err != nil {
		return lipgloss.NewStyle().Padding(2, 4).
			Foreground(lipgloss.Color("#FF5555")).
			Render(fmt.Sprintf("✗  %v\n\nctrl+r to retry", m.err))
	}
	if m.data == nil {
		return ""
	}

	// Year navigation header
	prevArrow := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("[ prev year ]")
	nextArrow := ""
	if m.year < time.Now().Year() {
		nextArrow = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("[ next year ]")
	}
	yearStr := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#39d353")).
		Render(fmt.Sprintf("%d", m.year))

	header := lipgloss.NewStyle().
		Bold(true).
		Padding(1, 4).
		Render(fmt.Sprintf("%s — %d contributions in %s   %s  %s",
			m.data.Login, m.data.TotalContributions, yearStr, prevArrow, nextArrow))

	// Month labels
	monthRow := lipgloss.NewStyle().Padding(0, 4).
		Foreground(lipgloss.Color("#555555")).
		Render("    " + monthLabels(m.data.Weeks))

	// Grid
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	grid := ""
	for day := 0; day < 7; day++ {
		row := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).
			Render(fmt.Sprintf("%s ", days[day]))

		for weekIdx, week := range m.data.Weeks {
			color := heatColors[0]
			block := "░"
			found := false

			for _, d := range week.ContributionDays {
				if d.Weekday == day {
					color = countToColor(d.ContributionCount)
					block = countToBlock(d.ContributionCount)
					found = true
					break
				}
			}
			if !found {
				// Empty cell (start/end of year padding)
				row += "  "
				continue
			}

			style := lipgloss.NewStyle().Foreground(color)
			if weekIdx == m.cursorX && day == m.cursorY {
				style = lipgloss.NewStyle().
					Background(lipgloss.Color("#FFFFFF")).
					Foreground(lipgloss.Color("#000000"))
				block = "█"
			}
			row += style.Render(block + " ")
		}
		grid += row + "\n"
	}

	// Hovered day detail
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Padding(0, 4).
		Render("No data for selected cell")
	if m.cursorX < len(m.data.Weeks) {
		week := m.data.Weeks[m.cursorX]
		for _, d := range week.ContributionDays {
			if d.Weekday == m.cursorY {
				bar := buildMiniBar(d.ContributionCount)
				status = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#888888")).
					Padding(0, 4).
					Render(fmt.Sprintf("📅  %s  —  %s  %d contributions",
						d.Date, bar, d.ContributionCount))
				break
			}
		}
	}

	// Legend
	legend := lipgloss.NewStyle().Padding(0, 4).Render(
		lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("Less ") +
			lipgloss.NewStyle().Foreground(heatColors[0]).Render("░ ") +
			lipgloss.NewStyle().Foreground(heatColors[1]).Render("▒ ") +
			lipgloss.NewStyle().Foreground(heatColors[2]).Render("▓ ") +
			lipgloss.NewStyle().Foreground(heatColors[3]).Render("█ ") +
			lipgloss.NewStyle().Foreground(heatColors[4]).Render("█ ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("More") +
			"   [ / ] year  •  arrows / hjkl navigate  •  ctrl+r refresh",
	)

	gridBlock := lipgloss.NewStyle().Padding(0, 4).Render(grid)

	return header + "\n" + monthRow + "\n" + gridBlock + "\n" + status + "\n\n" + legend
}

// buildMiniBar draws a tiny inline bar chart proportional to the count (max ~10).
func buildMiniBar(count int) string {
	if count == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("▏")
	}
	bars := count
	if bars > 10 {
		bars = 10
	}
	color := countToColor(count)
	block := ""
	for i := 0; i < bars; i++ {
		block += "█"
	}
	return lipgloss.NewStyle().Foreground(color).Render(block)
}
