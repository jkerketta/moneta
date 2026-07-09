# tests/test_report.py
import json
from pathlib import Path
from unittest.mock import MagicMock
import pytest
from moneta.report import (
    compute_composite,
    load_state,
    save_state,
    get_history,
    record_sentiment,
    detect_trend,
    check_alert,
    generate_advice,
)


def test_compute_composite():
    score = compute_composite(0.8, 0.6)
    assert score == pytest.approx(0.7)


def test_compute_composite_with_none():
    score = compute_composite(0.8, None)
    assert score == pytest.approx(0.8)


def test_compute_composite_all_none():
    score = compute_composite(None, None)
    assert score == 0.5


def test_load_state_missing(tmp_path: Path):
    state = load_state(tmp_path / "nonexistent.json")
    assert state == {"history": {}}


def test_save_and_load_state(tmp_path: Path):
    path = tmp_path / "state.json"
    state = {"history": {"TSLA": [{"composite": 0.5}]}}
    save_state(state, path)
    loaded = load_state(path)
    assert loaded == state


def test_record_sentiment(tmp_path: Path):
    path = tmp_path / "state.json"
    save_state({"history": {}}, path)
    record_sentiment("TSLA", {
        "finnhub_news": 0.6,
        "finnhub_social": 0.5,
        "composite": 0.5,
    }, path)
    state = load_state(path)
    assert "TSLA" in state["history"]
    assert len(state["history"]["TSLA"]) == 1


def test_get_history():
    state = {"history": {"TSLA": [{"composite": 0.5}]}}
    history = get_history("TSLA", state)
    assert len(history) == 1


def test_get_history_empty():
    state = {"history": {}}
    history = get_history("NONEXIST", state)
    assert history == []


def test_detect_trend_rising():
    history = [
        {"composite": 0.3},
        {"composite": 0.4},
        {"composite": 0.5},
    ]
    direction, slope = detect_trend(history)
    assert direction == "rising"
    assert slope > 0.02


def test_detect_trend_falling():
    history = [
        {"composite": 0.6},
        {"composite": 0.5},
        {"composite": 0.4},
    ]
    direction, slope = detect_trend(history)
    assert direction == "falling"
    assert slope < -0.02


def test_detect_trend_flat():
    history = [
        {"composite": 0.5},
        {"composite": 0.51},
        {"composite": 0.5},
    ]
    direction, slope = detect_trend(history)
    assert direction == "flat"


def test_detect_trend_insufficient():
    direction, slope = detect_trend([])
    assert direction == "flat"
    assert slope == 0.0


def test_check_alert_triggered(tmp_path: Path):
    path = tmp_path / "state.json"
    save_state({
        "history": {"TSLA": [{"composite": 0.7}]}
    }, path)
    alerted, drop = check_alert("TSLA", 0.5, path)
    assert alerted is True
    assert drop is not None
    assert drop < -20


def test_check_alert_not_triggered(tmp_path: Path):
    path = tmp_path / "state.json"
    save_state({
        "history": {"TSLA": [{"composite": 0.5}]}
    }, path)
    alerted, drop = check_alert("TSLA", 0.45, path)
    assert alerted is False


def test_check_alert_no_history(tmp_path: Path):
    path = tmp_path / "state.json"
    save_state({"history": {}}, path)
    alerted, drop = check_alert("TSLA", 0.5, path)
    assert alerted is False
    assert drop is None


def test_generate_advice_extreme_bullish():
    advice = generate_advice("TSLA", 0.85, {}, [])
    assert "trimming" in advice.lower() or "trim" in advice.lower()


def test_generate_advice_extreme_bearish():
    advice = generate_advice("TSLA", 0.15, {}, [])
    assert "contrarian" in advice.lower() or "capitulation" in advice.lower()


def test_generate_advice_neutral():
    advice = generate_advice("TSLA", 0.5, {}, [])
    assert "wait" in advice.lower() or "conviction" in advice.lower()



