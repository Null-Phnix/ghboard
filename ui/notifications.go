// ui/notifications.go
package ui

import (
	"fmt"
	"time"

	"github.com/Null-Phnix/ghboard/api"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type notifsLoadedMsg struct {
	notifs []api.Notification
	err    error
}

type refreshTickMsg struct{}

type NotificationsModel struct {
	rest        *api.RESTClient
	notifs      []api.Notification
	filter      string // "", "PullRequest", "Issue", "CheckSuite", "Release", "Discussion"
	loading     bool
	err         error
	cursor      int
	statusMsg   string
	lastRefresh time.Time
	spinner     spinner.Model
}

func NewNotificationsModel(rest *api.RESTClient) NotificationsModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
	return NotificationsModel{rest: rest, spinner: sp}
}

func (m NotificationsModel) Init() tea.Cmd {
	return tea.Batch(m.fetchCmd(), m.tickCmd(), m.spinner.Tick)
}

func (m NotificationsModel) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		notifs, err := m.rest.ListNotifications()
		return notifsLoadedMsg{notifs: notifs, err: err}
	}
}

func (m NotificationsModel) tickCmd() tea.Cmd {
	return tea.Tick(60*time.Second, func(t time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

func (m NotificationsModel) visible() []api.Notification {
	if m.filter == "" {
		return m.notifs
	}
	var out []api.Notification
	for _, n := range m.notifs {
		if n.Subject.Type == m.filter {
			out = append(out, n)
		}
	}
	return out
}

func (m NotificationsModel) Update(msg tea.Msg) (NotificationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case notifsLoadedMsg:
		m.loading = false
		m.notifs = msg.notifs
		m.err = msg.err
		m.lastRefresh = time.Now()
		visible := m.visible()
		if m.cursor >= len(visible) {
			m.cursor = max2(0, len(visible)-1)
		}
		return m, m.tickCmd()

	case refreshTickMsg:
		return m, m.fetchCmd()

	case tea.KeyMsg:
		vis := m.visible()
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(vis)-1 {
				m.cursor++
			}
		case "g":
			m.cursor = 0
		case "G":
			m.cursor = max2(0, len(vis)-1)
		case "r":
			if m.cursor < len(vis) {
				id := vis[m.cursor].ID
				title := vis[m.cursor].Subject.Title
				m.notifs = markOneRead(m.notifs, id)
				m.statusMsg = "✓ Marked read: " + truncate(title, 40)
				go m.rest.MarkRead(id)
			}
		case "R":
			count := 0
			for _, n := range m.notifs {
				if n.Unread {
					count++
				}
			}
			for i := range m.notifs {
				m.notifs[i].Unread = false
			}
			m.statusMsg = fmt.Sprintf("✓ Marked all %d read", count)
			go m.rest.MarkAllRead()
		case "d":
			if m.cursor < len(vis) {
				id := vis[m.cursor].ID
				title := vis[m.cursor].Subject.Title
				go m.rest.MarkRead(id)
				m.notifs = dismissNotif(m.notifs, id)
				vis2 := m.visible()
				if m.cursor >= len(vis2) {
					m.cursor = max2(0, len(vis2)-1)
				}
				m.statusMsg = "✓ Dismissed: " + truncate(title, 40)
			}
		case "o":
			if m.cursor < len(vis) {
				n := vis[m.cursor]
				url := "https://github.com/notifications"
				if n.Repository.HTMLURL != "" {
					url = n.Repository.HTMLURL
				}
				openBrowser(url)
			}
		case "f":
			filters := []string{"", "PullRequest", "Issue", "CheckSuite", "Release", "Discussion"}
			cur := 0
			for i, f := range filters {
				if f == m.filter {
					cur = i
					break
				}
			}
			m.filter = filters[(cur+1)%len(filters)]
			m.cursor = 0
			if m.filter == "" {
				m.statusMsg = "Filter: all"
			} else {
				m.statusMsg = "Filter: " + m.filter
			}
		case "ctrl+r":
			m.loading = true
			m.statusMsg = ""
			return m, m.fetchCmd()
		}
	}
	return m, nil
}

var typeColors = map[string]lipgloss.Color{
	"PullRequest": "#8BE9FD",
	"Issue":       "#FF79C6",
	"CheckSuite":  "#FFB86C",
	"Release":     "#50FA7B",
	"Discussion":  "#BD93F9",
}

var typeShort = map[string]string{
	"PullRequest": "PR ",
	"Issue":       "ISS",
	"CheckSuite":  "CI ",
	"Release":     "REL",
	"Discussion":  "DSC",
}

func typeBadge(t string) string {
	label := typeShort[t]
	if label == "" {
		label = truncate(t, 3)
	}
	color := typeColors[t]
	if color == "" {
		color = "#888888"
	}
	return lipgloss.NewStyle().
		Background(color).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 1).
		Render(label)
}

