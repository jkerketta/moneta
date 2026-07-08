from unittest.mock import MagicMock, patch
import pytest
from moneta.sentiment import search_reddit_ticker, analyze_vader, get_reddit_sentiment


def test_analyze_vader_positive():
    texts = ["This stock is amazing, going to the moon!", "Best investment ever"]
    score = analyze_vader(texts)
    assert -1.0 <= score <= 1.0
    assert score > 0


def test_analyze_vader_negative():
    texts = ["This stock is terrible, complete garbage", "Worst decision ever"]
    score = analyze_vader(texts)
    assert score < 0


def test_analyze_vader_neutral():
    texts = ["The stock opened at $50 and closed at $51"]
    score = analyze_vader(texts)
    assert -0.5 <= score <= 0.5


def test_analyze_vader_empty():
    score = analyze_vader([])
    assert score == 0.0


def test_search_reddit_ticker_returns_texts():
    mock_reddit = MagicMock()
    mock_subreddit = MagicMock()
    mock_submission = MagicMock()
    mock_submission.title = "TSLA is looking great today"
    mock_submission.selftext = "I think this stock has strong fundamentals"
    mock_subreddit.search.return_value = [mock_submission]
    mock_reddit.subreddit.return_value = mock_subreddit

    texts = search_reddit_ticker("TSLA", ["wallstreetbets"], mock_reddit)
    assert len(texts) == 2
    assert "TSLA is looking great today" in texts
    assert "I think this stock has strong fundamentals" in texts
    mock_reddit.subreddit.assert_called_once_with("wallstreetbets+stocks+investing")
    mock_subreddit.search.assert_called_once_with("TSLA", limit=50)


def test_get_reddit_sentiment_returns_normalized():
    mock_reddit = MagicMock()
    mock_submission = MagicMock()
    mock_submission.title = "This stock is incredible!"
    mock_submission.selftext = "Best company out there, huge upside"
    mock_subreddit = MagicMock()
    mock_subreddit.search.return_value = [mock_submission]
    mock_reddit.subreddit.return_value = mock_subreddit

    score = get_reddit_sentiment("TSLA", ["wallstreetbets"], mock_reddit)
    assert 0.0 <= score <= 1.0
    assert score > 0.5  # positive texts


def test_get_reddit_sentiment_empty_results():
    mock_reddit = MagicMock()
    mock_subreddit = MagicMock()
    mock_subreddit.search.return_value = []
    mock_reddit.subreddit.return_value = mock_subreddit

    score = get_reddit_sentiment("TSLA", ["wallstreetbets"], mock_reddit)
    assert score == 0.5  # neutral
