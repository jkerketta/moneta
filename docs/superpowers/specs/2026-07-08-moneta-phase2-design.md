# Moneta Phase 2 — Design Spec

## Overview

Convert Moneta from a CLI tool into a full-screen Textual TUI application. Type `moneta` to launch an interactive menu-driven terminal dashboard. All existing CLI subcommands (`moneta scan`, `moneta check`, etc.) remain functional for scripting.

## Tech Stack (additions)

- **Textual** — TUI framework from Textualize, same author as Rich. Keyboard navigation, layouts, widgets, CSS theming, reactive data binding.
- Rich (already used), Finnhub, pyyaml, numpy, pytest.

## Architecture

```
moneta/
├── __main__.py        → MODIFY: entry point detects TUI/CLI mode
├── app.py             → NEW: Textual App subclass, the TUI entry point
├── screens/
│   ├── home.py        → NEW: Home menu screen (LazyVim-style)
│   ├── portfolio.py   → NEW: Add/list/remove/donut sub-menu
│   ├── scan.py        → NEW: Scan progress + results screen
│   ├── report.py      → NEW: Sentiment report screen
│   ├── dashboard.py   → NEW: Donut + sentiment side-by-side
│   └── scheduler.py   → NEW: Watch install/status screen
├── widgets/
│   ├── donut.py       → NEW: Custom Textual widget — allocation donut
│   ├── title.py       → NEW: Static ASCII art title with gradient
│   └── menu.py        → NEW: Reusable keyboard-nav menu widget
├── portfolio.py       → MODIFY: already exists
├── news.py            → UNCHANGED
├── report.py          → UNCHANGED
├── config.py          → UNCHANGED
├── display.py         → UNCHANGED (still used by CLI path)
├── cli.py             → MODIFY: portfolio add → interactive when no args
└── scheduler.py       → NEW: launchd plist management
```

## Color Theme

- **Primary accent:** `#FF6F61` (coral red)
- **Background:** Terminal default (dark)
- **Text:** White/bright
- **Title:** Coral gradient
- **Selection arrows:** Coral `▸`
- **Alert flags:** Coral bold
- **Composite bars:** Coral for positive (>0.5), gold for neutral, red for negative

## Home Screen (`moneta` — no args)

```
  ╭──────────────────────────────────────────────────────────────╮
  │                                                              │
  │         ███╗   ███╗  ██████╗  ███╗   ██╗ ███████╗            │
  │         ████╗ ████║ ██╔═══██╗ ████╗  ██║ ██╔════╝            │
  │         ██╔████╔██║ ██║   ██║ ██╔██╗ ██║ █████╗              │
  │         ██║╚██╔╝██║ ██║   ██║ ██║╚██╗██║ ██╔══╝              │
  │         ██║ ╚═╝ ██║ ╚██████╔╝ ██║ ╚████║ ███████╗            │
  │         ╚═╝     ╚═╝  ╚═════╝  ╚═╝  ╚═══╝ ╚══════╝            │
  │                                                              │
  │              Portfolio Sentiment Engine                       │
  │                                                              │
  │     ▸  [P]ortfolio        Manage your holdings                │
  │        [S]can Sentiment   Pull fresh Finnhub data             │
  │        [R]eport           View latest sentiment               │
  │        [D]ashboard        Portfolio + sentiment at a glance   │
  │        [C]hart            Allocation donut                    │
  │        [W]atch            Schedule daily scans                │
  │                                                              │
  │      Portfolio: 2 holdings  │  Last scan: Jul 08, 8:34 PM     │
  │                                                              │
  │    j/k ↑↓  navigate    Enter  select    q  quit              │
  ╰──────────────────────────────────────────────────────────────╯
```

**Keyboard shortcuts:**
- `j` / `↓` — down
- `k` / `↑` — up
- `Enter` — select
- Letter keys (`P`, `S`, `R`, `D`, `C`, `W`) — jump directly
- `q` / `Esc` — quit / go back
- `?` — help overlay

## Screen-by-Screen Design

### 1. Home Screen (`screens/home.py`)
- ASCII art MONETA title in coral `#FF6F61`
- Vertical menu list with keyboard navigation
- Status footer (portfolio count, last scan time)
- Each menu item: coral `▸` on selected, letter shortcut in brackets

### 2. Portfolio Screen (`screens/portfolio.py`)
Sub-menu after selecting `[P]ortfolio`:
```
    ▸  [A]dd Holding
       [L]ist Holdings
       [R]emove Holding
       [C]hart (Donut)
       [B]ack to Home
```
- **Add Holding**: Textual Input widgets for symbol, shares, cost basis
- **List**: Rich table with all holdings
- **Remove**: Typing field + confirm dialog
- **Chart**: Full-screen donut widget

### 3. Donut Widget (`widgets/donut.py`)
Custom Textual widget using Unicode block characters. Calculates which characters fall in the donut ring area and colors them by slice. Shows labels and percentages.

### 4. Scan Screen (`screens/scan.py`)
- Progress bar during Finnhub API calls
- Results displayed as colored table
- Auto-updates per ticker

### 5. Report Screen (`screens/report.py`)
- Same output as `moneta check` but in Textual table
- Color-coded rows: green (>0.5), yellow (0.3-0.5), red (<0.3)
- Coral for alert flags and trend arrows

### 6. Dashboard Screen (`screens/dashboard.py`)
Split layout: 40% left (donut) / 60% right (sentiment summary)

### 7. Scheduler Screen (`screens/scheduler.py`)
- Install/uninstall/view launchd plist status

### 8. CLI Compatibility
- All CLI commands remain functional
- `moneta` with no args → launches TUI

## Global Constraints

- Python 3.12+ only, no type: ignore, no mypy suppressions
- All external API calls mockable via pytest-mock
- Runtime data directory: `~/.moneta/`
- No commit messages with co-authors
- No emojis in code or docs
- Plain dash "-" not em dash "---"
- Regular merges (no squash)
- Micro-commits per meaningful change
- Coral `#FF6F61` accent throughout
- Keyboard-first navigation
- No browser opens — everything in-terminal
- No new Python dependencies beyond Textual
