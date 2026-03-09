# ghboard Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a feature-full GitHub terminal dashboard with three tabs: Contribution Heatmap, Star Manager, and Notification Center.

**Architecture:** Single Go binary using Bubble Tea for the TUI and Lip Gloss for styling. GitHub GraphQL API powers the heatmap; REST API powers stars and notifications. Local tags stored in JSON at `~/.config/ghboard/tags.json`.

**Tech Stack:** Go 1.26, Bubble Tea, Lip Gloss, Bubbles (list/spinner/textinput), GitHub REST + GraphQL APIs, encoding/json for persistence.

---

## Pre-flight

```bash
cd /Users/josii/Desktop/ghboard
go version  # should be 1.26.x
```

---

### Task 1: Project Scaffold

**Files:**
- Create: `main.go`
- Create: `go.mod`
- Create: `go.sum` (auto-generated)
- Create: `ui/app.go`
- Create: `ui/heatmap.go`
- Create: `ui/stars.go`
- Create: `ui/notifications.go`
- Create: `api/graphql.go`
- Create: `api/rest.go`
- Create: `store/tags.go`
- Create: `config/config.go`

**Step 1: Initialize Go module**

```bash
cd /Users/josii/Desktop/ghboard
go mod init github.com/Null-Phnix/ghboard
```

Expected: `go.mod` created with `module github.com/Null-Phnix/ghboard`

**Step 2: Install dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
```

**Step 3: Create directory structure**

```bash
mkdir -p ui api store config
```

**Step 4: Create placeholder main.go**

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Null-Phnix/ghboard/ui"
	"github.com/Null-Phnix/ghboard/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewApp(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 5: Verify it compiles (will fail until stubs exist — that's fine)**

```bash
go build ./... 2>&1 | head -20
```

**Step 6: Commit**

```bash
git init
git add .
git commit -m "feat: scaffold ghboard project"
```

---

### Task 2: Config Module

**Files:**
- Create: `config/config.go`
- Create: `config/config_test.go`

**Step 1: Write the failing test**

```go
// config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token-123")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "test-token-123" {
		t.Errorf("expected token 'test-token-123', got %q", cfg.Token)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{"token":"file-token-456"}`), 0600)
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GHBOARD_CONFIG", cfgPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "file-token-456" {
		t.Errorf("expected 'file-token-456', got %q", cfg.Token)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./config/... -v
```

Expected: FAIL — `Load` not defined

**Step 3: Implement config.go**

```go
// config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Token string `json:"token"`
}

func configPath() string {
	if p := os.Getenv("GHBOARD_CONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ghboard", "config.json")
}

func Load() (*Config, error) {
	// 1. Env var takes priority
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return &Config{Token: token}, nil
	}

	// 2. Config file
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil // no token yet, first-run flow handles it
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./config/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add config/
git commit -m "feat: config module with env + file loading"
```

---

### Task 3: Tag Store

**Files:**
- Create: `store/tags.go`
- Create: `store/tags_test.go`

**Step 1: Write failing tests**

```go
// store/tags_test.go
package store

import (
	"path/filepath"
	"testing"
)

func TestAddAndGetTags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tags.json")
	s := New(path)

	s.Set("owner/repo", []string{"tools", "ai"})
	tags := s.Get("owner/repo")
	if len(tags) != 2 || tags[0] != "tools" || tags[1] != "ai" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tags.json")
	s1 := New(path)
	s1.Set("owner/repo", []string{"reference"})
	s1.Save()

	s2 := New(path)
	tags := s2.Get("owner/repo")
	if len(tags) != 1 || tags[0] != "reference" {
		t.Errorf("tags not persisted: %v", tags)
	}
}
```

**Step 2: Run to verify failure**

```bash
go test ./store/... -v
```

**Step 3: Implement store/tags.go**

```go
// store/tags.go
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	path string
	mu   sync.RWMutex
	data map[string][]string // repo full name -> tags
}

func New(path string) *Store {
	s := &Store{
		path: path,
		data: make(map[string][]string),
	}
	s.load()
	return s
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ghboard", "tags.json")
}

func (s *Store) Get(repo string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[repo]
}

func (s *Store) Set(repo string, tags []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[repo] = tags
}

func (s *Store) Remove(repo string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, repo)
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600)
}

func (s *Store) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &s.data)
}
```

**Step 4: Run tests**

```bash
go test ./store/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add store/
git commit -m "feat: local tag store with JSON persistence"
```

