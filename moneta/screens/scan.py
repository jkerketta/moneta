from __future__ import annotations

from textual.app import ComposeResult
from textual.screen import Screen
from textual.widgets import Static, Footer, ProgressBar

from moneta.config import ensure_dirs, load_env, get_finnhub_client
from moneta.portfolio import list_holdings
from moneta.report import run_scan

CORAL = "#FF6F61"


class ScanScreen(Screen):
    BINDINGS = [("escape", "back", "Back"), ("b", "back", "Back")]

    def __init__(self) -> None:
        super().__init__()
        self._done = False

    def compose(self) -> ComposeResult:
        yield Static("[bold]Scan Sentiment[/bold]", id="title")
        yield Static("")
        yield Static("", id="status")
        yield ProgressBar(total=100, id="progress", show_percentage=True)
        yield Static("", id="results")
        yield Footer()

    def on_mount(self) -> None:
        self.run_scan()

    def run_scan(self) -> None:
        import threading

        def _scan():
            ensure_dirs()
            load_env()
            holdings = list_holdings()
            if not holdings:
                self.call_from_thread(self._update_status, "No holdings configured. Add some first.")
                self.call_from_thread(self._done_display)
                return

            try:
                client = get_finnhub_client()
            except ValueError as e:
                self.call_from_thread(self._update_status, str(e))
                self.call_from_thread(self._done_display)
                return

            pb = self.query_one("#progress", ProgressBar)
            pb.update(total=len(holdings), progress=0)
            self.call_from_thread(self._update_status, f"Scanning {len(holdings)} holding(s)...")

            results = run_scan(ticker=None, finnhub_client=client)
            pb.update(progress=len(holdings))

            self.call_from_thread(self._update_status, "Scan complete!")
            text = ""
            for r in results:
                alert = "  ⚠" if r.get("alerted") else ""
                text += f"{r['ticker']}: composite {r['composite']:.2f} ({r.get('trend', 'flat')}){alert}\n"
                text += f"  {r.get('advice', '')}\n\n"
            self.call_from_thread(self._update_results, text)
            self.call_from_thread(self._done_display)

        self._update_status("Starting scan...")
        t = threading.Thread(target=_scan, daemon=True)
        t.start()

    def _update_status(self, msg: str) -> None:
        self.query_one("#status", Static).update(msg)

    def _update_results(self, text: str) -> None:
        self.query_one("#results", Static).update(text)

    def _done_display(self) -> None:
        self._done = True

    def action_back(self) -> None:
        self.app.pop_screen()
