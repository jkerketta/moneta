package theme

import "github.com/charmbracelet/lipgloss"

// Rose Pine Moon palette
var (
	ColorBg        = lipgloss.Color("#232136")
	ColorText      = lipgloss.Color("#e0def4")
	ColorMuted     = lipgloss.Color("#908caa")
	ColorPurple    = lipgloss.Color("#c4a7e7")
	ColorYellow    = lipgloss.Color("#f6c177")
	ColorFoam      = lipgloss.Color("#9ccfd8")
	ColorGreen     = lipgloss.Color("#31748f")
	ColorRed       = lipgloss.Color("#eb6f92")
	ColorSelection = lipgloss.Color("#2a273f")
	ColorBorder    = lipgloss.Color("#393552")

	// Aliases so every component (including the former "styles" package)
	// draws from one shared palette.
	ColorSuccess = ColorGreen
	ColorError   = ColorRed
	ColorWarning = ColorYellow
	ColorSubtext = ColorMuted

	// PurpleShades is a lavender ramp, lightest to darkest, used to color
	// the portfolio allocation donut so every slice reads as "MONETA purple"
	// while staying visually distinct.
	PurpleShades = []lipgloss.Color{
		lipgloss.Color("#e6dbf7"),
		lipgloss.Color("#c4a7e7"),
		lipgloss.Color("#a888d4"),
		lipgloss.Color("#8c6bc0"),
		lipgloss.Color("#7050a0"),
		lipgloss.Color("#54476b"),
	}

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

	// Cherry blossom petal styles (Rose Pine Moon)
	Blossom  = lipgloss.NewStyle().Foreground(ColorRed)
	BlossomL = lipgloss.NewStyle().Foreground(ColorPurple)
	Petal    = lipgloss.NewStyle().Foreground(ColorFoam)
	PetalG   = lipgloss.NewStyle().Foreground(ColorYellow)
	Particle = lipgloss.NewStyle().Foreground(ColorMuted)
)

// ShadeFor returns a deterministic purple shade for slice index i out of n
// total slices, spreading indices evenly across the ramp so a portfolio
// with any number of holdings still reads as distinct shades.
func ShadeFor(i, n int) lipgloss.Color {
	if n <= 1 {
		return PurpleShades[1]
	}
	pos := i * (len(PurpleShades) - 1) / max(1, n-1)
	if pos >= len(PurpleShades) {
		pos = len(PurpleShades) - 1
	}
	return PurpleShades[pos]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