---

### Task 4: GitHub REST API Client

**Files:**
- Create: `api/rest.go`
- Create: `api/rest_test.go`

**Step 1: Write failing tests**

```go
// api/rest_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListStars(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/starred" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]StarredRepo{
			{FullName: "owner/repo", Description: "test", Language: "Go", StargazersCount: 42},
		})
	}))
	defer srv.Close()

	client := NewRESTClient("test-token", srv.URL)
	repos, err := client.ListStars(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || repos[0].FullName != "owner/repo" {
		t.Errorf("unexpected repos: %v", repos)
	}
}

func TestListNotifications(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Notification{
			{ID: "1", Unread: true, Reason: "mention", Subject: Subject{Title: "Test PR", Type: "PullRequest"}},
		})
	}))
	defer srv.Close()

	client := NewRESTClient("test-token", srv.URL)
	notifs, err := client.ListNotifications()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifs) != 1 || notifs[0].ID != "1" {
		t.Errorf("unexpected notifications: %v", notifs)
	}
}
```

**Step 2: Run to verify failure**

```bash
go test ./api/... -v
```

**Step 3: Implement api/rest.go**

```go
// api/rest.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RESTClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewRESTClient(token, baseURL string) *RESTClient {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &RESTClient{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *RESTClient) get(path string, out any) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid token")
	}
	if resp.StatusCode == 403 {
		return fmt.Errorf("rate limited (resets: %s)", resp.Header.Get("X-RateLimit-Reset"))
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// --- Stars ---

type StarredRepo struct {
	FullName        string `json:"full_name"`
	Description     string `json:"description"`
	Language        string `json:"language"`
	StargazersCount int    `json:"stargazers_count"`
	HTMLURL         string `json:"html_url"`
	UpdatedAt       string `json:"updated_at"`
}

func (c *RESTClient) ListStars(page int) ([]StarredRepo, error) {
	var repos []StarredRepo
	path := fmt.Sprintf("/user/starred?per_page=100&page=%d", page)
	return repos, c.get(path, &repos)
}

func (c *RESTClient) Unstar(owner, repo string) error {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/user/starred/%s/%s", c.baseURL, owner, repo), nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return fmt.Errorf("unstar failed: %s", resp.Status)
	}
	return nil
}

// --- Notifications ---

type Subject struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

type NotifRepo struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

type Notification struct {
	ID         string    `json:"id"`
	Unread     bool      `json:"unread"`
	Reason     string    `json:"reason"`
	UpdatedAt  string    `json:"updated_at"`
	Subject    Subject   `json:"subject"`
	Repository NotifRepo `json:"repository"`
}

func (c *RESTClient) ListNotifications() ([]Notification, error) {
	var notifs []Notification
	return notifs, c.get("/notifications?all=false&per_page=100", &notifs)
}

func (c *RESTClient) MarkRead(id string) error {
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/notifications/threads/%s", c.baseURL, id), nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *RESTClient) MarkAllRead() error {
	req, _ := http.NewRequest("PUT", c.baseURL+"/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
```

**Step 4: Run tests**

```bash
go test ./api/... -v -run TestListStars
go test ./api/... -v -run TestListNotifications
```

Expected: PASS

**Step 5: Commit**

```bash
git add api/rest.go api/rest_test.go
git commit -m "feat: GitHub REST client for stars and notifications"
```

---

### Task 5: GitHub GraphQL Client (Contributions)

**Files:**
- Create: `api/graphql.go`
- Create: `api/graphql_test.go`

**Step 1: Write failing test**

```go
// api/graphql_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchContributions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": map[string]any{
				"user": map[string]any{
					"login": "Null-Phnix",
					"contributionsCollection": map[string]any{
						"totalCommitContributions": 150,
						"contributionCalendar": map[string]any{
							"totalContributions": 200,
							"weeks": []map[string]any{
								{
									"contributionDays": []map[string]any{
										{"date": "2026-01-01", "contributionCount": 5, "weekday": 4},
									},
								},
							},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewGraphQLClient("test-token", srv.URL)
	data, err := client.FetchContributions(2026)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Login != "Null-Phnix" {
		t.Errorf("expected login 'Null-Phnix', got %q", data.Login)
	}
	if data.TotalContributions != 200 {
		t.Errorf("expected 200 total, got %d", data.TotalContributions)
	}
}
```

**Step 2: Run to verify failure**

