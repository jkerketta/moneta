from pathlib import Path
import pytest
from moneta.portfolio import Holding, load_holdings, save_holdings, add_holding, remove_holding, list_holdings


def test_holding_dataclass():
    h = Holding(symbol="TSLA", shares=10.0, cost_basis=250.0)
    assert h.symbol == "TSLA"
    assert h.shares == 10.0
    assert h.cost_basis == 250.0


def test_holding_default_cost_basis():
    h = Holding(symbol="TSLA", shares=10.0)
    assert h.cost_basis is None


def test_save_and_load_holdings(tmp_path: Path):
    holdings = [
        Holding(symbol="TSLA", shares=10.0, cost_basis=250.0),
        Holding(symbol="SPCE", shares=100.0),
    ]
    holdings_file = tmp_path / "holdings.yaml"
    save_holdings(holdings, holdings_file)
    loaded = load_holdings(holdings_file)
    assert loaded == holdings


def test_load_empty_holdings(tmp_path: Path):
    holdings_file = tmp_path / "holdings.yaml"
    loaded = load_holdings(holdings_file)
    assert loaded == []


def test_add_holding(tmp_path: Path):
    holdings_file = tmp_path / "holdings.yaml"
    add_holding("TSLA", 10.0, None, holdings_file)
    holdings = load_holdings(holdings_file)
    assert len(holdings) == 1
    assert holdings[0].symbol == "TSLA"
    assert holdings[0].shares == 10.0


def test_add_holding_updates_existing(tmp_path: Path):
    holdings_file = tmp_path / "holdings.yaml"
    add_holding("TSLA", 10.0, None, holdings_file)
    add_holding("TSLA", 5.0, None, holdings_file)
    holdings = load_holdings(holdings_file)
    assert len(holdings) == 1
    assert holdings[0].shares == 15.0


def test_remove_holding(tmp_path: Path):
    holdings_file = tmp_path / "holdings.yaml"
    add_holding("TSLA", 10.0, None, holdings_file)
    add_holding("SPCE", 100.0, None, holdings_file)
    result = remove_holding("TSLA", holdings_file)
    assert result is True
    holdings = load_holdings(holdings_file)
    assert len(holdings) == 1
    assert holdings[0].symbol == "SPCE"


def test_remove_nonexistent_holding(tmp_path: Path):
    holdings_file = tmp_path / "holdings.yaml"
    result = remove_holding("NONEXIST", holdings_file)
    assert result is False


def test_list_holdings(tmp_path: Path):
    holdings_file = tmp_path / "holdings.yaml"
    add_holding("TSLA", 10.0, None, holdings_file)
    add_holding("SPCE", 100.0, None, holdings_file)
    holdings = list_holdings(holdings_file)
    assert len(holdings) == 2
