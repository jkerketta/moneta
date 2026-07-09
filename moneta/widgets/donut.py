from __future__ import annotations

import math
from typing import Sequence

from rich.style import Style
from rich.text import Text, Span
from textual.widgets import Static

from moneta.portfolio import Holding

CORAL = "#FF6F61"
PALETTE = [
    "#FF6F61", "#4ECDC4", "#45B7D1", "#96CEB4",
    "#FFEAA7", "#DDA0DD", "#98D8C8", "#F7DC6F",
    "#BB8FCE", "#85C1E9", "#F0B27A", "#82E0AA",
]


def _render_donut(holdings: Sequence[Holding], width: int, height: int) -> Text:
    if not holdings:
        return Text("No holdings to chart.", style="dim italic")

    total = sum(h.shares for h in holdings)
    slices = [(h.symbol, h.shares / total) for h in holdings]

    chars = width
    rows = height
    px_w = chars
    px_h = rows * 2

    outer_r = min(px_w, px_h) / 2 - 0.5
    inner_r = outer_r * 0.55
    cx = px_w / 2 - 0.5
    cy = px_h / 2 - 0.5

    color_grid = [["#000000"] * px_w for _ in range(px_h)]

    def get_angle(x: float, y: float) -> float:
        return math.degrees(math.atan2(-(y - cy), x - cx)) % 360

    def get_slice_color(angle: float) -> str:
        cumulative = 0.0
        for i, (_, pct) in enumerate(slices):
            start_a = cumulative * 360
            end_a = (cumulative + pct) * 360
            if start_a <= angle < end_a:
                return PALETTE[i % len(PALETTE)]
            cumulative += pct
        return PALETTE[0]

    for py in range(px_h):
        for px in range(px_w):
            dx = px - cx
            dy = py - cy
            dist = math.sqrt(dx * dx + dy * dy)
            if inner_r <= dist <= outer_r:
                angle = get_angle(px, py)
                color_grid[py][px] = get_slice_color(angle)

    lines: list[Text] = []
    for row in range(rows):
        spans: list[Span] = []
        last_color = None
        span_start = 0
        for col in range(chars):
            top_color = color_grid[row * 2][col]
            bot_color = color_grid[row * 2 + 1][col] if row * 2 + 1 < px_h else "#000000"

            top_on = top_color != "#000000"
            bot_on = bot_color != "#000000"

            if top_on and bot_on:
                char = "█"
                fg = top_color
            elif top_on and not bot_on:
                char = "▀"
                fg = top_color
            elif not top_on and bot_on:
                char = "▄"
                fg = bot_color
            else:
                char = " "
                fg = "#000000"

            if fg != last_color:
                if last_color is not None:
                    spans.append(Span(span_start, col, Style(color=last_color)))
                span_start = col
                last_color = fg

        if last_color is not None:
            spans.append(Span(span_start, chars, Style(color=last_color)))

        line_text = Text(" " * chars)
        line_text.spans = spans
        lines.append(line_text)

    result = Text("\n").join(lines)
    result.append("\n\n")
    for i, (symbol, pct) in enumerate(slices):
        color = PALETTE[i % len(PALETTE)]
        block = Text("  ")
        block.spans = [Span(0, 2, Style(bgcolor=color))]
        result.append_text(block)
        result.append(f" {symbol:<6} {pct*100:.1f}%\n")

    return result


class DonutWidget(Static):
    def __init__(self, holdings: Sequence[Holding], width: int = 30, height: int = 15) -> None:
        super().__init__()
        self.holdings = list(holdings)
        self.donut_width = width
        self.donut_height = height

    def render(self) -> Text:
        return _render_donut(self.holdings, self.donut_width, self.donut_height)