```bash
go test ./api/... -v -run TestFetchContributions
```

**Step 3: Implement api/graphql.go**

```go
// api/graphql.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GraphQLClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewGraphQLClient(token, baseURL string) *GraphQLClient {
	if baseURL == "" {
		baseURL = "https://api.github.com/graphql"
	}
	return &GraphQLClient{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

type ContributionDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
	Weekday           int    `json:"weekday"`
}

type ContributionWeek struct {
	ContributionDays []ContributionDay `json:"contributionDays"`
}

type ContributionData struct {
	Login              string
	TotalContributions int
	Weeks              []ContributionWeek
}

func (c *GraphQLClient) FetchContributions(year int) (*ContributionData, error) {
	from := fmt.Sprintf("%d-01-01T00:00:00Z", year)
	to := fmt.Sprintf("%d-12-31T23:59:59Z", year)

	query := fmt.Sprintf(`{
		user: viewer {
			login
			contributionsCollection(from: "%s", to: "%s") {
				totalCommitContributions
				contributionCalendar {
					totalContributions
					weeks {
						contributionDays {
							date
							contributionCount
							weekday
						}
					}
				}
			}
		}
	}`, from, to)

	body, _ := json.Marshal(map[string]string{"query": query})
	req, err := http.NewRequest("POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			User struct {
				Login                   string `json:"login"`
				ContributionsCollection struct {
					ContributionCalendar struct {
						TotalContributions int                `json:"totalContributions"`
						Weeks              []ContributionWeek `json:"weeks"`
					} `json:"contributionCalendar"`
				} `json:"contributionsCollection"`
			} `json:"user"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	u := result.Data.User
	return &ContributionData{
		Login:              u.Login,
		TotalContributions: u.ContributionsCollection.ContributionCalendar.TotalContributions,
		Weeks:              u.ContributionsCollection.ContributionCalendar.Weeks,
	}, nil
}
```

**Step 4: Run tests**

```bash
go test ./api/... -v -run TestFetchContributions
```

Expected: PASS

**Step 5: Commit**

```bash
git add api/graphql.go api/graphql_test.go
git commit -m "feat: GitHub GraphQL client for contribution data"
```

---

### Task 6: Root App Model + Tab Switching

**Files:**
- Create: `ui/app.go`

**Step 1: Implement ui/app.go**

```go
// ui/app.go
package ui

import (
	"github.com/Null-Phnix/ghboard/api"
	"github.com/Null-Phnix/ghboard/config"
	"github.com/Null-Phnix/ghboard/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tab int

const (
	tabHeatmap tab = iota
	tabStars
	tabNotifications
)

var tabNames = []string{"  Heatmap  ", "  Stars  ", "  Notifications  "}

type App struct {
	cfg          *config.Config
	rest         *api.RESTClient
	gql          *api.GraphQLClient
	tags         *store.Store
	activeTab    tab
	width        int
	height       int
	heatmap      HeatmapModel
	stars        StarsModel
	notifications NotificationsModel
	showHelp     bool
}

func NewApp(cfg *config.Config) *App {
	rest := api.NewRESTClient(cfg.Token, "")
	gql := api.NewGraphQLClient(cfg.Token, "")
	tags := store.New(store.DefaultPath())

	return &App{
		cfg:           cfg,
		rest:          rest,
		gql:           gql,
		tags:          tags,
		activeTab:     tabHeatmap,
		heatmap:       NewHeatmapModel(gql),
		stars:         NewStarsModel(rest, tags),
		notifications: NewNotificationsModel(rest),
	}
}

func (a *App) Init() tea.Cmd {
	return a.heatmap.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		if a.showHelp {
			a.showHelp = false
			return a, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			a.tags.Save()
			return a, tea.Quit
		case "?":
			a.showHelp = true
			return a, nil
		case "1":
			a.activeTab = tabHeatmap
			return a, a.heatmap.Init()
		case "2":
			a.activeTab = tabStars
			return a, a.stars.Init()
		case "3":
			a.activeTab = tabNotifications
			return a, a.notifications.Init()
		case "tab":
			a.activeTab = (a.activeTab + 1) % 3
		}
	}

	var cmd tea.Cmd
	switch a.activeTab {
	case tabHeatmap:
		a.heatmap, cmd = a.heatmap.Update(msg)
	case tabStars:
		a.stars, cmd = a.stars.Update(msg)
	case tabNotifications:
		a.notifications, cmd = a.notifications.Update(msg)
	}
	return a, cmd
}

