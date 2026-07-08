from __future__ import annotations

from typing import Optional

from moneta.config import DEFAULT_SUBREDDITS


def search_reddit_ticker(
    ticker: str,
    subreddits: Optional[list[str]] = None,
    reddit_client=None,
    limit: int = 50,
) -> list[str]:
    if subreddits is None:
        subreddits = DEFAULT_SUBREDDITS
    else:
        # Always merge passed subreddits with defaults (deduped, passed first)
        seen = set(subreddits)
        subreddits = list(subreddits)
        for s in DEFAULT_SUBREDDITS:
            if s not in seen:
                subreddits.append(s)
                seen.add(s)
    subreddit_name = "+".join(subreddits)
    subreddit = reddit_client.subreddit(subreddit_name)
    texts = []
    for submission in subreddit.search(ticker, limit=limit):
        texts.append(submission.title)
        if submission.selftext:
            texts.append(submission.selftext)
    return texts


def analyze_vader(texts: list[str]) -> float:
    if not texts:
        return 0.0
    from vaderSentiment.vaderSentiment import SentimentIntensityAnalyzer

    analyzer = SentimentIntensityAnalyzer()
    scores = [analyzer.polarity_scores(text)["compound"] for text in texts]
    return sum(scores) / len(scores)


def get_reddit_sentiment(
    ticker: str,
    subreddits: Optional[list[str]] = None,
    reddit_client=None,
) -> float:
    texts = search_reddit_ticker(ticker, subreddits, reddit_client)
    if not texts:
        return 0.5
    vader_score = analyze_vader(texts)
    return (vader_score + 1) / 2