func reasonBadge(reason string) string {
	label := reason
	color := lipgloss.Color("#555555")
	switch reason {
	case "mention":
		label = "@mention"
		color = "#FF79C6"
	case "review_requested":
		label = "review"
		color = "#8BE9FD"
	case "assign":
		label = "assigned"
		color = "#FFB86C"
	case "subscribed":
		label = "subscribed"
		color = "#555555"
	case "comment":
		label = "comment"
		color = "#888888"
	case "ci_activity":
		label = "CI"
		color = "#FFB86C"
	}
	return lipgloss.NewStyle().Foreground(color).Render(label)
}

func (m NotificationsModel) View(w, h int) string {
	unread := 0
	for _, n := range m.notifs {
		if n.Unread {
			unread++
		}
	}

	filterStr := ""
	if m.filter != "" {
		filterStr = "  [" + m.filter + "]"
	}
	unreadBadge := ""
	if unread > 0 {
		unreadBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("#FF5555")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("%d", unread))
	}

	header := lipgloss.NewStyle().Bold(true).Padding(1, 4).
		Render("🔔  Notifications " + unreadBadge + filterStr)

	if m.loading {
		return header + "\n\n" + lipgloss.NewStyle().Padding(0, 4).
			Foreground(lipgloss.Color("#888888")).Render(m.spinner.View() + " Loading…")
	}
	if m.err != nil {
		return header + "\n\n" + lipgloss.NewStyle().Padding(0, 4).
			Foreground(lipgloss.Color("#FF5555")).
			Render(fmt.Sprintf("✗  %v\n\nctrl+r to retry", m.err))
	}

	vis := m.visible()
	if len(vis) == 0 {
		msg := "No notifications"
		if m.filter != "" {
			msg = fmt.Sprintf("No %s notifications  (f to clear filter)", m.filter)
		}
		return header + "\n\n" + lipgloss.NewStyle().Padding(0, 4).
			Foreground(lipgloss.Color("#555555")).Render(msg)
	}

	listHeight := h - 6
	if listHeight < 1 {
		listHeight = 1
	}
	start := 0
	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}

	// Group by repo
	rows := ""
	lastRepo := ""
	shownCount := 0
	for i := start; i < len(vis) && shownCount < listHeight; i++ {
		n := vis[i]

		if n.Repository.FullName != lastRepo {
			lastRepo = n.Repository.FullName
			if shownCount > 0 {
				rows += "\n"
			}
			rows += lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Bold(true).
				Padding(0, 4).
				Render(lastRepo) + "\n"
			shownCount++
		}

		unreadDot := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("○")
		if n.Unread {
			unreadDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#39d353")).Render("●")
		}

		prefix := "   "
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
		if i == m.cursor {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("#39d353")).Render("  ▶")
			titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
		}
		if n.Unread {
			titleStyle = titleStyle.Bold(true)
		}

		maxTitle := w - 30
		if maxTitle < 20 {
			maxTitle = 20
		}
		title := titleStyle.Render(truncate(n.Subject.Title, maxTitle))

		updatedAt := ""
		if n.UpdatedAt != "" {
			t, err := time.Parse(time.RFC3339, n.UpdatedAt)
			if err == nil {
				updatedAt = "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).
					Render(relativeTime(t))
			}
		}

		row := fmt.Sprintf("%s %s %s %s %s%s\n",
			prefix,
			unreadDot,
			typeBadge(n.Subject.Type),
			title,
			reasonBadge(n.Reason),
			updatedAt,
		)
		rows += row
		shownCount++
	}

	// Scroll info
	scrollInfo := ""
	if len(vis) > listHeight {
		pct := 100 * (m.cursor + 1) / len(vis)
		scrollInfo = fmt.Sprintf("  %d/%d (%d%%)", m.cursor+1, len(vis), pct)
	}

	// Status bar
	status := m.statusMsg
	if status == "" {
		status = "r: read  •  R: all read  •  d: dismiss  •  o: open  •  f: filter  •  ctrl+r: refresh"
	}
	refreshedStr := ""
	if !m.lastRefresh.IsZero() {
		refreshedStr = "  • updated " + relativeTime(m.lastRefresh)
	}
	statusBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 4).
		Render(status + scrollInfo + refreshedStr)

	return header + rows + "\n" + statusBar
}

func markOneRead(notifs []api.Notification, id string) []api.Notification {
	for i := range notifs {
		if notifs[i].ID == id {
			notifs[i].Unread = false
		}
	}
	return notifs
}

func dismissNotif(notifs []api.Notification, id string) []api.Notification {
	out := make([]api.Notification, 0, len(notifs))
	for _, n := range notifs {
		if n.ID != id {
			out = append(out, n)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