var (
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF7F")).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#00FF7F"))

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1)
)

func (a *App) View() string {
	if a.showHelp {
		return a.helpView()
	}

	// Tab bar
	tabs := ""
	for i, name := range tabNames {
		if tab(i) == a.activeTab {
			tabs += activeTabStyle.Render(name)
		} else {
			tabs += inactiveTabStyle.Render(name)
		}
	}

	tabBar := lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Width(a.width).
		Render(tabs)

	statusBar := statusBarStyle.Render("1/2/3: switch tabs  •  ?: help  •  q: quit")

	contentHeight := a.height - 4
	var content string
	switch a.activeTab {
	case tabHeatmap:
		content = a.heatmap.View(a.width, contentHeight)
	case tabStars:
		content = a.stars.View(a.width, contentHeight)
	case tabNotifications:
		content = a.notifications.View(a.width, contentHeight)
	}

	return tabBar + "\n" + content + "\n" + statusBar
}

func (a *App) helpView() string {
	help := `
  ghboard — GitHub Terminal Dashboard

  GLOBAL
    1 / 2 / 3   Switch tabs
    Tab         Cycle tabs
    ?           Toggle help
    q           Quit

  HEATMAP
    ← → ↑ ↓    Navigate days
    [ ]         Previous / next year

  STARS
    /           Fuzzy search
    t           Add/edit tags
    f           Filter by tag or language
    u           Unstar (confirm prompt)
    o           Open in browser
    ctrl+r      Refresh

  NOTIFICATIONS
    r           Mark as read
    R           Mark all as read
    d           Dismiss
    o           Open in browser
    f           Filter by type
`
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(help)
}
```

**Step 2: Verify it compiles (stubs for other ui files needed first)**

Create temporary stubs:

```go
// ui/heatmap.go (stub)
package ui
import (
    "github.com/Null-Phnix/ghboard/api"
    tea "github.com/charmbracelet/bubbletea"
)
type HeatmapModel struct { gql *api.GraphQLClient }
func NewHeatmapModel(gql *api.GraphQLClient) HeatmapModel { return HeatmapModel{gql: gql} }
func (m HeatmapModel) Init() tea.Cmd { return nil }
func (m HeatmapModel) Update(msg tea.Msg) (HeatmapModel, tea.Cmd) { return m, nil }
func (m HeatmapModel) View(w, h int) string { return "heatmap coming soon" }
```

```go
// ui/stars.go (stub)
package ui
import (
    "github.com/Null-Phnix/ghboard/api"
    "github.com/Null-Phnix/ghboard/store"
    tea "github.com/charmbracelet/bubbletea"
)
type StarsModel struct { rest *api.RESTClient; tags *store.Store }
func NewStarsModel(rest *api.RESTClient, tags *store.Store) StarsModel { return StarsModel{rest: rest, tags: tags} }
func (m StarsModel) Init() tea.Cmd { return nil }
func (m StarsModel) Update(msg tea.Msg) (StarsModel, tea.Cmd) { return m, nil }
func (m StarsModel) View(w, h int) string { return "stars coming soon" }
```

```go
// ui/notifications.go (stub)
package ui
import (
    "github.com/Null-Phnix/ghboard/api"
    tea "github.com/charmbracelet/bubbletea"
)
type NotificationsModel struct { rest *api.RESTClient }
func NewNotificationsModel(rest *api.RESTClient) NotificationsModel { return NotificationsModel{rest: rest} }
func (m NotificationsModel) Init() tea.Cmd { return nil }
func (m NotificationsModel) Update(msg tea.Msg) (NotificationsModel, tea.Cmd) { return m, nil }
func (m NotificationsModel) View(w, h int) string { return "notifications coming soon" }
```

```bash
go build ./...
```

Expected: compiles cleanly

**Step 3: Run it**

```bash
GITHUB_TOKEN=your_token go run .
```

Expected: TUI launches, tabs visible, `q` quits

**Step 4: Commit**

```bash
git add ui/
git commit -m "feat: root app model with tab switching and help overlay"
```

---

### Task 7: Heatmap Tab

**Files:**
- Modify: `ui/heatmap.go` (replace stub)

**Step 1: Implement full heatmap**

```go
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
	gql      *api.GraphQLClient
	year     int
	data     *api.ContributionData
	loading  bool
	err      error
	cursorX  int // week index
	cursorY  int // day index (0=Sun)
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

