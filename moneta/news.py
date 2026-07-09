from typing import Any


def get_finnhub_news_sentiment(ticker: str, client: Any = None) -> float:
    try:
        data = client.news_sentiment(ticker)
        return data.get("companyNewsScore", 0.5)
    except Exception:
        return 0.5


def get_finnhub_social_sentiment(ticker: str, client: Any = None) -> float:
    try:
        data = client.stock_social_sentiment(ticker)
        sentiment = data.get("sentiment", {})
        bullish = sentiment.get("bullishPercent", 50)
        return bullish / 100.0
    except Exception:
        return 0.5
