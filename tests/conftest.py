from pathlib import Path
from typing import Generator
from unittest.mock import MagicMock

import pytest


@pytest.fixture
def mock_finnhub_client() -> MagicMock:
    client = MagicMock()
    client.news_sentiment.return_value = {
        "companyNewsScore": 0.5,
        "sentiment": {"bearishPercent": 0.3, "bullishPercent": 0.7},
        "buzz": {"articlesInLastWeek": 100},
        "symbol": "TSLA",
    }
    client.stock_social_sentiment.return_value = {
        "sentiment": {"bearishPercent": 0.4, "bullishPercent": 0.6},
    }
    return client


@pytest.fixture
def mock_reddit_client() -> MagicMock:
    client = MagicMock()
    submission = MagicMock()
    submission.title = "TSLA is looking great today"
    submission.selftext = "Strong fundamentals and good outlook"
    subreddit = MagicMock()
    subreddit.search.return_value = [submission]
    client.subreddit.return_value = subreddit
    return client


@pytest.fixture
def temp_state_file(tmp_path: Path) -> Generator[Path, None, None]:
    state_file = tmp_path / "state.json"
    state_file.write_text('{"history": {}}')
    yield state_file


@pytest.fixture
def temp_holdings_file(tmp_path: Path) -> Generator[Path, None, None]:
    holdings_file = tmp_path / "holdings.yaml"
    holdings_file.write_text("holdings:\n  - symbol: TSLA\n    shares: 10\n")
    yield holdings_file