func (m HeatmapModel) View(w, h int) string {
	if m.loading {
		return lipgloss.NewStyle().Padding(2, 4).Render("⟳ Loading contributions...")
	}
	if m.err != nil {
		return lipgloss.NewStyle().Padding(2, 4).Foreground(lipgloss.Color("#FF5555")).
			Render(fmt.Sprintf("Error: %v\n\nctrl+r to retry", m.err))
	}
	if m.data == nil {
		return ""
	}

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#39d353")).
		Padding(1, 4).
		Render(fmt.Sprintf("%s — %d Contributions in %d", m.data.Login, m.data.TotalContributions, m.year))

	// Grid
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	grid := ""

	for day := 0; day < 7; day++ {
		row := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).
			Render(fmt.Sprintf("%s ", days[day]))

		for weekIdx, week := range m.data.Weeks {
			cell := " "
			color := heatColors[0]
			block := "░"

			for _, d := range week.ContributionDays {
				if d.Weekday == day {
					color = countToColor(d.ContributionCount)
					block = countToBlock(d.ContributionCount)
					break
				}
			}

			style := lipgloss.NewStyle().Foreground(color)
			if weekIdx == m.cursorX && day == m.cursorY {
				style = style.Background(lipgloss.Color("#FFFFFF")).Foreground(lipgloss.Color("#000000"))
				block = "█"
			}
			_ = cell
			row += style.Render(block + " ")
		}
		grid += row + "\n"
	}

	// Status bar for hovered day
	status := ""
	if m.cursorX < len(m.data.Weeks) {
		week := m.data.Weeks[m.cursorX]
		for _, d := range week.ContributionDays {
			if d.Weekday == m.cursorY {
				status = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#888888")).
					Padding(0, 4).
					Render(fmt.Sprintf("📅 %s — %d contributions", d.Date, d.ContributionCount))
				break
			}
		}
	}

	// Legend
	legend := lipgloss.NewStyle().Padding(0, 4).Render(
		"Less " +
			lipgloss.NewStyle().Foreground(heatColors[0]).Render("░") +
			lipgloss.NewStyle().Foreground(heatColors[1]).Render("▒") +
			lipgloss.NewStyle().Foreground(heatColors[2]).Render("▓") +
			lipgloss.NewStyle().Foreground(heatColors[3]).Render("█") +
			lipgloss.NewStyle().Foreground(heatColors[4]).Render("█") +
			" More  •  [ ] to change year",
	)

	gridBlock := lipgloss.NewStyle().Padding(0, 4).Render(grid)

	return header + "\n" + gridBlock + "\n" + status + "\n\n" + legend
}
```

**Step 2: Build and test manually**

```bash
GITHUB_TOKEN=your_token go run .
```

Expected: Tab 1 shows contribution grid, arrow keys move cursor, `[` `]` change year

**Step 3: Commit**

```bash
git add ui/heatmap.go
git commit -m "feat: contribution heatmap tab with year navigation"
```

---

### Task 8: Star Manager Tab

**Files:**
- Modify: `ui/stars.go` (replace stub)

**Step 1: Implement full star manager**

```go
// ui/stars.go
package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Null-Phnix/ghboard/api"
	"github.com/Null-Phnix/ghboard/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type starsLoadedMsg struct {
	repos []api.StarredRepo
	err   error
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
	confirming bool // unstar confirm
	tagEditing bool
	tagInput   string
}

func NewStarsModel(rest *api.RESTClient, tags *store.Store) StarsModel {
	return StarsModel{rest: rest, tags: tags}
}

func (m StarsModel) Init() tea.Cmd {
	if m.repos != nil {
		return nil // already loaded
	}
	m.loading = true
	return func() tea.Msg {
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
	}
}

func (m *StarsModel) applyFilter() {
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
		if m.search != "" && !strings.Contains(strings.ToLower(r.FullName), strings.ToLower(m.search)) {
			continue
		}
		m.filtered = append(m.filtered, r)
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
	}
}

