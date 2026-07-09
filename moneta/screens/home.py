from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.screen import Screen
from textual.widgets import Static, Footer
from rich.text import Text

from moneta.widgets.title import MonetaTitle

CORAL = "#FF6F61"

MENU_ITEMS = [
    ("P", "Portfolio", "Manage your holdings"),
    ("S", "Scan Sentiment", "Pull fresh Finnhub data"),
    ("R", "Report", "View latest sentiment"),
    ("D", "Dashboard", "Portfolio + sentiment at a glance"),
    ("C", "Chart", "Allocation donut"),
    ("W", "Watch", "Schedule daily scans"),
]


class MenuItem(Static):
    def __init__(self, key: str, label: str, desc: str, selected: bool = False) -> None:
        super().__init__()
        self.item_key = key
        self.item_label = label
        self.item_desc = desc
        self._selected = selected

    def render(self) -> Text:
        arrow = Text(f"  {'▸' if self._selected else ' '}  ", style=f"bold {CORAL}" if self._selected else "dim")
        key = Text(f"[{self.item_key}]", style=f"bold {CORAL}")
        label = Text(f"  {self.item_label}", style="bold white")
        spacer = " " * max(1, 18 - len(self.item_label))
        desc = Text(self.item_desc, style="dim white")
        return Text.assemble(arrow, key, label, spacer, desc)


class HomeScreen(Screen):
    BINDINGS = [
        ("j", "cursor_down", "Down"),
        ("k", "cursor_up", "Up"),
        ("down", "cursor_down", "Down"),
        ("up", "cursor_up", "Up"),
        ("enter", "select", "Select"),
        ("p", "action_portfolio", "Portfolio"),
        ("s", "action_scan", "Scan"),
        ("r", "action_report", "Report"),
        ("d", "action_dashboard", "Dashboard"),
        ("c", "action_chart", "Chart"),
        ("w", "action_watch", "Watch"),
    ]

    def __init__(self) -> None:
        super().__init__()
        self._selected_index = 0

    def compose(self) -> ComposeResult:
        yield MonetaTitle()
        yield Static("     Portfolio Sentiment Engine", style=f"bold {CORAL}")
        yield Static("")
        self.menu_items: list[MenuItem] = []
        for key, label, desc in MENU_ITEMS:
            item = MenuItem(key, label, desc)
            self.menu_items.append(item)
            yield item
        yield Static("")
        yield Static("  j/k \u2191\u2193 navigate    Enter select    letter jump    q quit", style="dim white")
        yield Footer()

    def on_mount(self) -> None:
        self._update_selection()

    def _update_selection(self) -> None:
        for i, item in enumerate(self.menu_items):
            item._selected = i == self._selected_index
            item.refresh()

    def action_cursor_down(self) -> None:
        if self._selected_index < len(self.menu_items) - 1:
            self._selected_index += 1
            self._update_selection()

    def action_cursor_up(self) -> None:
        if self._selected_index > 0:
            self._selected_index -= 1
            self._update_selection()

    def action_select(self) -> None:
        key = self.menu_items[self._selected_index].item_key
        self._handle_key(key)

    def action_portfolio(self) -> None:
        self._handle_key("P")

    def action_scan(self) -> None:
        self._handle_key("S")

    def action_report(self) -> None:
        self._handle_key("R")

    def action_dashboard(self) -> None:
        self._handle_key("D")

    def action_chart(self) -> None:
        self._handle_key("C")

    def action_watch(self) -> None:
        self._handle_key("W")

    def _handle_key(self, key: str) -> None:
        from moneta.screens.portfolio import PortfolioScreen
        if key == "P":
            self.app.push_screen(PortfolioScreen())
        elif key == "S":
            from moneta.screens.scan import ScanScreen
            self.app.push_screen(ScanScreen())
        elif key == "R":
            from moneta.screens.report import ReportScreen
            self.app.push_screen(ReportScreen())
        elif key == "D":
            from moneta.screens.dashboard import DashboardScreen
            self.app.push_screen(DashboardScreen())
        elif key == "C":
            from moneta.screens.portfolio import PortfolioChartScreen
            self.app.push_screen(PortfolioChartScreen())
        elif key == "W":
            from moneta.screens.scheduler import SchedulerScreen
            self.app.push_screen(SchedulerScreen())
