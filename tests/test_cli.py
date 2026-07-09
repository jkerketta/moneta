from pathlib import Path
from unittest.mock import patch, MagicMock
import pytest
from typer.testing import CliRunner
from moneta.cli import app

runner = CliRunner()


def test_portfolio_add():
    with patch("moneta.cli.add_holding") as mock_add:
        result = runner.invoke(app, ["portfolio", "add", "TSLA", "--shares", "10"])
        assert result.exit_code == 0
        mock_add.assert_called_once_with("TSLA", 10.0, None)


def test_portfolio_add_with_cost():
    with patch("moneta.cli.add_holding") as mock_add:
        result = runner.invoke(app, ["portfolio", "add", "TSLA", "--shares", "10", "--cost", "250"])
        assert result.exit_code == 0
        mock_add.assert_called_once_with("TSLA", 10.0, 250.0)


def test_portfolio_remove():
    with patch("moneta.cli.remove_holding", return_value=True) as mock_remove:
        result = runner.invoke(app, ["portfolio", "remove", "TSLA"])
        assert result.exit_code == 0
        mock_remove.assert_called_once_with("TSLA")


def test_portfolio_list():
    with patch("moneta.cli.list_holdings", return_value=[]) as mock_list:
        result = runner.invoke(app, ["portfolio", "list"])
        assert result.exit_code == 0
        mock_list.assert_called_once()


def test_portfolio_chart():
    holdings = [MagicMock(symbol="TSLA", shares=10.0)]
    with patch("moneta.cli.list_holdings", return_value=holdings), \
         patch("moneta.cli.generate_chart") as mock_chart, \
         patch("webbrowser.open") as mock_open:
        result = runner.invoke(app, ["portfolio", "chart"])
        assert result.exit_code == 0
        mock_chart.assert_called_once()
        mock_open.assert_called_once()


def test_scan():
    mock_result = {
        "ticker": "TSLA", "composite": 0.5, "trend": "flat",
        "advice": "Hold steady.", "alerted": False,
        "change": None, "finnhub_news": 0.5,
        "finnhub_social": 0.5,
    }
    mock_report = MagicMock(return_value=[mock_result])
    with patch("moneta.config.get_finnhub_client"), \
         patch("moneta.report.run_scan", mock_report):
        result = runner.invoke(app, ["scan"])
        assert result.exit_code == 0


def test_check():
    with patch("moneta.report.load_state", return_value={"history": {}}), \
         patch("moneta.portfolio.list_holdings", return_value=[]):
        result = runner.invoke(app, ["check"])
        assert result.exit_code == 0