func (m StarsModel) Update(msg tea.Msg) (StarsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case starsLoadedMsg:
		m.loading = false
		m.repos = msg.repos
		m.err = msg.err
		m.applyFilter()
		return m, nil

	case tea.KeyMsg:
		if m.tagEditing {
			switch msg.String() {
			case "enter":
				if m.cursor < len(m.filtered) {
					repo := m.filtered[m.cursor]
					tags := strings.Split(m.tagInput, ",")
					var cleaned []string
					for _, t := range tags {
						t = strings.TrimSpace(t)
						if t != "" {
							cleaned = append(cleaned, t)
						}
					}
					m.tags.Set(repo.FullName, cleaned)
					m.tags.Save()
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
				m.tagInput += msg.String()
			}
			return m, nil
		}

		if m.searching {
			switch msg.String() {
			case "enter", "esc":
				m.searching = false
				m.applyFilter()
			case "backspace":
				if len(m.search) > 0 {
					m.search = m.search[:len(m.search)-1]
					m.applyFilter()
				}
			default:
				m.search += msg.String()
				m.applyFilter()
			}
			return m, nil
		}

		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				m.confirming = false
				if m.cursor < len(m.filtered) {
					repo := m.filtered[m.cursor]
					parts := strings.SplitN(repo.FullName, "/", 2)
					if len(parts) == 2 {
						go m.rest.Unstar(parts[0], parts[1])
					}
					m.repos = remove(m.repos, repo.FullName)
					m.applyFilter()
				}
			default:
				m.confirming = false
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "/":
			m.searching = true
			m.search = ""
		case "t":
			if m.cursor < len(m.filtered) {
				existing := m.tags.Get(m.filtered[m.cursor].FullName)
				m.tagInput = strings.Join(existing, ", ")
				m.tagEditing = true
			}
		case "u":
			m.confirming = true
		case "o":
			if m.cursor < len(m.filtered) {
				exec.Command("open", m.filtered[m.cursor].HTMLURL).Start()
			}
		case "f":
			// cycle filter off → tags → languages → off
			m.filterTag = ""
			m.filterLang = ""
			m.applyFilter()
		case "ctrl+r":
			m.repos = nil
			return m, m.Init()
		}
	}
	return m, nil
}

var tagColors = []lipgloss.Color{
	"#FF79C6", "#8BE9FD", "#50FA7B", "#FFB86C", "#BD93F9", "#FF5555", "#F1FA8C",
}

func tagStyle(tag string) string {
	h := 0
	for _, c := range tag {
		h += int(c)
	}
	color := tagColors[h%len(tagColors)]
	return lipgloss.NewStyle().
		Background(color).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1).
		Render(tag)
}

func (m StarsModel) View(w, h int) string {
	if m.loading {
		return lipgloss.NewStyle().Padding(2, 4).Render("⟳ Loading stars...")
	}
	if m.err != nil {
		return lipgloss.NewStyle().Padding(2, 4).Foreground(lipgloss.Color("#FF5555")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}

	header := lipgloss.NewStyle().Bold(true).Padding(1, 4).
		Render(fmt.Sprintf("⭐ %d stars", len(m.repos)))

	if m.search != "" {
		header += lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render(fmt.Sprintf("  /  %s", m.search))
	}

	if m.tagEditing && m.cursor < len(m.filtered) {
		return header + "\n\n" +
			lipgloss.NewStyle().Padding(0, 4).Render(
				fmt.Sprintf("Tags for %s\n\nEnter comma-separated tags: %s█\n\nEnter to save, Esc to cancel",
					m.filtered[m.cursor].FullName, m.tagInput))
	}

	if m.confirming && m.cursor < len(m.filtered) {
		return header + "\n\n" +
			lipgloss.NewStyle().Padding(0, 4).Foreground(lipgloss.Color("#FF5555")).
				Render(fmt.Sprintf("Unstar %s? (y/n)", m.filtered[m.cursor].FullName))
	}

	// List
	listHeight := h - 6
	start := 0
	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}

	rows := ""
	for i := start; i < len(m.filtered) && i < start+listHeight; i++ {
		r := m.filtered[i]
		repoTags := m.tags.Get(r.FullName)

		line := ""
		name := lipgloss.NewStyle().Bold(i == m.cursor).Render(r.FullName)
		lang := ""
		if r.Language != "" {
			lang = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(" [" + r.Language + "]")
		}
		tagStr := ""
		for _, t := range repoTags {
			tagStr += " " + tagStyle(t)
		}

		prefix := "  "
		if i == m.cursor {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("#39d353")).Render("▶ ")
		}

		desc := ""
		if r.Description != "" {
			d := r.Description
			if len(d) > 60 {
				d = d[:57] + "..."
			}
			desc = "\n    " + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(d)
		}

		line = prefix + name + lang + tagStr + desc + "\n"
		rows += line
	}

	statusBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 4).
		Render("/: search  •  t: tags  •  u: unstar  •  o: open  •  ctrl+r: refresh")

	return header + "\n" + lipgloss.NewStyle().Padding(0, 2).Render(rows) + "\n" + statusBar
}

