from __future__ import annotations

import webbrowser

import typer

from moneta import config
from moneta.config import ensure_dirs, load_env
from moneta.portfolio import add_holding, remove_holding, list_holdings, generate_chart

app = typer.Typer()
portfolio_app = typer.Typer()
app.add_typer(portfolio_app, name="portfolio", help="Manage your stock holdings.")


@portfolio_app.command()
def add(
    symbol: str = typer.Argument(None, help="Stock ticker symbol (omit for interactive)"),
    shares: float = typer.Option(None, "--shares", "-s", help="Number of shares"),
    cost: float = typer.Option(None, "--cost", "-c", help="Cost basis per share"),
):
    """Add a holding to your portfolio."""
    ensure_dirs()
    interactive = symbol is None
    if symbol is None:
        symbol = typer.prompt("Symbol").strip().upper()
    if shares is None:
        shares = float(typer.prompt("Shares"))
    if cost is None and interactive:
        cost_input = typer.prompt("Cost basis (optional)", default="")
        cost = float(cost_input) if cost_input.strip() else None
    add_holding(symbol, shares, cost)
    typer.echo(f"Added {shares} shares of {symbol.upper()}")


@portfolio_app.command()
def remove(
    symbol: str = typer.Argument(..., help="Stock ticker symbol"),
):
    """Remove a holding from your portfolio."""
    ensure_dirs()
    if remove_holding(symbol):
        typer.echo(f"Removed {symbol.upper()}")
    else:
        typer.echo(f"{symbol.upper()} not found in portfolio", err=True)
        raise typer.Exit(1)


@portfolio_app.command(name="list")
def list_all():
    """List all holdings in your portfolio."""
    ensure_dirs()
    holdings = list_holdings()
    if not holdings:
        typer.echo("No holdings configured. Use 'moneta portfolio add' to add one.")
        return
    for h in holdings:
        cost = f" (cost basis: ${h.cost_basis:.2f})" if h.cost_basis else ""
        typer.echo(f"{h.symbol}: {h.shares} shares{cost}")


@portfolio_app.command()
def chart():
    """Show allocation donut chart."""
    ensure_dirs()
    holdings = list_holdings()
    if not holdings:
        typer.echo("No holdings to chart. Add some first.")
        raise typer.Exit(1)
    generate_chart(holdings)
    webbrowser.open(config.CHART_FILE.as_uri())
    typer.echo(f"Chart saved to {config.CHART_FILE}")


@app.command()
def scan(
    ticker: str = typer.Option(None, "--ticker", "-t", help="Scan a single ticker only"),
):
    """Run sentiment scan and save results to state."""
    ensure_dirs()
    load_env()

    finnhub_client = config.get_finnhub_client()

    from moneta.report import run_scan

    results = run_scan(ticker, finnhub_client)
    if not results:
        typer.echo("No results. Check that you have holdings configured.")
        return

    from moneta.display import render_report

    render_report(results)


@app.command()
def check(
    fresh: bool = typer.Option(False, "--fresh", "-f", help="Run a fresh scan first"),
):
    """Show sentiment report from cached state."""
    ensure_dirs()

    if fresh:
        load_env()
        finnhub_client = config.get_finnhub_client()
        from moneta.report import run_scan

        results = run_scan(None, finnhub_client)
    else:
        from moneta.report import load_state
        from moneta.portfolio import list_holdings

        holdings = list_holdings()
        if not holdings:
            typer.echo("No holdings configured.")
            return

        state = load_state()
        results = []
        for h in holdings:
            history = state.get("history", {}).get(h.symbol, [])
            if not history:
                typer.echo(f"No cached data for {h.symbol}. Run 'moneta scan' first.")
                continue
            latest = history[-1].copy()
            latest["ticker"] = h.symbol
            results.append(latest)

    if not results:
        typer.echo("No data available. Run 'moneta scan' first.")
        return

    from moneta.display import render_report

    render_report(results)


@app.command()
def watch(
    action: str = typer.Argument("status", help="install, uninstall, or status"),
):
    """Manage daily scan scheduling via launchd."""
    from moneta.scheduler import install_plist, uninstall_plist, status

    if action == "install":
        install_plist()
        typer.echo("Daily scan scheduled at 8:00 AM.")
    elif action == "uninstall":
        uninstall_plist()
        typer.echo("Scheduled scan removed.")
    elif action == "status":
        typer.echo(status())
    else:
        typer.echo("Usage: moneta watch [install|uninstall|status]", err=True)
        raise typer.Exit(1)
