from __future__ import annotations

from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.text import Text
from rich.layout import Layout
from rich.columns import Columns
from rich import box


def _bar(value: float, width: int = 20) -> str:
    filled = int(value * width)
    bar = "█" * filled + "░" * (width - filled)
    return f"[green]{bar}[/green]" if value >= 0.5 else f"[red]{bar}[/red]"


def _trend_arrow(trend: str) -> str:
    return {"rising": "▲", "falling": "▼", "flat": "●"}.get(trend, "●")


def format_report(results: list[dict]) -> str:
    if not results:
        return "[yellow]No holdings configured. Run 'moneta portfolio add' first.[/yellow]"

    table = Table(box=box.SQUARE, show_header=False, padding=(0, 1))

    for r in results:
        ticker = r["ticker"]
        composite = r["composite"]
        trend = r["trend"]
        arrow = _trend_arrow(trend)
        change = r.get("change")
        alerted = r.get("alerted", False)

        change_str = f"  {arrow} {change:+.0f}%" if change is not None else ""
        alert_flag = "  ⚠ ALERT" if alerted else ""

        header = Text(f"{ticker}{change_str}{alert_flag}", style="bold")
        if alerted:
            header.stylize("bold red")

        bar = _bar(composite)

        sources = (
            f"Finnhub: {r.get('finnhub_news', 'N/A') or 'N/A'}  "
            f"Social: {r.get('finnhub_social', 'N/A') or 'N/A'}  "
            f"Reddit: {r.get('reddit_vader', 'N/A') or 'N/A'}"
        )
        advice = r.get("advice", "")

        body = Text()
        body.append(f"Composite: {bar}  {composite:.2f}\n")
        body.append(f"{sources}\n")
        body.append(f"Advice:    {advice}")

        color = "red" if alerted else ("green" if composite >= 0.5 else "yellow")
        table.add_row(
            Panel(body, title=header, border_style=color, padding=(0, 1))
        )

    console = Console(width=120)
    segments = console.render(table, console.options)
    return "".join(s.text for s in segments)


def render_report(results: list[dict]) -> None:
    console = Console()
    report = format_report(results)
    console.print(f"[bold cyan]Moneta Daily Sentiment Report[/bold cyan]\n\n{report}")
