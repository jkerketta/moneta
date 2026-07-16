package home

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

const banner = `
███╗   ███╗ █████╗ ███╗   ██╗███████╗████████╗ █████╗
████╗ ████║██╔══██╗████╗  ██║██╔════╝╚══██╔══╝██╔══██╗
██╔████╔██║██║  ██║██╔██╗ ██║█████╗     ██║   ███████║
██║╚██╔╝██║██║  ██║██║╚██╗██║██╔══╝     ██║   ██╔══██║
██║ ╚═╝ ██║╚█████╔╝██║ ╚████║███████╗   ██║   ██║  ██║
╚═╝     ╚═╝ ╚════╝ ╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝
`

const tagline = "Your portfolio, priced in real time"

type menuItem struct {
	title string
}

var menuItems = []menuItem{
	{title: "View Portfolio"},
	{title: "Change Theme"},
	{title: "Quit"},
}

type Model struct {
	selected       int
	quitting       bool
	selectingTheme bool
	themeIdx       int
	previewMode    bool
	Width          int
	Height         int
}

func New() Model {
	return Model{}
}

func (m Model) Selected() string {
	return menuItems[m.selected].title
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "j", "down", "tab":
			if m.selectingTheme {
				m.themeIdx = (m.themeIdx + 1) % len(theme.Themes)
				return m, nil
			}
			m.selected = (m.selected + 1) % len(menuItems)
			return m, nil

		case "k", "up", "shift+tab":
			if m.selectingTheme {
				m.themeIdx = (m.themeIdx - 1 + len(theme.Themes)) % len(theme.Themes)
				return m, nil
			}
			m.selected = (m.selected - 1 + len(menuItems)) % len(menuItems)
			return m, nil

		case "enter":
			if m.selectingTheme {
				theme.Preview(m.themeIdx)
				m.selectingTheme = false
				m.previewMode = true
				m.selected = 1
				return m, nil
			}
			if m.previewMode {
				theme.Confirm()
				m.previewMode = false
				return m, nil
			}
			if menuItems[m.selected].title == "Change Theme" {
				m.selectingTheme = true
				m.themeIdx = theme.CurrentIdx()
				return m, nil
			}
			return m, nil

		case "escape", "esc":
			if m.previewMode {
				theme.Revert()
				m.previewMode = false
				return m, nil
			}
			if m.selectingTheme {
				m.selectingTheme = false
				return m, nil
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

const canvas = 90

func petalLine(char string, cols ...int) string {
	var b strings.Builder
	pos := 0
	for _, col := range cols {
		for pos < col {
			b.WriteByte(' ')
			pos++
		}
		b.WriteString(char)
		pos++
	}
	for pos < canvas {
		b.WriteByte(' ')
		pos++
	}
	return b.String()
}

// composePetals builds the MONETA banner with petals scattered around it.
// Every line is exactly `canvas` columns wide — petal rows, banner lines
// with side petals, and bare banner lines — so the composition has straight
// edges and no centering artifacts.
func composePetals(bannerStyle lipgloss.Style) string {
	petalStyle := lipgloss.NewStyle().Foreground(theme.Current().Accent).Bold(true)

	raw := strings.Split(banner, "\n")
	lines := make([]string, 0, 6)
	for _, l := range raw {
		if l != "" {
			lines = append(lines, l)
		}
	}

	// pad wraps a banner line with a 2-char left prefix and 2-char right
	// suffix. The leftChar/rightChar, when non-empty, replace the space
	// with a styled decoration character.
	pad := func(s string, leftChar, rightChar string) string {
		l := "  "
		if leftChar != "" {
			l = petalStyle.Render(leftChar) + " "
		}
		r := "  "
		if rightChar != "" {
			r = " " + petalStyle.Render(rightChar)
		}
		return l + bannerStyle.Render(s) + r
	}

	// centerLine pads each line to exactly canvas display columns with equal
	// spacing on both sides. Every line ends up the same width and centered
	// individually, so JoinVertical(Left) produces a perfectly centered block.
	centerLine := func(s string) string {
		w := lipgloss.Width(s)
		if w >= canvas {
			return s
		}
		l := (canvas - w) / 2
		r := canvas - w - l
		return strings.Repeat(" ", l) + s + strings.Repeat(" ", r)
	}

	t := theme.Current()

	var stack []string

	stack = append(stack, centerLine(petalStyle.Render(petalLine(t.Primary, 35, 69))))
	stack = append(stack, centerLine(petalStyle.Render(petalLine(t.Secondary, 24, 57))))

	stack = append(stack, centerLine(pad(lines[0], "", "")))
	stack = append(stack, centerLine(pad(lines[1], t.Particle, "")))
	stack = append(stack, centerLine(pad(lines[2], "", "")))
	stack = append(stack, centerLine(pad(lines[3], "", t.Primary)))
	stack = append(stack, centerLine(pad(lines[4], "", "")))
	stack = append(stack, centerLine(pad(lines[5], "", "")))

	stack = append(stack, centerLine(petalStyle.Render(petalLine(t.AccentIcon, 3, 36, 67))))
	stack = append(stack, centerLine(petalStyle.Render(petalLine(t.Particle, 18, 52))))

	return lipgloss.JoinVertical(lipgloss.Left, stack...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.selectingTheme {
		return m.themeSelectorView()
	}

	bannerStyle := lipgloss.NewStyle().
		Foreground(theme.Current().Accent).
		Bold(true)

	taglineStyle := lipgloss.NewStyle().
		Foreground(theme.ColorMuted).
		Italic(true)

	bannerComposed := composePetals(bannerStyle)

	var rows []string
	for i, it := range menuItems {
		selected := i == m.selected

		marker := "  "
		titleStyle := lipgloss.NewStyle().Foreground(theme.ColorMuted)

		if selected {
			marker = "▸ "
			titleStyle = lipgloss.NewStyle().Foreground(theme.Current().Accent).Bold(true)
		}

		rows = append(rows, marker+titleStyle.Render(it.title))
	}

	menu := lipgloss.JoinVertical(lipgloss.Center, rows...)

	helpText := lipgloss.NewStyle().
		Foreground(theme.ColorMuted).Render("↑↓/jk navigate  ↵ select  q quit")
	if m.previewMode {
		helpText = lipgloss.NewStyle().
			Foreground(theme.ColorMuted).Render("  ↵ confirm  Esc revert  ↑↓ select  q quit")
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		bannerComposed,
		taglineStyle.Render(tagline),
		"",
		"",
		menu,
		"",
		"",
		helpText,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Padding(2, 6).
		Render(content)

	width, height := m.Width, m.Height
	if width <= 0 {
		width = lipgloss.Width(box)
	}
	if height <= 0 {
		height = lipgloss.Height(box)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceForeground(theme.ColorBg))
}

func (m Model) themeSelectorView() string {
	accent := theme.Current().Accent

	var themeRows []string
	for i, t := range theme.Themes {
		selected := i == m.themeIdx
		icon := t.Primary
		label := t.Name

		marker := "  "
		style := lipgloss.NewStyle().Foreground(theme.ColorMuted)
		if selected {
			marker = "▸ "
			style = lipgloss.NewStyle().Foreground(accent).Bold(true)
			icon = t.Primary
		}
		iconStyle := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(icon)
		nameStyle := style.Render(label)
		themeRows = append(themeRows, marker+iconStyle+" "+nameStyle)
	}
	themeList := lipgloss.JoinVertical(lipgloss.Center, themeRows...)

	title := lipgloss.NewStyle().Foreground(accent).Bold(true).Render("Themes")
	hint := lipgloss.NewStyle().Foreground(theme.ColorMuted).
		Render("↑↓ select  ↵ preview  Esc back")

	body := lipgloss.JoinVertical(lipgloss.Center, title, "", themeList, "", hint)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(2, 6).
		Render(body)

	width, height := m.Width, m.Height
	if width <= 0 {
		width = lipgloss.Width(box)
	}
	if height <= 0 {
		height = lipgloss.Height(box)
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceForeground(theme.ColorBg))
}