func remove(repos []api.StarredRepo, fullName string) []api.StarredRepo {
	var out []api.StarredRepo
	for _, r := range repos {
		if r.FullName != fullName {
			out = append(out, r)
		}
	}
	return out
}
```

**Step 2: Build and test manually**

```bash
GITHUB_TOKEN=your_token go run .
```

Press `2` → star manager loads, `/` to search, `t` to tag, `o` to open in browser

**Step 3: Commit**

```bash
git add ui/stars.go
git commit -m "feat: star manager with tags, search, filter, unstar"
```

---

### Task 9: Notification Center Tab

**Files:**
- Modify: `ui/notifications.go` (replace stub)

**Step 1: Implement full notification center**

```go
// ui/notifications.go
package ui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Null-Phnix/ghboard/api"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type notifsLoadedMsg struct {
	notifs []api.Notification
	err    error
}

type refreshTickMsg struct{}

type NotificationsModel struct {
	rest    *api.RESTClient
	notifs  []api.Notification
	filter  string // "", "PullRequest", "Issue", "CheckSuite", "Release"
	loading bool
	err     error
	cursor  int
}

func NewNotificationsModel(rest *api.RESTClient) NotificationsModel {
	return NotificationsModel{rest: rest}
}

func (m NotificationsModel) Init() tea.Cmd {
	return tea.Batch(m.fetchCmd(), m.tickCmd())
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

func (m NotificationsModel) filtered() []api.Notification {
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
	case notifsLoadedMsg:
		m.loading = false
		m.notifs = msg.notifs
		m.err = msg.err
		if m.cursor >= len(m.notifs) {
			m.cursor = 0
		}
		return m, m.tickCmd()

	case refreshTickMsg:
		return m, m.fetchCmd()

	case tea.KeyMsg:
		visible := m.filtered()
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(visible)-1 {
				m.cursor++
			}
		case "r":
			if m.cursor < len(visible) {
				id := visible[m.cursor].ID
				go m.rest.MarkRead(id)
				m.notifs = markRead(m.notifs, id)
			}
		case "R":
			go m.rest.MarkAllRead()
			for i := range m.notifs {
				m.notifs[i].Unread = false
			}
		case "d":
			if m.cursor < len(visible) {
				id := visible[m.cursor].ID
				go m.rest.MarkRead(id)
				m.notifs = remove2(m.notifs, id)
				if m.cursor >= len(m.filtered()) {
					m.cursor = max(0, len(m.filtered())-1)
				}
			}
		case "o":
			if m.cursor < len(visible) {
				url := "https://github.com/notifications"
				if visible[m.cursor].Repository.HTMLURL != "" {
					url = visible[m.cursor].Repository.HTMLURL
				}
				exec.Command("open", url).Start()
			}
		case "f":
			filters := []string{"", "PullRequest", "Issue", "CheckSuite", "Release"}
			cur := 0
			for i, f := range filters {
				if f == m.filter {
					cur = i
					break
				}
			}
			m.filter = filters[(cur+1)%len(filters)]
			m.cursor = 0
		case "ctrl+r":
			m.loading = true
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

func typeBadge(t string) string {
	short := map[string]string{
		"PullRequest": "PR",
		"Issue":       "ISS",
		"CheckSuite":  "CI",
		"Release":     "REL",
		"Discussion":  "DSC",
	}
	label := short[t]
	if label == "" {
		label = t[:min(3, len(t))]
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

func (m NotificationsModel) View(w, h int) string {
	visible := m.filtered()
	unread := 0
	for _, n := range m.notifs {
		if n.Unread {
			unread++
		}
	}

	filterStr := ""
	if m.filter != "" {
		filterStr = " [" + m.filter + "]"
	}
	header := lipgloss.NewStyle().Bold(true).Padding(1, 4).
		Render(fmt.Sprintf("🔔 Notifications (%d unread)%s", unread, filterStr))

	if m.loading {
		return header + "\n\n" + lipgloss.NewStyle().Padding(0, 4).Render("⟳ Loading...")
	}
	if m.err != nil {
		return header + "\n\n" + lipgloss.NewStyle().Padding(0, 4).Foreground(lipgloss.Color("#FF5555")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}
	if len(visible) == 0 {
		return header + "\n\n" + lipgloss.NewStyle().Padding(0, 4).Render("No notifications")
	}

	listHeight := h - 6
	start := 0
	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}

	// Group by repo
	rows := ""
	lastRepo := ""
	for i := start; i < len(visible) && i < start+listHeight; i++ {
		n := visible[i]

		if n.Repository.FullName != lastRepo {
			lastRepo = n.Repository.FullName
			rows += "\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Bold(true).
				Padding(0, 4).
				Render(lastRepo) + "\n"
		}

		unreadDot := " "
		if n.Unread {
			unreadDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#39d353")).Render("●")
		}

		prefix := "   "
		if i == m.cursor {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("#39d353")).Render("  ▶")
		}

		title := n.Subject.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		row := fmt.Sprintf("%s %s %s %s\n",
			prefix,
			unreadDot,
			typeBadge(n.Subject.Type),
			title,
		)
		rows += row
	}

	statusBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 4).
		Render("r: read  •  R: all read  •  d: dismiss  •  o: open  •  f: filter  •  ctrl+r: refresh")

	return header + rows + "\n" + statusBar
}

func markRead(notifs []api.Notification, id string) []api.Notification {
	for i := range notifs {
		if notifs[i].ID == id {
			notifs[i].Unread = false
		}
	}
	return notifs
}

func remove2(notifs []api.Notification, id string) []api.Notification {
	var out []api.Notification
	for _, n := range notifs {
		if n.ID != id {
			out = append(out, n)
		}
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func strings_contains_fold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
```

**Step 2: Build and test manually**

```bash
GITHUB_TOKEN=your_token go run .
```

Press `3` → notifications load, `r` marks read, `f` cycles filters, `o` opens in browser

**Step 3: Commit**

```bash
git add ui/notifications.go
git commit -m "feat: notification center with mark read, dismiss, filter, auto-refresh"
```

---

### Task 10: First-Run Token Prompt

**Files:**
- Modify: `main.go`

**Step 1: Update main.go to handle missing token**

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Null-Phnix/ghboard/config"
	"github.com/Null-Phnix/ghboard/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	if cfg.Token == "" {
		fmt.Println("ghboard — GitHub Terminal Dashboard")
		fmt.Println()
		fmt.Println("No GitHub token found.")
		fmt.Println("Create one at: https://github.com/settings/tokens")
		fmt.Println("Required scopes: repo, notifications, read:user")
		fmt.Println()
		fmt.Print("Enter your token: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			fmt.Fprintln(os.Stderr, "No token provided. Exiting.")
			os.Exit(1)
		}

		cfg.Token = token
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		} else {
			fmt.Println("Token saved to ~/.config/ghboard/config.json")
		}
		fmt.Println()
	}

	p := tea.NewProgram(ui.NewApp(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Build final binary**

```bash
go build -o ghboard .
./ghboard
```

Expected: If no token → prompts for one. If token exists → launches TUI directly.

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: first-run token prompt with config persistence"
```

---

### Task 11: GitHub Repo + Polish

**Step 1: Create GitHub repo and push**

```bash
gh repo create Null-Phnix/ghboard --public --description "GitHub terminal dashboard: heatmap, stars, notifications" --push --source .
```

**Step 2: Add README**

Create `README.md` with install instructions, feature list, screenshots placeholder, keybindings table.

**Step 3: Tag v0.1.0**

```bash
git tag v0.1.0
git push origin v0.1.0
```

**Step 4: Test `go install`**

```bash
go install github.com/Null-Phnix/ghboard@latest
ghboard
```

Expected: installs and runs cleanly

**Step 5: Final commit**

```bash
git add README.md
git commit -m "docs: add README with install and usage"
git push
```

---

## Summary

| Task | What it builds |
|------|---------------|
| 1 | Project scaffold, go mod, deps |
| 2 | Config (env + file, 600 perms) |
| 3 | Tag store (JSON persistence) |
| 4 | GitHub REST client (stars + notifications) |
| 5 | GitHub GraphQL client (contributions) |
| 6 | Root app model + tab switching + help |
| 7 | Heatmap tab (grid, cursor, year toggle) |
| 8 | Star manager (search, tags, unstar, open) |
| 9 | Notification center (mark read, dismiss, filter, auto-refresh) |
| 10 | First-run token prompt |
| 11 | GitHub repo + README + v0.1.0 |
