# Moneta

Portfolio sentiment engine. A CLI tool that tracks stock holdings,

searches Reddit and Finnhub for sentiment signals,

generates per-ticker advice, and renders a terminal report.

## Tech Stack

- Python 3.12+
- Typer (CLI framework)
- Rich (terminal output, tables, colors, panels)
- PRAW (Reddit API)
- VADER Sentiment (local NLP)
- finnhub-python (news + social sentiment)
- pyyaml (holdings.yaml)
- python-dotenv (.env for secrets)
- matplotlib (allocation donut chart)
- pytest (testing)
- launchd (macOS scheduling, optional v1.1)

## Directory Layout

```
moneta/
├── moneta/
│   ├── __init__.py
│   ├── __main__.py          # python -m moneta
│   ├── cli.py               # Typer commands
│   ├── config.py            # Paths, env, client factories
│   ├── portfolio.py         # Holdings CRUD + chart
│   ├── sentiment.py         # PRAW + VADER
│   ├── news.py              # Finnhub APIs
│   ├── report.py            # Composite, state, alerts, advice
│   └── display.py           # Rich TUI formatting
├── tests/
│   ├── __init__.py
│   ├── conftest.py
│   ├── test_portfolio.py
│   ├── test_sentiment.py
│   ├── test_news.py
│   ├── test_report.py
│   ├── test_display.py
│   └── test_cli.py
├── docs/superpowers/plans/
│   └── 2026-07-08-moneta.md
├── .env.example
├── .gitignore
├── requirements.txt
├── AGENTS.md
└── README.md
```

Runtime data lives at `~/.moneta/`:

```
~/.moneta/
├── holdings.yaml     # user-editable portfolio
├── state.json        # auto-managed sentiment history
└── chart.png         # generated allocation donut
```

## Data Schemas

### holdings.yaml

```yaml
holdings:
  - symbol: TSLA
    shares: 10
    cost_basis: 250.00       # optional
```

### state.json

```json
{
  "history": {
    "TSLA": [
      {
        "date": "2026-07-07",
        "finnhub_news": 0.45,
        "finnhub_social": 0.52,
        "reddit_vader": 0.38,
        "composite": 0.45
      }
    ]
  }
}
```

## Scoring

### Composite formula

- Finnhub news sentiment: 0.0 - 1.0 (news_sentiment API)
- Finnhub social sentiment: 0.0 - 1.0 (stock_social_sentiment API)
- Reddit VADER: -1.0 to +1.0, normalized to 0-1: `(vader + 1) / 2`
- Composite = `(finnhub_news + finnhub_social + reddit_vader_norm) / 3`

If a source fails, skip it and average the remaining sources.

### Advice rules

Advice is generated from composite score + trend + source agreement.

| Condition | Advice |
|-----------|--------|
| composite > 0.8 | Excessive optimism - consider trimming |
| composite < 0.2 | Capitulation - potential contrarian entry |
| 0.6-0.8, rising | Strong momentum - hold or add on pullbacks |
| 0.6-0.8, falling | Positive but fading - don't add, monitor |
| 0.4-0.6 | No clear signal - wait for conviction |
| 0.2-0.4, falling | Negative accelerating - consider reducing |
| 0.2-0.4, rising | Early recovery signs - watch for confirmation |
| drop >30% day-over-day | Sharp reversal - review fundamentals |
| sources diverge >0.4 | Mixed signals - institutions vs retail disagree |

### Alert

Alert if composite drops >20% relative to the previous recorded composite.

### Trend detection

Use last 3 data points. Linear slope:

- |slope| < 0.02/day: flat
- slope >= 0.02/day: rising
- slope <= -0.02/day: falling

## CLI Commands

```
moneta portfolio add    TSLA --shares 10 [--cost 250]
moneta portfolio remove TSLA
moneta portfolio list
moneta portfolio chart
moneta scan                           # run sentiment pipeline, save state
moneta check                          # show report from cached state
moneta check --fresh                  # scan + show report
moneta watch --install                # daily 8am launchd plist
moneta watch --uninstall              # remove plist
moneta watch --status                 # show last/next scan time
```

## Testing

- pytest with pytest-mock for all external APIs
- External dependencies (PRAW, Finnhub, matplotlib, requests) are always mocked in tests
- Tests rely on temp directories (tmp_path fixture) for runtime files
- Typer commands tested via `typer.testing.CliRunner`

## .env

```
FINNHUB_API_KEY=your_key
REDDIT_CLIENT_ID=your_id
REDDIT_CLIENT_SECRET=your_secret
REDDIT_USER_AGENT=moneta/1.0 by u/your_username
```

## Running

```bash
python -m moneta portfolio list
python -m moneta scan
python -m moneta check
```

## Development

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
pytest
```
