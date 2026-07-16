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
	{title: "Quit"},
}

type Model struct {
	selected int
	quitting bool
	Width    int
	Height   int
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
			m.selected = (m.selected + 1) % len(menuItems)
		case "k", "up", "shift+tab":
			m.selected = (m.selected - 1 + len(menuItems)) % len(menuItems)
		case "enter":
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

const canvas = 90

func petalLine(cols ...int) string {
	var b strings.Builder
	pos := 0
	for _, col := range cols {
		for pos < col {
			b.WriteByte(' ')
			pos++
		}
		b.WriteString("✿")
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
	petalStyle := lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true)

	raw := strings.Split(banner, "\n")
	lines := make([]string, 0, 6)
	for _, l := range raw {
		if l != "" {
			lines = append(lines, l)
		}
	}

	// pad wraps a banner line with a 2-char left prefix and 2-char right
	// suffix. Either side gets a petal when its bool is true.
	pad := func(s string, left, right bool) string {
		l := "  "
		if left {
			l = petalStyle.Render("✿") + " "
		}
		r := "  "
		if right {
			r = " " + petalStyle.Render("✿")
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

	var stack []string

	stack = append(stack, centerLine(petalStyle.Render(petalLine(35, 69))))
	stack = append(stack, centerLine(petalStyle.Render(petalLine(24, 57))))

	stack = append(stack, centerLine(pad(lines[0], false, false)))
	stack = append(stack, centerLine(pad(lines[1], true, false)))
	stack = append(stack, centerLine(pad(lines[2], false, false)))
	stack = append(stack, centerLine(pad(lines[3], false, true)))
	stack = append(stack, centerLine(pad(lines[4], false, false)))
	stack = append(stack, centerLine(pad(lines[5], false, false)))

	stack = append(stack, centerLine(petalStyle.Render(petalLine(3, 36, 67))))
	stack = append(stack, centerLine(petalStyle.Render(petalLine(18, 52))))

	return lipgloss.JoinVertical(lipgloss.Left, stack...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	bannerStyle := lipgloss.NewStyle().
		Foreground(theme.ColorPurple).
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
			titleStyle = lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true)
		}

		rows = append(rows, marker+titleStyle.Render(it.title))
	}

	menu := lipgloss.JoinVertical(lipgloss.Center, rows...)

	helpText := lipgloss.NewStyle().
		Foreground(theme.ColorMuted).
		Render("↑↓/jk navigate  ↵ select  q quit")

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
