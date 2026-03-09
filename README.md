# ghboard

**Your GitHub life, in the terminal.**

`ghboard` is a terminal dashboard for GitHub developers who live in the CLI. Three tabs, zero browser required.

```
  Heatmap   Stars   Notifications
──────────────────────────────────────────────────────────────────────
Null-Phnix — 1,247 contributions in 2026   [ prev year ]

    Jan       Feb       Mar       Apr       May
Sun ░ ░ ░ ░ ▒ ▒ ▒ ▓ ▓ ▓ █ █ █ ▓ ▓ ▒ ▒ ░ ░ ░ ░ ░ ░ ░
Mon ░ ░ ░ ▒ ▒ ▓ ▓ █ █ ▓ ▓ ▓ ▒ ▒ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░
Tue ▒ ▒ ▓ █ █ █ ▓ ▓ ▒ ▒ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░
Wed ░ ░ ▒ ▒ ▓ █ ▓ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░
...

📅 2026-03-09 — ████ 7 contributions
Less ░ ▒ ▓ █ █ More  •  [ / ] year  •  arrows / hjkl navigate
──────────────────────────────────────────────────────────────────────
1/2/3: switch tabs  •  Tab: cycle  •  ?: help  •  q: quit
```

## Features

| Tab | What it does |
|-----|-------------|
| **Heatmap** | GitHub-style contribution grid with year navigation, day-level detail, live cursor |
| **Stars** | Browse all starred repos, fuzzy search, tag repos with custom labels, unstar, open in browser |
| **Notifications** | Inbox-zero workflow — mark read, dismiss, filter by type (PR/Issue/CI/Release), auto-refreshes every 60s |

## Install

**Homebrew** *(coming soon)*
```bash
brew install Null-Phnix/tap/ghboard
```

**go install** *(Go 1.21+)*
```bash
go install github.com/Null-Phnix/ghboard@latest
```

**Download binary** from [Releases](https://github.com/Null-Phnix/ghboard/releases)

## Setup

```bash
ghboard
```

On first run, `ghboard` asks for a GitHub personal access token. Create one at [github.com/settings/tokens](https://github.com/settings/tokens) with these scopes:

- `repo` — read starred repos
- `notifications` — read & manage notifications
- `read:user` — fetch contribution data

The token is stored at `~/.config/ghboard/config.json` (mode `0600`).

You can also set `GITHUB_TOKEN` env var to skip the prompt.

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `1` / `2` / `3` | Switch to Heatmap / Stars / Notifications tab |
| `Tab` | Cycle to next tab |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

### Heatmap

| Key | Action |
|-----|--------|
| `←→↑↓` / `hjkl` | Move cursor |
| `[` / `]` | Previous / next year |
| `Ctrl+R` | Refresh data |

### Stars

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Navigate list |
| `g` / `G` | Jump to top / bottom |
| `/` | Fuzzy search (name, description, language) |
| `Esc` | Clear search |
| `t` | Add / edit comma-separated tags |
| `f` | Clear active filter |
| `u` | Unstar selected repo (confirm with `y`) |
| `o` | Open in browser |
| `Ctrl+R` | Refresh |

### Notifications

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Navigate list |
| `g` / `G` | Jump to top / bottom |
| `r` | Mark selected as read |
| `R` | Mark all as read |
| `d` | Dismiss (mark read + remove from list) |
| `o` | Open repo in browser |
| `f` | Cycle filter: All → PR → Issue → CI → Release → Discussion |
| `Ctrl+R` | Refresh now |

## Configuration

`~/.config/ghboard/config.json`

```json
{
  "token": "ghp_..."
}
```

Tags are stored separately at `~/.config/ghboard/tags.json` and persist across sessions.

## Tech Stack

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — TUI framework (Elm architecture)
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — styling & layout
- **[Bubbles](https://github.com/charmbracelet/bubbles)** — UI components
- **GitHub REST API** — stars, notifications
- **GitHub GraphQL API** — contribution data

## Contributing

PRs welcome. Run tests with:

```bash
go test ./...
```

## License

MIT
