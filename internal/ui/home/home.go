package home

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

const banner = `
███╗   ███╗ █████╗ ███╗   ██╗███████╗████████╗ █████╗
████╗ ████║██╔══██╗████╗  ██║██╔════╝╚══██╔══╝██╔══██╗
██╔████╔██║██║  ██║██╔██╗ ██║███████╗   ██║   ███████║
██║╚██╔╝██║██║  ██║██║╚██╗██║╚════██║   ██║   ██╔══██║
██║ ╚═╝ ██║╚█████╔╝██║ ╚████║███████║   ██║   ██║  ██║
╚═╝     ╚═╝ ╚════╝ ╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝
`

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type Model struct {
	list      list.Model
	quitting  bool
	Width     int
	Height    int
}

func New() Model {
	items := []list.Item{
		item{title: "View Portfolio", desc: "Browse your holdings and charts"},
		item{title: "News", desc: "Live Finnhub news for your holdings"},
		item{title: "Quit", desc: "Exit MONETA"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return Model{list: l}
}

func (m Model) Selected() string {
	return m.list.SelectedItem().(item).title
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		helpHeight := 1
		footerHeight := 1
		topMargin := 3
		bannerHeight := 7
		listHeight := m.Height - topMargin - bannerHeight - helpHeight - footerHeight - 2
		if listHeight < 4 {
			listHeight = 4
		}
		m.list.SetSize(msg.Width-4, listHeight)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	bannerStyle := lipgloss.NewStyle().
		Foreground(theme.ColorPurple).
		Bold(true)

	helpText := lipgloss.NewStyle().
		Foreground(theme.ColorMuted).
		Render("  \u2191\u2193 navigate  \u21B5 select  q quit")

	bannerView := bannerStyle.Render(banner)
	listView := m.list.View()

	content := lipgloss.JoinVertical(lipgloss.Center,
		bannerView,
		"",
		listView,
		"",
		helpText,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Padding(1, 2).
		Width(m.Width - 4)

	return box.Render(content)
}
