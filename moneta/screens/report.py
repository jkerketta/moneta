from __future__ import annotations

from textual.app import ComposeResult
from textual.screen import Screen
from textual.widgets import Static, Footer
from rich.table import Table
from rich.text import Text
from rich.console import Console

from moneta.portfolio import list_holdings
from moneta.report import load_state

CORAL = "#FF6F61"


class ReportScreen(Screen):
    BINDINGS = [("escape", "back", "Back"), ("b", "back", "Back")]

    def compose(self) -> ComposeResult:
        yield Static("[bold]Sentiment Report[/bold]", id="title")
        yield Static("", id="content")
        yield Footer()

    def on_mount(self) -> None:
        state = load_state()
        holdings = list_holdings()

        if not holdings:
            self.query_one("#content", Static).update(Text("No holdings configured.", style="dim italic"))
            return

        table = Table(box=None, show_header=False, padding=(0, 2))
        for h in holdings:
            history = state.get("history", {}).get(h.symbol, [])
            if not history:
                continue
            latest = history[-1]
            composite = latest.get("composite", 0.5)
            trend = latest.get("trend", "flat")
            arrow = {"rising": "▲", "falling": "▼", "flat": "●"}.get(trend, "●")
            advice = latest.get("advice", "")
            alert = latest.get("alerted", False)

            color = "green" if composite >= 0.7 else ("yellow" if composite >= 0.4 else "red")
            alert_flag = f"  [{CORAL}]⚠ ALERT[/{CORAL}]" if alert else ""
            trend_style = f"style={color}"

            table.add_row(
                f"[bold]{h.symbol}[/bold]{alert_flag}",
                f"[{color}]{arrow} {composite:.2f}[/{color}]",
                f"[dim]{advice}[/dim]",
            )

        if not table.rows:
            self.query_one("#content", Static).update(
                Text("No scan data yet. Run scan first.", style="dim italic")
            )
            return

        console = Console(width=80, force_terminal=True)
        with console.capture() as capture:
            console.print(table)
        self.query_one("#content", Static).update(capture.get())
