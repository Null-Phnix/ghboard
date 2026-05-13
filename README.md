# ghboard 🜏

[![CI](https://github.com/Null-Phnix/ghboard/actions/workflows/ci.yml/badge.svg)](https://github.com/Null-Phnix/ghboard/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Null-Phnix/ghboard)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/Null-Phnix/ghboard)](https://github.com/Null-Phnix/ghboard/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Stay in your terminal. Browse stars, manage notifications, track contributions — without touching a browser.**

> *Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) · GitHub REST + GraphQL APIs · single static binary*

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

---

## Table of Contents

- [Why I Built This](#why-i-built-this)
- [What It Does](#what-it-does)
- [Current Pain Points](#current-pain-points)
- [End Goals](#end-goals--where-this-is-headed)
- [Install](#install)
- [Setup](#setup)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Configuration](#configuration)
- [Recording a Demo](#recording-a-demo)
- [Tech Stack](#tech-stack)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Why I Built This

I live in the terminal. My workflow is: terminal → editor → terminal → git → terminal → Claude Code → terminal. Switching to a browser to check GitHub notifications, remember which repos I starred, or verify my contribution streak breaks that flow completely.

`gh` exists and it's good for git operations. But `gh` doesn't show me a contribution heatmap. It doesn't let me fuzzy-search my 400 starred repos. It doesn't show me notifications grouped by repo with one-keystroke dismiss.

ghboard puts all three in one keystroke away — no browser tab, no context switch, no mouse. Just `ghboard` and I'm looking at everything.

---

## What It Does

| Tab | What it does |
|-----|-------------|
| 🗓 **Heatmap** | Full-year GitHub contribution grid · cursor navigation · `[` `]` year toggle · mini bar chart per day |
| ⭐ **Stars** | Browse all starred repos · fuzzy search · custom tags · language color dots · unstar · open in browser |
| 🔔 **Notifications** | Mark read / dismiss · filter by type · grouped by repo · relative timestamps · auto-refresh every 60 s |

---

## Current Pain Points

These are the battles I'm actively fighting:

1. **GitHub GraphQL API is a maze** — The contribution heatmap data requires a specific GraphQL query with nested nodes and edges. GitHub changes the schema occasionally and the heatmap breaks. I had to reverse-engineer the query from GitHub's own frontend requests.

2. **Notification pagination is unbounded** — GitHub returns ALL unread notifications in one request. If you have 500 unread (yes, it happens), the JSON response is 2MB+. ghboard handles it but it's slow. I need pagination but GitHub's notification API doesn't support it well.

3. **Star tags are local-only** — Tags you assign in ghboard live in `~/.config/ghboard/tags.json`. If you star a repo on GitHub's website, ghboard sees it but has no tags. The sync is one-way. I want to store tags in GitHub's own "lists" feature but there's no API for that.

4. **No caching** — Every launch hits the GitHub API. If you're on a train with spotty wifi, ghboard hangs. I need offline caching but haven't built it yet.

5. **GoReleaser is overkill for one binary** — The release pipeline uses GoReleaser to build for 6 platforms. It works but adds 5 minutes to every release. For a 2,000-line Go program, `go build` cross-compilation would be faster.

---

## End Goals — Where This Is Headed

### Short Term (now → 3 months)
- **Offline caching** — cache heatmap, stars, and notifications locally, refresh on demand
- **Better notification pagination** — chunk large notification sets, show progress
- **Homebrew tap** — the config exists but the tap isn't live yet

### Medium Term (3–6 months)
- **GitLab support** — most requested feature, different API but similar data model
- **Integration with agent ecosystem** — Blackreach could use ghboard's starred repo list as research seeds; Huginn could crawl starred repos for updates
- **Activity feed** — show recent commits, PRs, and issues from watched repos

### Long Term (6–12 months)
- **Universal code host dashboard** — GitHub + GitLab + Gitea + SourceHut, one TUI
- **Repo health scoring** — analyze starred repos for activity, maintenance, security issues
- **Contribution goals** — "You need 3 more commits today to maintain your streak" — gamification but actually useful

---

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

---

## Setup

```bash
ghboard
```

On first run you'll be prompted for a GitHub personal access token.
Create one at → **[github.com/settings/tokens](https://github.com/settings/tokens/new?scopes=repo,notifications,read:user)**

Required scopes: `repo` · `notifications` · `read:user`

The token is saved to `~/.config/ghboard/config.json` (`0600`).
You can also `export GITHUB_TOKEN=ghp_...` to skip the prompt.

---

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

---

## Configuration

`~/.config/ghboard/config.json`
```json
{
  "token": "ghp_..."
}
```

Tags are stored at `~/.config/ghboard/tags.json` and persist across sessions.

---

## Recording a Demo

A [VHS](https://github.com/charmbracelet/vhs) tape file is included:

```bash
brew install vhs
GITHUB_TOKEN=ghp_... vhs demo.tape
```

This produces `demo.gif` — a scriptable, reproducible terminal recording.

---

## Tech Stack

| | |
|--|--|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework (Elm architecture) |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Styling & layout |
| [Bubbles](https://github.com/charmbracelet/bubbles) | Spinner components |
| GitHub REST API | Stars & notifications |
| GitHub GraphQL API | Contribution heatmap data |

---

## Roadmap

- [ ] Animated GIF demo
- [x] Homebrew tap (configured, coming soon)
- [x] Pre-built binaries (GoReleaser)
- [ ] Sort stars by: recently starred, most ⭐, language
- [ ] Tag-based filtering in the Stars tab
- [ ] GitLab support *(most requested — [upvote here](https://github.com/Null-Phnix/ghboard/issues))*
- [ ] GitHub Enterprise support
- [ ] Configurable refresh interval

---

## Contributing

```bash
git clone https://github.com/Null-Phnix/ghboard
cd ghboard
go test ./...      # run tests
go build ./...     # verify build
```

PRs and issues welcome. If you want GitLab or another provider, open an issue to show demand.

---

## License

[MIT](LICENSE) © Null-Phnix
