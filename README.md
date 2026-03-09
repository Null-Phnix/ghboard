# ghboard

[![CI](https://github.com/Null-Phnix/ghboard/actions/workflows/ci.yml/badge.svg)](https://github.com/Null-Phnix/ghboard/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Null-Phnix/ghboard)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/Null-Phnix/ghboard)](https://github.com/Null-Phnix/ghboard/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Stay in your terminal. Browse stars, manage notifications, track contributions — without touching a browser.**

> *Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) · GitHub REST + GraphQL APIs · single static binary*

---

<!-- Record with: vhs demo.tape  (https://github.com/charmbracelet/vhs) -->
<!-- ![ghboard demo](demo.gif) -->

```
╔══════════════════════════════════════════════════════════════════════╗
║  Heatmap   Stars   Notifications                                     ║
╠══════════════════════════════════════════════════════════════════════╣
║  Null-Phnix — 1,247 contributions in 2026   [ prev year ]           ║
║                                                                      ║
║      Jan       Feb       Mar       Apr       May       Jun           ║
║  Sun ░ ░ ░ ▒ ▒ ▓ █ █ ▓ ▓ ▒ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░            ║
║  Mon ░ ░ ▒ ▒ ▓ █ █ ▓ ▒ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░            ║
║  Tue ▒ ▓ █ █ ▓ ▒ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░            ║
║  Wed ░ ▒ ▓ █ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░            ║
║  Thu ░ ░ ▒ ▓ █ ▓ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░            ║
║  Fri ░ ░ ░ ▒ ▒ ▓ █ ▓ ▒ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░            ║
║  Sat ░ ░ ░ ░ ▒ ▒ ▓ █ █ ▓ ▒ ▒ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░ ░            ║
║                                                                      ║
║  📅 2026-03-09 — ████████ 12 contributions                          ║
║  Less ░ ▒ ▓ █ █ More  •  [ / ] year  •  arrows / hjkl navigate      ║
╚══════════════════════════════════════════════════════════════════════╝
```

## Why ghboard?

If you live in the terminal, switching to a browser to check GitHub notifications, remember which repos you starred, or check your contribution streak breaks your flow. `ghboard` puts all three in one keystroke away — no browser tab, no context switch.

| Tab | What it does |
|-----|-------------|
| 🗓 **Heatmap** | Full-year GitHub contribution grid · cursor navigation · `[` `]` year toggle · mini bar chart per day |
| ⭐ **Stars** | Browse all starred repos · fuzzy search · custom tags · language color dots · unstar · open in browser |
| 🔔 **Notifications** | Mark read / dismiss · filter by type · grouped by repo · relative timestamps · auto-refresh every 60 s |

## Install

**go install** *(requires Go 1.21+)*
```bash
go install github.com/Null-Phnix/ghboard@latest
```

**Download a pre-built binary**

Grab the latest release for your platform from the [releases page](https://github.com/Null-Phnix/ghboard/releases/latest).

Supported platforms:
- macOS — Intel (x86_64) and Apple Silicon (ARM64)
- Linux — x86_64 and ARM64
- Windows — x86_64

Extract the archive and place the `ghboard` binary somewhere on your `$PATH`.

**Homebrew** *(coming soon)*
```bash
brew install Null-Phnix/tap/ghboard
```
> The tap is not live yet — watch this repo for updates.

**From source**
```bash
git clone https://github.com/Null-Phnix/ghboard
cd ghboard
go build -o ghboard .
```

## Setup

```bash
ghboard
```

On first run you'll be prompted for a GitHub personal access token.
Create one at → **[github.com/settings/tokens](https://github.com/settings/tokens/new?scopes=repo,notifications,read:user)**

Required scopes: `repo` · `notifications` · `read:user`

The token is saved to `~/.config/ghboard/config.json` (`0600`).
You can also `export GITHUB_TOKEN=ghp_...` to skip the prompt.

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `1` / `2` / `3` | Switch tabs |
| `Tab` | Cycle to next tab |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

### 🗓 Heatmap

| Key | Action |
|-----|--------|
| `←→↑↓` / `hjkl` | Move cursor |
| `[` / `]` | Previous / next year |
| `Ctrl+R` | Refresh |

### ⭐ Stars

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Navigate |
| `g` / `G` | Top / bottom |
| `/` | Fuzzy search |
| `Esc` | Clear search |
| `t` | Edit tags (comma-separated) |
| `f` | Clear filter |
| `u` | Unstar (confirm `y`) |
| `o` | Open in browser |
| `Ctrl+R` | Refresh |

### 🔔 Notifications

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Navigate |
| `g` / `G` | Top / bottom |
| `r` | Mark as read |
| `R` | Mark ALL read |
| `d` | Dismiss |
| `o` | Open repo in browser |
| `f` | Cycle type filter (All → PR → Issue → CI → Release → Discussion) |
| `Ctrl+R` | Refresh now |

## Configuration

`~/.config/ghboard/config.json`
```json
{
  "token": "ghp_..."
}
```

Tags are stored at `~/.config/ghboard/tags.json` and persist across sessions.

## Recording a Demo

A [VHS](https://github.com/charmbracelet/vhs) tape file is included:

```bash
brew install vhs
GITHUB_TOKEN=ghp_... vhs demo.tape
```

This produces `demo.gif` — a scriptable, reproducible terminal recording.

## Tech Stack

| | |
|--|--|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework (Elm architecture) |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Styling & layout |
| [Bubbles](https://github.com/charmbracelet/bubbles) | Spinner components |
| GitHub REST API | Stars & notifications |
| GitHub GraphQL API | Contribution heatmap data |

## Roadmap

- [ ] Animated GIF demo
- [x] Homebrew tap (configured, coming soon)
- [x] Pre-built binaries (GoReleaser)
- [ ] Sort stars by: recently starred, most ⭐, language
- [ ] Tag-based filtering in the Stars tab
- [ ] GitLab support *(most requested — [upvote here](https://github.com/Null-Phnix/ghboard/issues))*
- [ ] GitHub Enterprise support
- [ ] Configurable refresh interval

## Contributing

```bash
git clone https://github.com/Null-Phnix/ghboard
cd ghboard
go test ./...      # run tests
go build ./...     # verify build
```

PRs and issues welcome. If you want GitLab or another provider, open an issue to show demand.

## License

[MIT](LICENSE) © Null-Phnix
