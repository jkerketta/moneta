from __future__ import annotations

from textual.app import ComposeResult
from textual.screen import Screen
from textual.widgets import Static, Footer

from moneta.scheduler import status, install_plist, uninstall_plist

CORAL = "#FF6F61"


class SchedulerScreen(Screen):
    BINDINGS = [
        ("i", "install", "Install"),
        ("u", "uninstall", "Uninstall"),
        ("escape", "back", "Back"),
        ("b", "back", "Back"),
    ]

    def compose(self) -> ComposeResult:
        yield Static("[bold]Daily Scan Schedule[/bold]", id="title")
        yield Static("")
        yield Static("", id="current-status")
        yield Static("")
        yield Static("  [I]nstall   Set up daily scan at 8:00 AM", style="bold white")
        yield Static("  [U]ninstall Remove scheduled scan", style="bold white")
        yield Static("  [B]ack      Return to home", style=f"bold {CORAL}")
        yield Static("")
        yield Static("", id="action-status")
        yield Footer()

    def on_mount(self) -> None:
        self._refresh_status()

    def _refresh_status(self) -> None:
        self.query_one("#current-status", Static).update(f"Current: {status()}")

    def action_install(self) -> None:
        try:
            install_plist()
            self.query_one("#action-status", Static).update(
                f"[bold {CORAL}]Daily scan installed at 8:00 AM.[/bold {CORAL}]"
            )
        except Exception as e:
            self.query_one("#action-status", Static).update(f"[bold red]Error: {e}[/bold red]")
        self._refresh_status()

    def action_uninstall(self) -> None:
        try:
            uninstall_plist()
            self.query_one("#action-status", Static).update("Scheduled scan removed.")
        except Exception as e:
            self.query_one("#action-status", Static).update(f"[bold red]Error: {e}[/bold red]")
        self._refresh_status()

    def action_back(self) -> None:
        self.app.pop_screen()
