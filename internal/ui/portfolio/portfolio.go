package portfolio

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/data"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/chart"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

type mode int

const (
	modeBrowse mode = iota
	modeAdd
	modeRemove
)

type Model struct {
	Holdings    []models.Holding
	Chart       chart.Model
	provider    *data.Yahoo
	mode        mode
	selectedIdx int

	// Add form fields
	addSymbol string
	addShares string
	addPrice  string
	addStep   int

	// Remove confirmation
	removeConfirm string

	Width  int
	Height int
}

func New() Model {
	return Model{
		Chart:    chart.New(),
		provider: data.NewYahoo(),
	}
}

func (m Model) InForm() bool {
	return m.mode != modeBrowse
}

func (m Model) Init() tea.Cmd {
	return m.fetchQuotes()
}

func (m Model) fetchQuotes() tea.Cmd {
	symbols := make([]string, len(m.Holdings))
	for i, h := range m.Holdings {
		symbols[i] = h.Symbol
	}
	if len(symbols) == 0 {
		return nil
	}
	return func() tea.Msg {
		quotes, err := m.provider.GetQuotes(symbols)
		return quotesMsg{quotes: quotes, err: err}
	}
}

func (m Model) fetchHistory(symbol string) tea.Cmd {
	return func() tea.Msg {
		candles, err := m.provider.GetHistory(symbol, models.Range24H)
		return historyMsg{symbol: symbol, data: candles, err: err}
	}
}

type quotesMsg struct {
	quotes []models.Quote
	err    error
}

type historyMsg struct {
	symbol string
	data   []models.Candle
	err    error
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Chart.SetSize(msg.Width-msg.Width/3-4, msg.Height-4)

	case tea.KeyMsg:
		if m.mode == modeAdd {
			return m.updateAddForm(msg)
		}
		if m.mode == modeRemove {
			return m.updateRemoveConfirm(msg)
		}

		switch msg.String() {
		case "escape", "esc":
			return m, nil // parent handles this

		case "q":
			return m, nil

		case "a":
			m.mode = modeAdd
			m.addSymbol = ""
			m.addShares = ""
			m.addPrice = ""
			m.addStep = 0
			return m, nil

		case "d":
			if len(m.Holdings) > 0 {
				m.mode = modeRemove
				m.removeConfirm = ""
				m.selectedIdx = 0
			}
			return m, nil

		case "j", "down":
			if m.selectedIdx < len(m.Holdings)-1 {
				m.selectedIdx++
			}
			return m, m.fetchHistory(m.Holdings[m.selectedIdx].Symbol)

		case "k", "up":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
			return m, m.fetchHistory(m.Holdings[m.selectedIdx].Symbol)

		case "tab":
			m.Chart.CycleChartType()
			return m, nil
		}

	case quotesMsg:
		if msg.err == nil {
			m.applyQuotes(msg.quotes)
		}

	case historyMsg:
		if msg.err == nil && msg.data != nil {
			// Store in chart
		}
	}

	// Route chart updates
	var cmd tea.Cmd
	m.Chart, cmd = m.Chart.Update(msg)
	return m, cmd
}

func (m Model) updateAddForm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc":
		m.mode = modeBrowse
		return m, nil
	case "enter":
		switch m.addStep {
		case 0:
			if m.addSymbol != "" {
				m.addStep = 1
			}
		case 1:
			if m.addShares != "" {
				m.addStep = 2
			}
		case 2:
			if m.addPrice != "" {
				shares, _ := strconv.ParseFloat(m.addShares, 64)
				price, _ := strconv.ParseFloat(m.addPrice, 64)
				m.Holdings = append(m.Holdings, models.Holding{
					Symbol:   strings.ToUpper(m.addSymbol),
					Shares:   shares,
					AvgPrice: price,
				})
				m.savePortfolio()
				m.mode = modeBrowse
			}
		}
		return m, nil
	case "backspace":
		switch m.addStep {
		case 0:
			if len(m.addSymbol) > 0 {
				m.addSymbol = m.addSymbol[:len(m.addSymbol)-1]
			}
		case 1:
			if len(m.addShares) > 0 {
				m.addShares = m.addShares[:len(m.addShares)-1]
			}
		case 2:
			if len(m.addPrice) > 0 {
				m.addPrice = m.addPrice[:len(m.addPrice)-1]
			}
		}
		return m, nil
	default:
		ch := msg.String()
		if len(ch) == 1 {
			switch m.addStep {
			case 0:
				if (ch >= "a" && ch <= "z") || (ch >= "A" && ch <= "Z") || ch == "." {
					m.addSymbol += strings.ToUpper(ch)
				}
			case 1:
				if ch >= "0" && ch <= "9" || ch == "." {
					m.addShares += ch
				}
			case 2:
				if ch >= "0" && ch <= "9" || ch == "." {
					m.addPrice += ch
				}
			}
		}
		return m, nil
	}
}

