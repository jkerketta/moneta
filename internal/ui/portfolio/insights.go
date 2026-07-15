package portfolio

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/data"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

// moverThreshold is how much a holding must move today (in percent) before
// it shows up in the "Today's Movers" alert list.
const moverThreshold = 3.0

// moversCap is the max number of movers shown before collapsing into "+N more".
const moversCap = 3

type mover struct {
	symbol string
	pct    float64
}

func computeMovers(holdings []models.Holding, quotes map[string]models.Quote) []mover {
	var out []mover
	for _, h := range holdings {
		q, ok := quotes[h.Symbol]
		if !ok {
			continue
		}
		if q.ChangePct >= moverThreshold || q.ChangePct <= -moverThreshold {
			out = append(out, mover{symbol: h.Symbol, pct: q.ChangePct})
		}
	}
	sort.Slice(out, func(i, j int) bool { return absf(out[i].pct) > absf(out[j].pct) })
	return out
}

func absf(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func sectionHeader(title string) string {
	return lipgloss.NewStyle().Foreground(theme.ColorFoam).Bold(true).Render(title)
}

// sentimentMeterView renders a small labeled gauge from -1 (bearish) to +1
// (bullish) derived from locally-scored news headlines.
func sentimentMeterView(s data.SentimentSummary) string {
	color := theme.ColorMuted
	icon := "◆"
	switch s.Label {
	case data.SentimentBullish:
		color = theme.ColorGreen
		icon = "▲"
	case data.SentimentBearish:
		color = theme.ColorRed
		icon = "▼"
	}

	labelLine := lipgloss.NewStyle().Foreground(color).Bold(true).Render(icon + " " + string(s.Label))

	const width = 16
	pos := int((s.Score + 1) / 2 * float64(width))
	pos = clampInt(pos, 0, width)

	var bar strings.Builder
	for i := 0; i < width; i++ {
		if i == pos {
			bar.WriteString(lipgloss.NewStyle().Foreground(color).Bold(true).Render("●"))
		} else {
			bar.WriteString(lipgloss.NewStyle().Foreground(theme.ColorBorder).Render("─"))
		}
	}

	detail := lipgloss.NewStyle().Foreground(theme.ColorMuted).
		Render(fmt.Sprintf("%d bullish · %d bearish · %d articles", s.Positive, s.Negative, s.Total))

	if s.Total == 0 {
		detail = lipgloss.NewStyle().Foreground(theme.ColorMuted).Italic(true).Render("Not enough news yet")
	}

	return lipgloss.JoinVertical(lipgloss.Left, labelLine, bar.String(), detail)
}

func moversView(movers []mover) string {
	if len(movers) == 0 {
		return lipgloss.NewStyle().Foreground(theme.ColorMuted).Italic(true).Render("No notable moves today.")
	}

	var rows []string
	for i, mv := range movers {
		if i >= moversCap {
			rows = append(rows, lipgloss.NewStyle().Foreground(theme.ColorMuted).Italic(true).
				Render(fmt.Sprintf("  +%d more", len(movers)-moversCap)))
			break
		}
		style := theme.PositiveChange
		arrow := "▲"
		if mv.pct < 0 {
			style = theme.NegativeChange
			arrow = "▼"
		}
		sym := lipgloss.NewStyle().Foreground(theme.ColorText).Render(fmt.Sprintf("%-8s", truncate(mv.symbol, 8)))
		rows = append(rows, style.Render(arrow)+" "+sym+style.Render(fmt.Sprintf("%+.2f%%", mv.pct)))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func timeAgo(unix int64) string {
	t := time.Unix(unix, 0)
	diff := time.Since(t)
	if diff < 0 {
		diff = 0
	}
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}

func newsListView(news []models.NewsItem, scrollIdx, maxVisible int, width int) string {
	if len(news) == 0 {
		return lipgloss.NewStyle().Foreground(theme.ColorMuted).Italic(true).Render("No recent market news.")
	}

	start := clampInt(scrollIdx, 0, max(len(news)-1, 0))
	end := min(start+maxVisible, len(news))

	var rows []string
	for _, item := range news[start:end] {
		headline := lipgloss.NewStyle().Foreground(theme.ColorText).Render(truncate(item.Headline, max(width, 20)))
		meta := lipgloss.NewStyle().Foreground(theme.ColorMuted).
			Render(fmt.Sprintf("  %s · %s · %s", item.Related, item.Source, timeAgo(item.Datetime)))
		rows = append(rows, headline+"\n"+meta)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	if len(news) > maxVisible {
		pager := lipgloss.NewStyle().Foreground(theme.ColorMuted).Italic(true).
			Render(fmt.Sprintf("  %d-%d of %d · J/K scroll", start+1, end, len(news)))
		body = lipgloss.JoinVertical(lipgloss.Left, body, pager)
	}
	return body
}

// insightsView renders the right-hand "cute window": sentiment, today's
// movers, and scrolling news, all in one purple-bordered card.
func (m Model) insightsView(width, height int) string {
	if width < 4 {
		width = 4
	}

	title := lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true).Render("MARKET PULSE")
	divider := lipgloss.NewStyle().Foreground(theme.ColorBorder).Render(strings.Repeat("─", max(width-4, 4)))

	sentimentBody := sentimentMeterView(m.sentiment)
	movers := computeMovers(m.Holdings, m.quotes)
	alertsBody := moversView(movers)

	var newsBody string
	switch {
	case len(m.Holdings) == 0:
		newsBody = lipgloss.NewStyle().Foreground(theme.ColorMuted).Italic(true).Render("Add a position to see news.")
	case m.newsLoading:
		newsBody = lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("Loading news…")
	case m.newsErr != "":
		newsBody = lipgloss.NewStyle().Foreground(theme.ColorRed).Render("Error: " + m.newsErr)
	default:
		maxVisible := clampInt(height/6, 2, 6)
		newsBody = newsListView(m.news, m.newsScroll, maxVisible, width-4)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		divider,
		"",
		sectionHeader("SENTIMENT"),
		sentimentBody,
		"",
		sectionHeader("TODAY'S MOVERS"),
		alertsBody,
		"",
		sectionHeader("NEWS"),
		newsBody,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Padding(1, 2).
		Width(width).
		Height(height).
		Render(content)
}
