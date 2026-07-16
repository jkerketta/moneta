<h1 align="center">moneta</h1>
<p align="center"><em>Your portfolio, live in your terminal.</em></p>

<p align="center">
  <img src="https://img.shields.io/badge/go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white">
  <img src="https://img.shields.io/github/license/jkerketta/moneta?style=flat-square">
  <img src="https://img.shields.io/badge/platform-linux%20|%20macOS%20|%20windows-lightgrey?style=flat-square">
</p>

<p align="center">
  <img src="screenshots/Screenshot%202026-07-16%20at%202.35.38 PM.png" width="80%">
</p>

---

A terminal-native portfolio tracker for developers who live in the command line.
Track your holdings, browse real-time market news, check sentiment, and flip
through interactive price charts — all without leaving your terminal.

## ✨ Features

- **Portfolio tracking** — P&L, allocation donut, per-position performance
- **Live market data** — S&amp;P 500, Dow, NASDAQ, crude oil, gold at a glance
- **Market news + sentiment** — real-time headlines with bullish/bearish scoring
- **Ticker-specific news** — press `n` on any holding for a headline detail card
- **Interactive charts** — line, area, and candlestick views with sparklines
- **5 themes** — Rose Pine Moon, Golden Hour, Coral Reef, Emerald Forest, Midnight Ice
- **Zero API keys** — all data from Yahoo Finance's public endpoints
- **Vim keys** — `j`/`k` navigation, modal overlays, keyboard-first design

## ⚡ Quick Start

```bash
git clone https://github.com/jkerketta/moneta.git
cd moneta
go build ./cmd/stock-tui
./stock-tui
```

Then add your first position with `a`.

## 🎨 Themes

Press `Change Theme` from the home menu to preview and switch between five
themes — each with its own accent color and decorative icon set.

| Theme | Accent | Icons |
|-------|--------|-------|
| Rose Pine Moon | `#c4a7e7` purple | ✿ ❀ ✾ · |
| Golden Hour | `#f6c177` yellow | ★ ☽ ✧ ✦ |
| Coral Reef | `#e8736a` coral | ❂ ✧ ▹ · |
| Emerald Forest | `#4a9c5d` green | ☘ ❧ ✻ ✧ |
| Midnight Ice | `#7ec8e3` blue | ❄ ✧ ◇ · |

## ⌨️ Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `↵` | Open chart |
| `n` | Ticker news detail |
| `a` | Add position |
| `d` | Remove position |
| `r` | Refresh data |
| `?` | Help overlay |
| `←` / `→` | Browse market news pages |
| `1` / `2` / `3` / `4` | Chart range (1h, 24h, 7d, 30d) |
| `Tab` | Cycle chart type (line → area → candle) |
| `Esc` | Close overlay / go back |
| `q` | Quit |

<p align="center">
  <img src="screenshots/Screenshot%202026-07-16%20at%202.36.30 PM.png" width="80%">
</p>

## ⚙️ Configuration

Create a `portfolio.yaml` in the project root:

```yaml
holdings:
  - symbol: GOOGL
    shares: 5
    avg_price: 180
    currency: USD
  - symbol: AAPL
    shares: 10
    avg_price: 150
    currency: USD
```

Or press `a` in the app to add positions interactively — the file is updated on
every change.

## ⚠️ Data Source

All market data comes from **Yahoo Finance's public endpoints**. No API key
required, no rate limits to manage. The app respects the API and batches
requests efficiently. For best results, avoid refreshing more than once
every 5–10 seconds.

## 🏗️ Development

```bash
go run ./cmd/stock-tui    # Run
go build ./cmd/stock-tui  # Build
go vet ./...              # Lint
```

## 📄 License

[MIT](LICENSE)
