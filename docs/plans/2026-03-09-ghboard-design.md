# ghboard — Design Document
_2026-03-09_

## Overview

A feature-full terminal UI for GitHub built in Go + Bubble Tea. Three tabs: Contribution Heatmap, Star Manager, Notification Center. Single binary, no external dependencies.

---

## Architecture

**Name:** `ghboard`

**Stack:**
- Go + Bubble Tea (TUI framework)
- Lip Gloss (styling — colors, borders, layout)
- GitHub GraphQL API — contribution heatmap data
- GitHub REST API — stars + notifications
- `~/.config/ghboard/config.json` — token + preferences
- `~/.config/ghboard/tags.json` — local star tags/categories

**Project structure:**
```
ghboard/
├── main.go
├── ui/
│   ├── app.go              # root model, tab switching
│   ├── heatmap.go          # contribution heatmap tab
│   ├── stars.go            # star manager tab
│   └── notifications.go    # notification center tab
├── api/
│   ├── graphql.go          # contribution data
│   └── rest.go             # stars + notifications
├── store/
│   └── tags.go             # local tag persistence
└── config/
    └── config.go           # token loading
```

**Navigation:**
- `1` `2` `3` or `Tab` to switch tabs
- `q` to quit
- `?` for help overlay

---

## Features

### Tab 1: Contribution Heatmap
- Full year grid (52 weeks × 7 days) rendered with `░▒▓█` block characters, green color scale
- Stats: current streak, longest streak, total contributions this year
- Arrow key navigation — hover a day to see exact date + count in status bar
- Toggle years with `[` `]`
- Username resolved automatically from token

### Tab 2: Star Manager
- Scrollable list of all starred repos (name, description, language, star count, last updated)
- Local tags per repo shown as colored badges (e.g. `tools`, `ai`, `inspo`, `reference`)
- Tag colors auto-assigned, consistent across sessions

**Keybindings:**
- `t` — add/edit tags
- `u` — unstar (with confirm prompt)
- `o` — open in browser
- `f` — filter by tag or language
- `/` — fuzzy search by name

### Tab 3: Notification Center
- All notifications grouped by repo
- Type badges: `PR` `ISSUE` `CI` `RELEASE` `MENTION` with distinct colors
- Unread count in tab header: `Notifications (12)`
- Auto-refresh every 60 seconds

**Keybindings:**
- `r` — mark as read
- `R` — mark all as read
- `o` — open in browser
- `d` — dismiss/done
- `f` — filter by type

---

## Data Flow

**Auth:**
- First run with no config → inline token prompt in TUI
- Token stored at `~/.config/ghboard/config.json` with `600` permissions
- Invalid token → clear error screen with re-auth prompt

**Fetching:**
- Tabs load lazily — data fetched only on first visit
- Loading spinner while fetching
- In-memory cache for session duration
- Notifications auto-refresh every 60s via Go ticker
- Stars: manual refresh with `ctrl+r`

**Error handling:**
- Rate limit → show reset time in status bar, disable affected tab
- Network error → inline error message, retry with `ctrl+r`
- API errors → dismissible banners, never crash

---

## Installation

```bash
go install github.com/Null-Phnix/ghboard@latest
```

Single binary. Works anywhere Go is installed.
