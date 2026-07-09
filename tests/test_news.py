from unittest.mock import MagicMock
import pytest
from moneta.news import get_finnhub_news_sentiment, get_finnhub_social_sentiment


def test_news_sentiment_returns_score():
    mock_client = MagicMock()
    mock_client.news_sentiment.return_value = {
        "companyNewsScore": 0.75,
        "sentiment": {"bearishPercent": 0.2, "bullishPercent": 0.8},
        "buzz": {"articlesInLastWeek": 100},
        "symbol": "TSLA",
    }
    score = get_finnhub_news_sentiment("TSLA", mock_client)
    assert score == 0.75
    mock_client.news_sentiment.assert_called_once_with("TSLA")


def test_news_sentiment_missing_key():
    mock_client = MagicMock()
    mock_client.news_sentiment.return_value = {}
    score = get_finnhub_news_sentiment("TSLA", mock_client)
    assert score == 0.5


def test_social_sentiment_returns_score():
    mock_client = MagicMock()
    mock_client.stock_social_sentiment.return_value = {
        "sentiment": {"bearishPercent": 0.3, "bullishPercent": 0.7},
    }
    score = get_finnhub_social_sentiment("TSLA", mock_client)
    assert 0.0 <= score <= 1.0
    mock_client.stock_social_sentiment.assert_called_once_with("TSLA")


def test_social_sentiment_empty():
    mock_client = MagicMock()
    mock_client.stock_social_sentiment.return_value = {}
    score = get_finnhub_social_sentiment("TSLA", mock_client)
    assert score == 0.5
