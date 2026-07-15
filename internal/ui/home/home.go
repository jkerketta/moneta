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

// blossomRows returns decorative cherry blossom petal lines scattered above
// and below the MONETA banner, styled in the Rose Pine Moon palette.
func blossomRows() (top, bot string) {
	red := theme.Blossom.Render("✿")
	pur := theme.BlossomL.Render("❀")
	teal := theme.Petal.Render("❁")
	gold := theme.PetalG.Render("✾")
	dust := theme.Particle.Render("·")

	sp := func(n int) string { return strings.Repeat(" ", n) }

	top = strings.Join([]string{
		sp(2) + red + sp(10) + dust + sp(12) + pur + sp(20) + dust + sp(10) + teal + sp(3) + red,
		sp(6) + dust + sp(11) + red + sp(13) + red + sp(21) + pur + sp(10) + dust + sp(2),
		sp(2) + pur + sp(12) + dust + sp(10) + gold + sp(12) + dust + sp(9) + red + sp(12) + teal,
	}, "\n")

	bot = strings.Join([]string{
		sp(4) + red + sp(9) + dust + sp(10) + teal + sp(9) + gold + sp(9) + dust + sp(10) + pur,
		sp(2) + dust + sp(10) + pur + sp(12) + dust + sp(12) + red + sp(13) + gold + sp(2),
		sp(6) + teal + sp(14) + red + sp(9) + pur + sp(9) + dust + sp(9) + dust + sp(8) + red,
	}, "\n")

	return
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

	petalTop, petalBot := blossomRows()

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
		petalTop,
		"",
		bannerStyle.Render(banner),
		"",
		petalBot,
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