func (m Model) updateRemoveConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc":
		m.mode = modeBrowse
		return m, nil
	case "enter":
		if m.selectedIdx >= 0 && m.selectedIdx < len(m.Holdings) {
			m.Holdings = append(m.Holdings[:m.selectedIdx], m.Holdings[m.selectedIdx+1:]...)
			m.savePortfolio()
		}
		m.mode = modeBrowse
		return m, nil
	case "y":
		if m.selectedIdx >= 0 && m.selectedIdx < len(m.Holdings) {
			m.Holdings = append(m.Holdings[:m.selectedIdx], m.Holdings[m.selectedIdx+1:]...)
			m.savePortfolio()
		}
		m.mode = modeBrowse
		return m, nil
	case "n":
		m.mode = modeBrowse
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) applyQuotes(quotes []models.Quote) {
	// TODO: store quotes for live price display
}

func (m *Model) savePortfolio() {
	p := &models.Portfolio{Holdings: m.Holdings}
	data.SavePortfolio(data.PortfolioPath, p)
}

func (m Model) View() string {
	if m.mode == modeAdd {
		return m.addFormView()
	}
	if m.mode == modeRemove {
		return m.removeConfirmView()
	}
	return m.mainView()
}

func (m Model) mainView() string {
	if len(m.Holdings) == 0 {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
			"No holdings. Press 'a' to add one.",
			lipgloss.WithWhitespaceForeground(theme.ColorMuted))
	}

	total := 0.0
	for _, h := range m.Holdings {
		total += h.Shares * h.AvgPrice
	}

	var rows []string
	colors := []lipgloss.Color{
		theme.ColorPurple, theme.ColorFoam, theme.ColorYellow, theme.ColorRed,
		theme.ColorGreen, theme.ColorText, theme.ColorMuted,
	}

	for i, h := range m.Holdings {
		pct := (h.Shares * h.AvgPrice) / total * 100
		barColor := colors[i%len(colors)]

		barLen := int(pct / 5)
		if barLen < 1 {
			barLen = 1
		}
		if barLen > 20 {
			barLen = 20
		}

		barChars := ""
		for j := 0; j < barLen; j++ {
			barChars += "█"
		}

		prefix := "  "
		if i == m.selectedIdx {
			prefix = "\u25b6 "
		}

		bar := lipgloss.NewStyle().Foreground(barColor).Render(barChars)
		symbol := lipgloss.NewStyle().Foreground(theme.ColorText).Render("  " + h.Symbol + "  ")
		pctStr := lipgloss.NewStyle().Foreground(theme.ColorText).Render(fmtPct(pct))

		row := fmt.Sprintf("%s%s%s%s", prefix, bar, symbol, pctStr)
		rows = append(rows, row)
	}

	sidebar := lipgloss.JoinVertical(lipgloss.Left, rows...)
	sidebar = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Width(m.Width/3).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true).Render(" Positions "),
			"",
			sidebar,
		))

	chartView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Width(m.Width - m.Width/3 - 2).
		Height(m.Height - 4).
		Render(lipgloss.Place(
			m.Width-m.Width/3-4, m.Height-6,
			lipgloss.Center, lipgloss.Center,
			"Select a ticker to view chart",
			lipgloss.WithWhitespaceForeground(theme.ColorMuted),
		))

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, sidebar, chartView),
		lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("  a add  d remove  \u2191\u2193 select  Tab chart  Esc back"),
	)
}

func (m Model) addFormView() string {
	var fields []string
	switch m.addStep {
	case 0:
		fields = []string{
			lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true).Render("Add Position"),
			"",
			lipgloss.NewStyle().Foreground(theme.ColorText).Render("Symbol: " + m.addSymbol + "\u258c"),
			"",
			lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("Type ticker symbol and press Enter"),
		}
	case 1:
		fields = []string{
			lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true).Render("Add Position"),
			"",
			lipgloss.NewStyle().Foreground(theme.ColorText).Render("Symbol: " + m.addSymbol),
			lipgloss.NewStyle().Foreground(theme.ColorText).Render("Shares: " + m.addShares + "\u258c"),
			"",
			lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("Enter number of shares"),
		}
	case 2:
		fields = []string{
			lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true).Render("Add Position"),
			"",
			lipgloss.NewStyle().Foreground(theme.ColorText).Render("Symbol: " + m.addSymbol),
			lipgloss.NewStyle().Foreground(theme.ColorText).Render("Shares: " + m.addShares),
			lipgloss.NewStyle().Foreground(theme.ColorText).Render("Avg Price: $" + m.addPrice + "\u258c"),
			"",
			lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("Enter average price, then Enter to confirm"),
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, fields...)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorBorder).
			Padding(1, 3).
			Render(content),
		lipgloss.WithWhitespaceForeground(theme.ColorBg),
	)
}

func (m Model) removeConfirmView() string {
	sym := ""
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.Holdings) {
		sym = m.Holdings[m.selectedIdx].Symbol
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.ColorRed).Bold(true).Render("Remove Position"),
		"",
		lipgloss.NewStyle().Foreground(theme.ColorText).Render("Remove "+sym+" from portfolio?"),
		"",
		lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("  y (yes)  n (no)  Esc cancel"),
	)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorRed).
			Padding(1, 3).
			Render(content),
		lipgloss.WithWhitespaceForeground(theme.ColorBg),
	)
}

func fmtPct(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}
