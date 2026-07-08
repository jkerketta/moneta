from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Optional

import yaml

from moneta.config import HOLDINGS_FILE, CHART_FILE


@dataclass
class Holding:
    symbol: str
    shares: float
    cost_basis: Optional[float] = None


def load_holdings(path: Path = HOLDINGS_FILE) -> list[Holding]:
    if not path.exists():
        return []
    with open(path) as f:
        data = yaml.safe_load(f)
    if not data or "holdings" not in data:
        return []
    return [Holding(**h) for h in data["holdings"]]


def save_holdings(holdings: list[Holding], path: Path = HOLDINGS_FILE) -> None:
    data = {
        "holdings": [
            {
                "symbol": h.symbol,
                "shares": h.shares,
                **({"cost_basis": h.cost_basis} if h.cost_basis is not None else {}),
            }
            for h in holdings
        ]
    }
    with open(path, "w") as f:
        yaml.dump(data, f, default_flow_style=False)


def add_holding(
    symbol: str,
    shares: float,
    cost_basis: Optional[float] = None,
    path: Path = HOLDINGS_FILE,
) -> None:
    holdings = load_holdings(path)
    existing = [h for h in holdings if h.symbol == symbol.upper()]
    if existing:
        existing[0].shares += shares
        if cost_basis is not None:
            existing[0].cost_basis = cost_basis
    else:
        holdings.append(Holding(symbol=symbol.upper(), shares=shares, cost_basis=cost_basis))
    save_holdings(holdings, path)


def remove_holding(symbol: str, path: Path = HOLDINGS_FILE) -> bool:
    holdings = load_holdings(path)
    filtered = [h for h in holdings if h.symbol != symbol.upper()]
    if len(filtered) == len(holdings):
        return False
    save_holdings(filtered, path)
    return True


def list_holdings(path: Path = HOLDINGS_FILE) -> list[Holding]:
    return load_holdings(path)


def generate_chart(holdings: list[Holding], path: Path = CHART_FILE) -> None:
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib.pyplot as plt

    symbols = [h.symbol for h in holdings]
    shares = [h.shares for h in holdings]

    fig, ax = plt.subplots(figsize=(8, 8))
    wedges, texts, autotexts = ax.pie(
        shares,
        labels=symbols,
        autopct="%1.1f%%",
        startangle=90,
        wedgeprops=dict(width=0.5),
    )
    ax.set_title("Portfolio Allocation")
    fig.savefig(path, dpi=150, bbox_inches="tight")
    plt.close(fig)
