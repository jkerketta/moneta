package portfolio

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/chart"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

type Model struct {
	Holdings []models.Holding
	Chart    chart.Model
	err      error
	Width    int
	Height   int
}

func New() Model {
	return Model{
		Chart: chart.New(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	if len(m.Holdings) == 0 {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
			"No holdings. Press 'a' to add one.",
			lipgloss.WithWhitespaceForeground(theme.ColorMuted))
	}

	total := 0.0
	for _, h := range m.Holdings {
		total += h.Shares * h.AvgPrice
	}

	var rows []string
	colors := []string{
		"#c4a7e7", "#9ccfd8", "#f6c177", "#eb6f92",
		"#31748f", "#e0def4", "#908caa",
	}

	for i, h := range m.Holdings {
		pct := (h.Shares * h.AvgPrice) / total * 100
		barColor := colors[i%len(colors)]

		barLen := int(pct / 5)
		if barLen < 1 {
			barLen = 1
		}
		barChars := ""
		for j := 0; j < barLen && j < 20; j++ {
			barChars += "█"
		}

		row := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render(
			barChars,
		) + lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(
			"  "+h.Symbol+"  ",
		) + lipgloss.NewStyle().Foreground(theme.ColorText).Render(
			"  "+formatPct(pct),
		)

		rows = append(rows, row)
	}

	sidebar := lipgloss.JoinVertical(lipgloss.Left, rows...)
	sidebar = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Width(m.Width / 3).
		Render(sidebar)

	chartView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Width(m.Width - m.Width/3).
		Render(lipgloss.Place(
			m.Width-m.Width/3-4, m.Height-4,
			lipgloss.Center, lipgloss.Center,
			"Select a ticker to view chart",
			lipgloss.WithWhitespaceForeground(theme.ColorMuted),
		))

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, chartView)
}

func formatPct(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}
