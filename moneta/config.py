from __future__ import annotations

import os
from pathlib import Path

from dotenv import load_dotenv
import finnhub
import praw

MONETA_DIR = Path.home() / ".moneta"
HOLDINGS_FILE = MONETA_DIR / "holdings.yaml"
STATE_FILE = MONETA_DIR / "state.json"
CHART_FILE = MONETA_DIR / "chart.png"

DEFAULT_SUBREDDITS = ["wallstreetbets", "stocks", "investing"]
SENTIMENT_CACHE_HOURS = 24
ALERT_DROP_PERCENT = 0.20


def ensure_dirs():
    MONETA_DIR.mkdir(parents=True, exist_ok=True)


def load_env():
    load_dotenv()


def get_finnhub_client() -> finnhub.Client:
    api_key = os.environ.get("FINNHUB_API_KEY")
    if not api_key:
        raise ValueError("FINNHUB_API_KEY not set in .env or environment")
    return finnhub.Client(api_key=api_key)


def get_reddit_client() -> praw.Reddit:
    client_id = os.environ.get("REDDIT_CLIENT_ID")
    client_secret = os.environ.get("REDDIT_CLIENT_SECRET")
    user_agent = os.environ.get("REDDIT_USER_AGENT", "moneta/1.0")
    if not client_id or not client_secret:
        raise ValueError("REDDIT_CLIENT_ID and REDDIT_CLIENT_SECRET must be set")
    return praw.Reddit(
        client_id=client_id,
        client_secret=client_secret,
        user_agent=user_agent,
    )
