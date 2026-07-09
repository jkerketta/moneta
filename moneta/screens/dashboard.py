from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Horizontal
from textual.screen import Screen
from textual.widgets import Static, Footer
from rich.text import Text

from moneta.portfolio import list_holdings
from moneta.report import load_state
from moneta.widgets.donut import DonutWidget

CORAL = "#FF6F61"


class DashboardScreen(Screen):
    BINDINGS = [("escape", "back", "Back"), ("b", "back", "Back")]

    def compose(self) -> ComposeResult:
        yield Static("[bold]Dashboard[/bold]", id="title")
        yield Static("")
        yield Horizontal(
            Static("", id="donut-panel"),
            Static("", id="sentiment-panel"),
        )
        yield Footer()

    def on_mount(self) -> None:
        holdings = list_holdings()
        state = load_state()

        donut = DonutWidget(holdings, width=30, height=14)
        self.query_one("#donut-panel", Static).update(donut.render())

        sentiment_text = Text()
        has_data = False
        for h in holdings:
            history = state.get("history", {}).get(h.symbol, [])
            if not history:
                continue
            has_data = True
            latest = history[-1]
            composite = latest.get("composite", 0.5)
            trend = latest.get("trend", "flat")
            arrow = {"rising": "▲", "falling": "▼", "flat": "●"}.get(trend, "●")
            color = "green" if composite >= 0.7 else ("yellow" if composite >= 0.4 else "red")
            advice = latest.get("advice", "")
            alerted = latest.get("alerted", False)

            alert_mark = f" [{CORAL}]⚠[/{CORAL}]" if alerted else ""
            sentiment_text.append(f"{h.symbol}{alert_mark}\n", style="bold white")
            sentiment_text.append(f"  [{color}]{arrow} {composite:.2f}[/{color}]\n")
            sentiment_text.append(f"  {advice}\n\n")

        if not has_data:
            sentiment_text = Text("No scan data yet. Run scan first.", style="dim italic")

        self.query_one("#sentiment-panel", Static).update(sentiment_text)
