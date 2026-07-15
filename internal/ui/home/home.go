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

// bannerWidth is the display width in columns of the MONETA ASCII banner.
const bannerWidth = 70

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
	for pos < bannerWidth {
		b.WriteByte(' ')
		pos++
	}
	return b.String()
}

func composePetals(bannerStyle lipgloss.Style) string {
	petalStyle := lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true)

	raw := strings.Split(banner, "\n")
	lines := make([]string, 0, 6)
	for _, l := range raw {
		if l != "" {
			lines = append(lines, l)
		}
	}

	var stack []string

	stack = append(stack, petalStyle.Render(petalLine(6, 63)))
	stack = append(stack, petalStyle.Render(petalLine(42)))
	stack = append(stack, "")

	stack = append(stack, bannerStyle.Render(lines[0]))
	stack = append(stack, bannerStyle.Render(lines[1]))

	stack = append(stack, "")
	stack = append(stack, petalStyle.Render(petalLine(2, 66)))
	stack = append(stack, "")

	stack = append(stack, bannerStyle.Render(lines[2]))
	stack = append(stack, bannerStyle.Render(lines[3]))

	stack = append(stack, "")
	stack = append(stack, petalStyle.Render(petalLine(22, 50)))
	stack = append(stack, "")

	stack = append(stack, bannerStyle.Render(lines[4]))
	stack = append(stack, bannerStyle.Render(lines[5]))

	stack = append(stack, "")
	stack = append(stack, petalStyle.Render(petalLine(4, 62)))
	stack = append(stack, petalStyle.Render(petalLine(32)))

	return lipgloss.JoinVertical(lipgloss.Center, stack...)
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

	bannerWithPetals := composePetals(bannerStyle)

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
		bannerWithPetals,
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
