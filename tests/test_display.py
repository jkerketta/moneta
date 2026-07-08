# tests/test_display.py
from unittest.mock import patch
import pytest
from moneta.display import format_report, render_report


def test_format_report_contains_tickers():
    results = [
        {
            "ticker": "TSLA",
            "composite": 0.58,
            "finnhub_news": 0.62,
            "finnhub_social": 0.55,
            "reddit_vader": 0.57,
            "alerted": False,
            "change": 12.0,
            "trend": "rising",
            "advice": "Healthy sentiment - maintain current position.",
        },
        {
            "ticker": "SPCE",
            "composite": 0.31,
            "finnhub_news": 0.35,
            "finnhub_social": 0.28,
            "reddit_vader": 0.30,
            "alerted": True,
            "change": -34.0,
            "trend": "falling",
            "advice": "Negative trend accelerating - consider reducing exposure.",
        },
    ]
    output = format_report(results)
    assert "TSLA" in output
    assert "SPCE" in output
    assert "ALERT" in output or "alert" in output.lower()


def test_format_report_empty():
    output = format_report([])
    assert output is not None


def test_render_report_calls_rich():
    results = [{
        "ticker": "TSLA",
        "composite": 0.58,
        "finnhub_news": 0.62,
        "finnhub_social": 0.55,
        "reddit_vader": 0.57,
        "alerted": False,
        "change": 5.0,
        "trend": "rising",
        "advice": "Hold current position.",
    }]
    with patch("rich.console.Console.print") as mock_print:
        render_report(results)
        mock_print.assert_called_once()
