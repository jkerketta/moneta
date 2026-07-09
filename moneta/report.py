from __future__ import annotations

import json
from datetime import date
from pathlib import Path
from typing import Any, Optional

import numpy as np

from moneta.config import STATE_FILE, MONETA_DIR
from moneta.news import get_finnhub_news_sentiment, get_finnhub_social_sentiment


def compute_composite(
    finnhub_news: Optional[float],
    finnhub_social: Optional[float],
) -> float:
    scores = [s for s in [finnhub_news, finnhub_social] if s is not None]
    if not scores:
        return 0.5
    return sum(scores) / len(scores)


def load_state(path: Path = STATE_FILE) -> dict:
    if not path.exists():
        return {"history": {}}
    with open(path) as f:
        return json.load(f)


def save_state(state: dict, path: Path = STATE_FILE) -> None:
    MONETA_DIR.mkdir(parents=True, exist_ok=True)
    with open(path, "w") as f:
        json.dump(state, f, indent=2)


def get_history(ticker: str, state: dict) -> list[dict]:
    return state.get("history", {}).get(ticker.upper(), [])


def record_sentiment(
    ticker: str,
    result: dict,
    path: Path = STATE_FILE,
) -> None:
    state = load_state(path)
    ticker = ticker.upper()
    if ticker not in state["history"]:
        state["history"][ticker] = []
    state["history"][ticker].append(result)
    save_state(state, path)


def detect_trend(history: list[dict]) -> tuple[str, float]:
    if len(history) < 2:
        return "flat", 0.0
    composites = [h["composite"] for h in history[-3:]]
    if len(composites) < 2:
        return "flat", 0.0
    x = list(range(len(composites)))
    slope = np.polyfit(x, composites, 1)[0]
    if abs(slope) < 0.02:
        return "flat", slope
    return ("rising", slope) if slope > 0 else ("falling", slope)


def check_alert(
    ticker: str,
    composite: float,
    path: Path = STATE_FILE,
) -> tuple[bool, Optional[float]]:
    state = load_state(path)
    history = get_history(ticker, state)
    if not history:
        return False, None
    previous = history[-1]["composite"]
    if previous == 0:
        return False, None
    percent_change = ((composite - previous) / previous) * 100
    if percent_change < -20:
        return True, percent_change
    return False, percent_change


def generate_advice(
    ticker: str,
    composite: float,
    sources: dict[str, Optional[float]],
    history: list[dict],
) -> str:
    direction, _ = detect_trend(history)

    if composite > 0.8:
        return "Excessive optimism - consider taking some profits or trimming."
    if composite < 0.2:
        return "Market capitulation - potential contrarian entry point if fundamentals intact."

    if history and len(history) >= 1:
        prev = history[-1]["composite"]
        if prev > 0:
            change = ((composite - prev) / prev) * 100
            if change > 30:
                return "Euphoria spike - new positions are high risk."
            if change < -30:
                return "Sharp sentiment reversal - review fundamentals urgently."

    if 0.6 <= composite <= 0.8:
        if direction == "rising":
            return "Strong positive momentum - hold or add on pullbacks."
        if direction == "falling":
            return "Positive but fading - monitor for trend change, don't add."
        return "Healthy sentiment - maintain current position."

    if 0.2 < composite < 0.4:
        if direction == "falling":
            return "Negative trend accelerating - consider reducing exposure."
        if direction == "rising":
            return "Early recovery signs - potential bottom, watch for confirmation."
        return "Persistent negativity - risk remains, tight stops advised."

    return "No clear sentiment signal - wait for conviction."


def run_scan(
    ticker: Optional[str] = None,
    finnhub_client: Any = None,
    path: Path = STATE_FILE,
) -> list[dict]:
    from moneta.portfolio import list_holdings

    holdings = list_holdings()
    if ticker:
        holdings = [h for h in holdings if h.symbol == ticker.upper()]
        if not holdings:
            return []

    results = []
    today = date.today().isoformat()

    for holding in holdings:
        try:
            finnhub_news = get_finnhub_news_sentiment(holding.symbol, finnhub_client)
        except Exception:
            finnhub_news = None

        try:
            finnhub_social = get_finnhub_social_sentiment(holding.symbol, finnhub_client)
        except Exception:
            finnhub_social = None

        composite = compute_composite(finnhub_news, finnhub_social)

        result = {
            "date": today,
            "ticker": holding.symbol,
            "finnhub_news": finnhub_news,
            "finnhub_social": finnhub_social,
            "composite": composite,
        }

        record_sentiment(holding.symbol, result, path)

        state = load_state(path)
        history = get_history(holding.symbol, state)

        alerted, change = check_alert(holding.symbol, composite, path)

        result["alerted"] = alerted
        result["change"] = change
        result["trend"], result["slope"] = detect_trend(history)
        result["advice"] = generate_advice(
            holding.symbol,
            composite,
            {"finnhub_news": finnhub_news, "finnhub_social": finnhub_social},
            history[:-1],
        )

        results.append(result)

    return results
