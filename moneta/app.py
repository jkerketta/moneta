from __future__ import annotations

from textual.app import App
from textual.binding import Binding

from moneta.screens.home import HomeScreen


class MonetaApp(App):
    CSS = """
    Screen {
        background: $surface;
    }
    """

    BINDINGS = [
        Binding("q", "quit", "Quit", show=False),
    ]

    SCREENS = {"home": HomeScreen}

    def on_mount(self) -> None:
        self.push_screen("home")
