package theme

import "github.com/charmbracelet/lipgloss"

// Rose Pine Moon palette
var (
	ColorBg         = lipgloss.Color("#232136")
	ColorText       = lipgloss.Color("#e0def4")
	ColorMuted      = lipgloss.Color("#908caa")
	ColorPurple     = lipgloss.Color("#c4a7e7")
	ColorYellow     = lipgloss.Color("#f6c177")
	ColorFoam       = lipgloss.Color("#9ccfd8")
	ColorGreen      = lipgloss.Color("#31748f")
	ColorRed        = lipgloss.Color("#eb6f92")
	ColorSelection  = lipgloss.Color("#2a273f")
	ColorBorder     = lipgloss.Color("#393552")

	// Base
	Base = lipgloss.NewStyle().Foreground(ColorText)

	// Pane
	Pane = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)

	ActivePane = Pane.Copy().
			BorderForeground(ColorPurple)

	// List item
	ListItem = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

	SelectedItem = ListItem.Copy().
			Background(ColorSelection).
			Foreground(ColorPurple).
			Bold(true)

	PositiveChange = lipgloss.NewStyle().Foreground(ColorGreen)
	NegativeChange = lipgloss.NewStyle().Foreground(ColorRed)

	// Chart
	ChartLabel = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(8).
			Align(lipgloss.Right)
)
