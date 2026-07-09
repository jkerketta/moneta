package alerts

import (
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/data"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

type Model struct {
	news       []models.NewsItem
	symbols    []string
	loading    bool
	err        string
	scrollIdx  int
	Width, Height int
}

type newsLoadedMsg struct {
	news []models.NewsItem
	err  error
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return m.loadNews()
}

func (m Model) loadNews() tea.Cmd {
	return func() tea.Msg {
		news, err := data.FetchAllNews(m.symbols)
		return newsLoadedMsg{news: news, err: err}
	}
}

func (m Model) SetSymbols(symbols []string) Model {
	m.symbols = symbols
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case newsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.news = msg.news
			m.err = ""
			// Sort newest first
			sort.Slice(m.news, func(i, j int) bool {
				return m.news[i].Datetime > m.news[j].Datetime
			})
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.scrollIdx < len(m.news)-1 {
				m.scrollIdx++
			}
		case "k", "up":
			if m.scrollIdx > 0 {
				m.scrollIdx--
			}
		case "r":
			m.loading = true
			m.err = ""
			return m, m.loadNews()
		}
	}
	return m, nil
}

func timeAgo(unix int64) string {
	t := time.Unix(unix, 0)
	diff := time.Since(t)
	if diff < 0 {
		diff = 0
	}
	if diff < 1*time.Minute {
		return "just now"
	}
	if diff < 1*time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
}

func (m Model) View() string {
	if m.Width < 20 {
		return ""
	}

	header := lipgloss.NewStyle().
		Foreground(theme.ColorPurple).
		Bold(true).
		Render("FINNHUB NEWS")

	if len(m.symbols) == 0 {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorBorder).
			Padding(1, 2).
			Width(m.Width - 4).
			Height(m.Height - 2).
			Render(lipgloss.Place(m.Width-8, m.Height-6, lipgloss.Center, lipgloss.Center,
				"No holdings configured. Add positions first.",
				lipgloss.WithWhitespaceForeground(theme.ColorMuted)))
	}

	if m.loading {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorBorder).
			Padding(1, 2).
			Width(m.Width - 4).
			Height(m.Height - 2).
			Render(lipgloss.Place(m.Width-8, m.Height-6, lipgloss.Center, lipgloss.Center,
				"Loading news...",
				lipgloss.WithWhitespaceForeground(theme.ColorMuted)))
	}

	if m.err != "" {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorRed).
			Padding(1, 2).
			Width(m.Width - 4).
			Height(m.Height - 2).
			Render(lipgloss.Place(m.Width-8, m.Height-6, lipgloss.Center, lipgloss.Center,
				fmt.Sprintf("Error: %s\n\nCheck FINNHUB_API_KEY is set.", m.err),
				lipgloss.WithWhitespaceForeground(theme.ColorMuted)))
	}

	if len(m.news) == 0 {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorBorder).
			Padding(1, 2).
			Width(m.Width - 4).
			Height(m.Height - 2).
			Render(lipgloss.Place(m.Width-8, m.Height-6, lipgloss.Center, lipgloss.Center,
				"No recent news for your holdings.",
				lipgloss.WithWhitespaceForeground(theme.ColorMuted)))
	}

	var rows []string
	maxVisible := m.Height - 6
	start := m.scrollIdx
	end := start + maxVisible
	if end > len(m.news) {
		end = len(m.news)
	}

	for _, item := range m.news[start:end] {
		ago := timeAgo(item.Datetime)
		headline := lipgloss.NewStyle().
			Foreground(theme.ColorText).
			Render(item.Headline)
		source := lipgloss.NewStyle().
			Foreground(theme.ColorMuted).
			Render(fmt.Sprintf("  [%s] %s · %s", item.Related, item.Source, ago))

		row := fmt.Sprintf("%s\n%s", headline, source)
		rows = append(rows, row)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	status := ""
	if len(m.news) > 0 {
		status = fmt.Sprintf("  r refresh · %d news items · ↑↓ scroll · Esc back", len(m.news))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Padding(1, 2).
		Width(m.Width - 4).
		Height(m.Height - 2).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			content,
			"",
			lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(status),
		))
}
