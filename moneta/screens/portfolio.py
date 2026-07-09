from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.screen import Screen
from textual.widgets import Static, Input, Button, Footer, Label
from rich.text import Text

from moneta.config import ensure_dirs
from moneta.portfolio import add_holding, remove_holding, list_holdings, Holding
from moneta.widgets.donut import DonutWidget

CORAL = "#FF6F61"


class PortfolioScreen(Screen):
    BINDINGS = [
        ("a", "add", "Add"),
        ("l", "list_view", "List"),
        ("r", "remove", "Remove"),
        ("c", "chart", "Chart"),
        ("escape", "back", "Back"),
        ("b", "back", "Back"),
    ]

    def compose(self) -> ComposeResult:
        yield Static("[bold]Portfolio Management[/bold]", id="title")
        yield Static("")
        yield Static("  [A]dd Holding", style="bold white")
        yield Static("  [L]ist Holdings", style="bold white")
        yield Static("  [R]emove Holding", style="bold white")
        yield Static("  [C]hart (Donut)", style="bold white")
        yield Static("  [B]ack to Home", style=f"bold {CORAL}")
        yield Static("")
        yield Static("", id="content")
        yield Footer()

    def action_add(self) -> None:
        self.app.push_screen(AddHoldingScreen())

    def action_list_view(self) -> None:
        self.app.push_screen(ListHoldingsScreen())

    def action_remove(self) -> None:
        self.app.push_screen(RemoveHoldingScreen())

    def action_chart(self) -> None:
        self.app.push_screen(PortfolioChartScreen())

    def action_back(self) -> None:
        self.app.pop_screen()


class AddHoldingScreen(Screen):
    BINDINGS = [("escape", "back", "Back")]

    def compose(self) -> ComposeResult:
        yield Static("[bold]Add Holding[/bold]", id="title")
        yield Static("")
        yield Label("Symbol:")
        yield Input(placeholder="e.g. TSLA", id="symbol")
        yield Label("Shares:")
        yield Input(placeholder="e.g. 10", id="shares")
        yield Label("Cost basis per share (optional):")
        yield Input(placeholder="e.g. 250.00", id="cost")
        yield Static("")
        yield Button("Add", variant="primary", id="add-btn")
        yield Button("Cancel", id="cancel-btn")
        yield Static("", id="status")
        yield Footer()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "add-btn":
            self._do_add()
        else:
            self.app.pop_screen()

    def on_input_submitted(self, event: Input.Submitted) -> None:
        if event.input.id == "cost":
            self._do_add()
        elif event.input.id == "symbol":
            self.query_one("#shares", Input).focus()
        elif event.input.id == "shares":
            self.query_one("#cost", Input).focus()

    def _do_add(self) -> None:
        symbol = self.query_one("#symbol", Input).value.strip().upper()
        shares_str = self.query_one("#shares", Input).value.strip()
        cost_str = self.query_one("#cost", Input).value.strip()

        if not symbol:
            self.query_one("#status", Static).update(Text("Symbol is required.", style="bold red"))
            return
        if not shares_str:
            self.query_one("#status", Static).update(Text("Shares is required.", style="bold red"))
            return

        try:
            shares = float(shares_str)
        except ValueError:
            self.query_one("#status", Static).update(Text("Shares must be a number.", style="bold red"))
            return

        cost = None
        if cost_str:
            try:
                cost = float(cost_str)
            except ValueError:
                self.query_one("#status", Static).update(Text("Cost must be a number.", style="bold red"))
                return

        ensure_dirs()
        add_holding(symbol, shares, cost)
        self.query_one("#status", Static).update(
            Text(f"Added {shares} shares of {symbol}", style=f"bold {CORAL}")
        )
        self.query_one("#symbol", Input).value = ""
        self.query_one("#shares", Input).value = ""
        self.query_one("#cost", Input).value = ""
        self.query_one("#symbol", Input).focus()


class ListHoldingsScreen(Screen):
    BINDINGS = [("escape", "back", "Back"), ("b", "back", "Back")]

    def compose(self) -> ComposeResult:
        yield Static("[bold]Holdings[/bold]", id="title")
        yield Static("", id="content")
        yield Footer()

    def on_mount(self) -> None:
        holdings = list_holdings()
        if not holdings:
            text = Text("No holdings configured.", style="dim italic")
        else:
            text = Text()
            for h in holdings:
                cost = f" (cost: ${h.cost_basis:.2f})" if h.cost_basis else ""
                text.append(f"{h.symbol:<6} {h.shares:>8.1f} shares{cost}\n", style="bold white")
        self.query_one("#content", Static).update(text)

    def action_back(self) -> None:
        self.app.pop_screen()


class RemoveHoldingScreen(Screen):
    BINDINGS = [("escape", "back", "Back")]

    def compose(self) -> ComposeResult:
        yield Static("[bold]Remove Holding[/bold]", id="title")
        yield Static("")
        yield Label("Symbol to remove:")
        yield Input(placeholder="e.g. TSLA", id="symbol")
        yield Static("")
        yield Button("Remove", variant="error", id="remove-btn")
        yield Button("Cancel", id="cancel-btn")
        yield Static("", id="status")
        yield Footer()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "remove-btn":
            symbol = self.query_one("#symbol", Input).value.strip().upper()
            if not symbol:
                self.query_one("#status", Static).update(Text("Enter a symbol.", style="bold red"))
                return
            ensure_dirs()
            if remove_holding(symbol):
                self.query_one("#status", Static).update(Text(f"Removed {symbol}", style=f"bold {CORAL}"))
            else:
                self.query_one("#status", Static).update(Text(f"{symbol} not found.", style="bold red"))
            self.query_one("#symbol", Input).value = ""
        else:
            self.app.pop_screen()

    def on_input_submitted(self, event: Input.Submitted) -> None:
        if event.input.id == "symbol":
            self.query_one("#remove-btn", Button).press()

    def action_back(self) -> None:
        self.app.pop_screen()


class PortfolioChartScreen(Screen):
    BINDINGS = [("escape", "back", "Back"), ("b", "back", "Back"), ("q", "back", "Back")]

    def compose(self) -> ComposeResult:
        yield Static("[bold]Portfolio Allocation[/bold]", id="title")
        yield Static("", id="chart")

    def on_mount(self) -> None:
        holdings = list_holdings()
        chart = DonutWidget(holdings, width=40, height=18)
        self.query_one("#chart", Static).update(chart.render())

    def action_back(self) -> None:
        self.app.pop_screen()
