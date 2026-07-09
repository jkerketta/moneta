package alerts

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

type alertItem struct {
	icon, iconColor, text string
}

var fakeAlerts = []alertItem{
	{icon: "!", iconColor: "#eb6f92", text: "TSLA earnings in 3 days"},
	{icon: "+", iconColor: "#31748f", text: "NVDA crossed $950 (daily alert)"},
	{icon: "+", iconColor: "#31748f", text: "NVDA up 4.3% today"},
	{icon: "+", iconColor: "#9ccfd8", text: "RKLB mentioned in Yahoo Finance news"},
	{icon: "+", iconColor: "#eb6f92", text: "CHPS.TO down 2.1% today"},
	{icon: "+", iconColor: "#908caa", text: "XEQT flat, tracking index"},
}

type Model struct {
	Width  int
	Height int
}

func New() Model { return Model{} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	var rows []string
	for _, a := range fakeAlerts {
		icon := lipgloss.NewStyle().
			Foreground(lipgloss.Color(a.iconColor)).
			SetString(" [" + a.icon + "] ")
		text := lipgloss.NewStyle().Foreground(theme.ColorText).Render(a.text)
		rows = append(rows, icon.String()+text)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Padding(1, 2).
		Width(m.Width - 4).
		Height(m.Height - 2).
		Render(lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true).Render("Alerts & News"),
			"",
			content,
			"",
			lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("  \u2191\u2193 scroll  Esc back"),
		))
}
