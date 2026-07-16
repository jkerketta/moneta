package theme

import "github.com/charmbracelet/lipgloss"

// Rose Pine Moon base palette (shared across all themes)
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

	// Aliases
	ColorSuccess = ColorGreen
	ColorError   = ColorRed
	ColorWarning = ColorYellow
	ColorSubtext = ColorMuted

	// Theme-specific shade ramps for the allocation donut
	PurpleShades = []lipgloss.Color{
		"#e6dbf7", "#c4a7e7", "#a888d4", "#8c6bc0", "#7050a0", "#54476b",
	}
	YellowShades = []lipgloss.Color{
		"#fce9c9", "#f6c177", "#e8b04c", "#d4a032", "#af8328", "#7a5c1a",
	}
	CoralShades = []lipgloss.Color{
		"#fbc5bc", "#e8736a", "#d45a50", "#be453a", "#9a322a", "#70221b",
	}
	GreenShades = []lipgloss.Color{
		"#b8dfc5", "#4a9c5d", "#3d854f", "#2f6e40", "#235732", "#164024",
	}
	IceShades = []lipgloss.Color{
		"#d2eff5", "#7ec8e3", "#5aafca", "#4097b3", "#2f7d97", "#1f6178",
	}
)

// Theme holds a complete set of accent colors, shade ramps, and decoration
// icons for the MONETA UI. Switching themes updates the accent color used
// throughout the app and the decorative characters around the title.
type Theme struct {
	Name          string
	Accent        lipgloss.Color
	Shades        []lipgloss.Color
	Primary       string // main decorative icon
	Secondary     string // lighter variant
	AccentIcon    string // tertiary icon
	Particle      string // background dust
}

var Themes = []Theme{
	{
		Name:       "Rose Pine Moon",
		Accent:     ColorPurple,
		Shades:     PurpleShades,
		Primary:    "✿",
		Secondary:  "❀",
		AccentIcon: "✾",
		Particle:   "·",
	},
	{
		Name:       "Golden Hour",
		Accent:     ColorYellow,
		Shades:     YellowShades,
		Primary:    "★",
		Secondary:  "☆",
		AccentIcon: "☽",
		Particle:   "✦",
	},
	{
		Name:       "Coral Reef",
		Accent:     CoralShades[1],
		Shades:     CoralShades,
		Primary:    "❂",
		Secondary:  "✧",
		AccentIcon: "▹",
		Particle:   "·",
	},
	{
		Name:       "Emerald Forest",
		Accent:     GreenShades[1],
		Shades:     GreenShades,
		Primary:    "☘",
		Secondary:  "◇",
		AccentIcon: "✿",
		Particle:   "·",
	},
	{
		Name:       "Midnight Ice",
		Accent:     IceShades[1],
		Shades:     IceShades,
		Primary:    "❄",
		Secondary:  "✧",
		AccentIcon: "◇",
		Particle:   "·",
	},
}

var (
	currentIdx = 0
	savedIdx   = 0
)

func CurrentIdx() int   { return currentIdx }
func Current() *Theme   { return &Themes[currentIdx] }

// Preview sets the current theme to index i without confirming it.
// The previously confirmed theme is remembered for Revert.
func Preview(i int) {
	savedIdx = currentIdx
	if i >= 0 && i < len(Themes) {
		currentIdx = i
	}
}

// Confirm commits the current previewed theme as the permanent choice.
func Confirm() {
	savedIdx = currentIdx
}

// Revert restores the theme that was active before the last Preview call.
func Revert() {
	currentIdx = savedIdx
}

// Base
var (
	Base = lipgloss.NewStyle().Foreground(ColorText)

	// Pane
	Pane = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)

	// List item
	ListItem = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	PositiveChange = lipgloss.NewStyle().Foreground(ColorGreen)
	NegativeChange = lipgloss.NewStyle().Foreground(ColorRed)

	// Chart
	ChartLabel = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(8).
			Align(lipgloss.Right)
)

// ActivePane returns a pane style using the current theme accent for the border.
func ActivePane() lipgloss.Style {
	return Pane.Copy().BorderForeground(Current().Accent)
}

// SelectedItem returns a list-item style using the current theme accent.
func SelectedItem() lipgloss.Style {
	return ListItem.Copy().
		Background(ColorSelection).
		Foreground(Current().Accent).
		Bold(true)
}

// ShadeFor returns a deterministic shade for slice index i out of n,
// drawn from the current theme's shade ramp.
func ShadeFor(i, n int) lipgloss.Color {
	s := Current().Shades
	if len(s) == 0 {
		return Current().Accent
	}
	if n <= 1 {
		return s[len(s)/2]
	}
	pos := i * (len(s) - 1) / max(1, n-1)
	if pos >= len(s) {
		pos = len(s) - 1
	}
	return s[pos]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
