package portfolio

import (
	"fmt"
	"math"
	"sort"
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
	modeChart
	modeNewsDetail
)

// marketSignals are the broad-market tickers whose news fills the Market
// Pulse side panel — S&P 500, Dow, NASDAQ, Crude Oil, Gold.
var marketSignals = []string{"^GSPC", "^DJI", "^IXIC", "CL=F", "GC=F"}

// Model is the consolidated portfolio hub: allocation donut + positions on
// the left, and a sentiment/alerts/news window on the right. Adding a
// position and viewing a ticker's price history both open as small
// centered overlays rather than separate screens.
type Model struct {
	Holdings    []models.Holding
	Chart       chart.Model
	provider    *data.Yahoo
	mode        mode
	selectedIdx int

	quotes        map[string]models.Quote
	quotesLoading bool

	news        []models.NewsItem
	newsLoading bool
	newsErr     string
	newsPage    int
	sentiment   data.SentimentSummary

	marketQuotes map[string]models.Quote

	tickerNews        []models.NewsItem
	tickerNewsLoading bool
	tickerNewsErr     string
	tickerNewsScroll  int

	form addForm

	chartSymbol string
	chartRange  models.TimeRange

	Width  int
	Height int
}

func New() Model {
	return Model{
		Chart:        chart.New(),
		provider:     data.NewYahoo(),
		quotes:       make(map[string]models.Quote),
		marketQuotes: make(map[string]models.Quote),
		chartRange:   models.Range24H,
	}
}

// InForm reports whether the hub is currently showing a modal/overlay
// (add form, remove confirm, or chart), so the parent app knows not to
// treat Esc as "navigate back to the home screen".
func (m Model) InForm() bool {
	return m.mode != modeBrowse
}

func (m Model) Init() tea.Cmd {
	_, cmd := m.RefreshMarketData()
	return cmd
}

// RefreshMarketData kicks off live quotes for holdings + market data for the
// side panel. Safe to call at app startup and again when entering the
// portfolio screen.
func (m Model) RefreshMarketData() (Model, tea.Cmd) {
	m.quotesLoading = len(m.Holdings) > 0
	m.newsLoading = true
	return m, tea.Batch(m.fetchQuotes(), m.fetchNews(), m.fetchMarketQuotes())
}

// HandleAsyncMsg applies quote/news/history results even when the portfolio
// screen is not active (e.g. startup fetch while still on the home screen).
// Returns handled=false for unrelated messages so the caller can keep routing.
func (m Model) HandleAsyncMsg(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg.(type) {
	case quotesMsg, newsMsg, historyMsg, tickerNewsMsg, marketQuotesMsg:
		m, cmd := m.Update(msg)
		return m, cmd, true
	default:
		return m, nil, false
	}
}

func (m Model) symbols() []string {
	out := make([]string, len(m.Holdings))
	for i, h := range m.Holdings {
		out[i] = h.Symbol
	}
	return out
}

type quotesMsg struct {
	quotes []models.Quote
	err    error
}

type newsMsg struct {
	news []models.NewsItem
	err  error
}

type marketQuotesMsg struct {
	quotes []models.Quote
	err    error
}

type tickerNewsMsg struct {
	symbol string
	news   []models.NewsItem
	err    error
}

type historyMsg struct {
	symbol string
	tr     models.TimeRange
	data   []models.Candle
	err    error
}

func (m Model) fetchQuotes() tea.Cmd {
	symbols := m.symbols()
	if len(symbols) == 0 {
		return nil
	}
	provider := m.provider
	return func() tea.Msg {
		quotes, err := provider.GetQuotes(symbols)
		return quotesMsg{quotes: quotes, err: err}
	}
}

func (m Model) fetchNews() tea.Cmd {
	provider := m.provider
	return func() tea.Msg {
		news, err := provider.FetchAllNews(marketSignals)
		return newsMsg{news: news, err: err}
	}
}

func (m Model) fetchMarketQuotes() tea.Cmd {
	provider := m.provider
	return func() tea.Msg {
		quotes, err := provider.GetQuotes(marketSignals)
		return marketQuotesMsg{quotes: quotes, err: err}
	}
}

func (m Model) fetchTickerNews(symbol string) tea.Cmd {
	provider := m.provider
	return func() tea.Msg {
		news, err := provider.FetchNews(symbol, 10)
		return tickerNewsMsg{symbol: symbol, news: news, err: err}
	}
}

func (m Model) fetchHistory(symbol string, tr models.TimeRange) tea.Cmd {
	provider := m.provider
	return func() tea.Msg {
		candles, err := provider.GetHistory(symbol, tr)
		return historyMsg{symbol: symbol, tr: tr, data: candles, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		chartW := clampInt(m.Width-8, 20, 96)
		chartH := clampInt(m.Height-8, 10, 34)
		m.Chart.SetSize(chartW, chartH)
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeAdd:
			return m.updateAddForm(msg)
		case modeRemove:
			return m.updateRemoveConfirm(msg)
		case modeChart:
			return m.updateChartOverlay(msg)
		case modeNewsDetail:
			return m.updateNewsDetail(msg)
		default:
			return m.updateBrowse(msg)
		}

	case quotesMsg:
		m.quotesLoading = false
		if msg.err == nil {
			for _, q := range msg.quotes {
				m.quotes[q.Symbol] = q
			}
		}
		return m, nil

	case marketQuotesMsg:
		if msg.err == nil {
			for _, q := range msg.quotes {
				m.marketQuotes[q.Symbol] = q
			}
		}
		return m, nil

	case newsMsg:
		m.newsLoading = false
		if msg.err != nil {
			m.newsErr = msg.err.Error()
		} else {
			m.news = msg.news
			m.newsErr = ""
			m.newsPage = 0
			sort.Slice(m.news, func(i, j int) bool { return m.news[i].Datetime > m.news[j].Datetime })
			m.sentiment = data.ScoreSentiment(m.news)
		}
		return m, nil

	case tickerNewsMsg:
		m.tickerNewsLoading = false
		if msg.err != nil {
			m.tickerNewsErr = msg.err.Error()
		} else {
			m.tickerNews = msg.news
			m.tickerNewsErr = ""
			m.tickerNewsScroll = 0
			sort.Slice(m.tickerNews, func(i, j int) bool { return m.tickerNews[i].Datetime > m.tickerNews[j].Datetime })
		}
		return m, nil

	case historyMsg:
		if m.mode == modeChart && msg.symbol == m.chartSymbol {
			if msg.err != nil {
				m.Chart.SetError(msg.err)
			} else {
				m.Chart.SetData(msg.symbol, msg.tr, msg.data)
			}
		}
		return m, nil
	}

	// Anything else (e.g. text-input cursor blink ticks) goes to whichever
	// sub-component currently owns keyboard focus.
	if m.mode == modeAdd {
		var cmd tea.Cmd
		m.form, cmd, _, _ = m.form.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.Chart, cmd = m.Chart.Update(msg)
	return m, cmd
}

func (m Model) updateBrowse(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc":
		return m, nil // parent app handles navigating back home

	case "a":
		form, cmd := newAddForm()
		m.mode = modeAdd
		m.form = form
		return m, cmd

	case "d":
		if len(m.Holdings) > 0 {
			m.mode = modeRemove
		}
		return m, nil

	case "j", "down":
		if m.selectedIdx < len(m.Holdings)-1 {
			m.selectedIdx++
		}
		return m, nil

	case "k", "up":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil

	case "left":
		if m.newsPage > 0 {
			m.newsPage--
		}
		return m, nil

	case "right":
		m.newsPage++
		return m, nil

	case "enter":
		if len(m.Holdings) == 0 {
			return m, nil
		}
		sym := m.Holdings[m.selectedIdx].Symbol
		m.mode = modeChart
		m.chartSymbol = sym
		m.chartRange = models.Range24H
		m.Chart.SetLoading(true)
		return m, m.fetchHistory(sym, m.chartRange)

	case "n":
		if len(m.Holdings) == 0 {
			return m, nil
		}
		sym := m.Holdings[m.selectedIdx].Symbol
		m.mode = modeNewsDetail
		m.tickerNewsLoading = true
		m.tickerNews = nil
		m.tickerNewsErr = ""
		m.tickerNewsScroll = 0
		return m, m.fetchTickerNews(sym)

	case "r":
		return m.RefreshMarketData()
	}
	return m, nil
}

func (m Model) updateAddForm(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var submitted, cancelled bool
	m.form, cmd, submitted, cancelled = m.form.Update(msg)

	if cancelled {
		m.mode = modeBrowse
		return m, nil
	}
	if submitted {
		h, ok := m.form.validate()
		if ok {
			m.Holdings = append(m.Holdings, h)
			m.savePortfolio()
			m.mode = modeBrowse
			return m.RefreshMarketData()
		}
	}
	return m, cmd
}

func (m Model) updateRemoveConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc", "n":
		m.mode = modeBrowse
		return m, nil
	case "enter", "y":
		if m.selectedIdx >= 0 && m.selectedIdx < len(m.Holdings) {
			removed := m.Holdings[m.selectedIdx].Symbol
			m.Holdings = append(m.Holdings[:m.selectedIdx], m.Holdings[m.selectedIdx+1:]...)
			delete(m.quotes, removed)
			if m.selectedIdx >= len(m.Holdings) && m.selectedIdx > 0 {
				m.selectedIdx--
			}
			m.savePortfolio()
		}
		m.mode = modeBrowse
		return m, nil
	}
	return m, nil
}

func (m Model) updateChartOverlay(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc", "enter", "q":
		m.mode = modeBrowse
		return m, nil
	case "tab":
		m.Chart.CycleChartType()
		return m, nil
	case "1":
		m.chartRange = models.Range1H
	case "2":
		m.chartRange = models.Range24H
	case "3":
		m.chartRange = models.Range7D
	case "4":
		m.chartRange = models.Range30D
	default:
		return m, nil
	}
	m.Chart.SetLoading(true)
	return m, m.fetchHistory(m.chartSymbol, m.chartRange)
}

func (m Model) updateNewsDetail(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc", "n":
		m.mode = modeBrowse
		return m, nil
	case "j", "J", "down":
		if m.tickerNewsScroll < len(m.tickerNews)-1 {
			m.tickerNewsScroll++
		}
		return m, nil
	case "k", "K", "up":
		if m.tickerNewsScroll > 0 {
			m.tickerNewsScroll--
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) savePortfolio() {
	p := &models.Portfolio{Holdings: m.Holdings}
	data.SavePortfolio(data.PortfolioPath, p)
}

// positionRow is one holding's derived, display-ready metrics.
type positionRow struct {
	symbol   string
	price    float64
	hasQuote bool
	value    float64
	weight   float64
	plPct    float64
	hasPL    bool
	currency string
}

func (m Model) computePositions() (rows []positionRow, totalValue float64) {
	for _, h := range m.Holdings {
		price := h.AvgPrice
		hasQuote := false
		if q, ok := m.quotes[h.Symbol]; ok && q.Price > 0 {
			price = q.Price
			hasQuote = true
		}

		value := h.Shares * price
		plPct := 0.0
		hasPL := false
		if hasQuote && h.AvgPrice > 0 {
			plPct = (price - h.AvgPrice) / h.AvgPrice * 100
			hasPL = true
		}

		currency := h.Currency
		if currency == "" {
			currency = detectCurrency(h.Symbol)
		}

		rows = append(rows, positionRow{
			symbol:   h.Symbol,
			price:    price,
			hasQuote: hasQuote,
			value:    value,
			plPct:    plPct,
			hasPL:    hasPL,
			currency: currency,
		})
		totalValue += value
	}

	for i := range rows {
		if totalValue > 0 {
			rows[i].weight = rows[i].value / totalValue * 100
		}
	}
	return rows, totalValue
}

func renderPortfolioHeader(totalValue float64, currency string) string {
	label := lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("PORTFOLIO VALUE  ")
	value := lipgloss.NewStyle().Foreground(theme.ColorText).Bold(true).Render(formatMoney(totalValue, currency))
	return label + value
}

func (m Model) positionsTableView(rows []positionRow) string {
	if len(rows) == 0 {
		return ""
	}

	header := lipgloss.NewStyle().Foreground(theme.ColorMuted).
		Render(fmt.Sprintf("  %-8s %12s %14s %8s %9s", "SYMBOL", "PRICE", "VALUE", "WEIGHT", "P/L"))

	lines := []string{header}
	for i, row := range rows {
		marker := "  "
		symStyle := lipgloss.NewStyle().Foreground(theme.ColorText)
		if i == m.selectedIdx {
			marker = "▸ "
			symStyle = symStyle.Foreground(theme.ColorPurple).Bold(true)
		}

		priceColor := theme.ColorText
		if !row.hasQuote {
			priceColor = theme.ColorMuted
		}
		priceStr := lipgloss.NewStyle().Foreground(priceColor).Render(fmt.Sprintf("%12s", formatMoney(row.price, row.currency)))
		valueStr := lipgloss.NewStyle().Foreground(theme.ColorText).Render(fmt.Sprintf("%14s", formatMoney(row.value, row.currency)))
		weightStr := lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(fmt.Sprintf("%7.1f%%", row.weight))

		var plStr string
		switch {
		case m.quotesLoading && !row.hasPL:
			plStr = lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(fmt.Sprintf("%9s", "…"))
		case !row.hasPL:
			plStr = lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(fmt.Sprintf("%9s", "—"))
		default:
			plStyle := theme.PositiveChange
			if row.plPct < 0 {
				plStyle = theme.NegativeChange
			}
			plStr = plStyle.Render(fmt.Sprintf("%+8.2f%%", row.plPct))
		}

		sym := symStyle.Render(fmt.Sprintf("%-8s", truncate(row.symbol, 8)))
		lines = append(lines, fmt.Sprintf("%s%s %s %s %s %s", marker, sym, priceStr, valueStr, weightStr, plStr))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) View() string {
	switch m.mode {
	case modeAdd:
		return m.addFormView()
	case modeRemove:
		return m.removeConfirmView()
	case modeChart:
		return m.chartOverlayView()
	case modeNewsDetail:
		return m.newsDetailView()
	default:
		return m.mainView()
	}
}

func (m Model) mainView() string {
	if len(m.Holdings) == 0 {
		empty := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("No holdings yet."),
			"",
			lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true).Render("Press 'a' to add your first position"),
		)
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, empty,
			lipgloss.WithWhitespaceForeground(theme.ColorBg))
	}

	bodyH := max(m.Height-2, 10)
	rows, totalValue := m.computePositions()

	// The positions table needs roughly 53 content columns; below that a
	// side-by-side layout would squeeze both panes into an unreadable,
	// wrapping mess. Stack the insights window below the allocation pane
	// instead once the terminal gets too narrow for two comfortable columns.
	const minLeftW, minRightW = 62, 38

	var body string
	if m.Width >= minLeftW+minRightW+3 {
		leftW := clampInt(m.Width*3/5, minLeftW, m.Width-minRightW-3)
		rightW := m.Width - leftW - 3
		leftPane := m.leftPaneView(rows, totalValue, leftW-2, bodyH)
		rightPane := m.insightsView(rightW-2, bodyH)
		body = lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)
	} else {
		topH := clampInt(bodyH*3/5, 14, bodyH-8)
		botH := max(bodyH-topH-1, 8)
		leftPane := m.leftPaneView(rows, totalValue, m.Width-2, topH)
		rightPane := m.insightsView(m.Width-2, botH)
		body = lipgloss.JoinVertical(lipgloss.Left, leftPane, rightPane)
	}

	footer := lipgloss.NewStyle().Foreground(theme.ColorMuted).
		Render("  a add  d remove  ↵ chart  n news  r refresh  ↑↓ select  ← → browse news  Esc back")

	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

func (m Model) leftPaneView(rows []positionRow, totalValue float64, width, height int) string {
	primaryCurrency := "USD"
	if len(rows) > 0 && rows[0].currency != "" {
		primaryCurrency = rows[0].currency
	}
	header := renderPortfolioHeader(totalValue, primaryCurrency)
	if m.quotesLoading {
		header += lipgloss.NewStyle().Foreground(theme.ColorMuted).Italic(true).Render("  refreshing…")
	}

	innerW := width - 4 // account for Padding(1, 2)
	allocH := clampInt(height*2/5, 8, 17)
	alloc := allocationView(m.Holdings, m.quotes, innerW, allocH)
	alloc = lipgloss.PlaceHorizontal(innerW, lipgloss.Center, alloc)

	table := m.positionsTableView(rows)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		alloc,
		"",
		sectionHeader("POSITIONS"),
		table,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Padding(1, 2).
		Width(width).
		Height(height).
		Render(content)
}

func (m Model) addFormView() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorPurple).
		Padding(1, 3).
		Render(m.form.View())

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceForeground(theme.ColorBg))
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

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorRed).
		Padding(1, 3).
		Render(content)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceForeground(theme.ColorBg))
}

func (m Model) chartOverlayView() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorPurple).
		Render(m.Chart.View())

	hint := lipgloss.NewStyle().Foreground(theme.ColorMuted).
		Render("1/2/3/4 range  Tab chart type  Esc close")

	content := lipgloss.JoinVertical(lipgloss.Center, box, hint)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, content,
		lipgloss.WithWhitespaceForeground(theme.ColorBg))
}

func (m Model) newsDetailView() string {
	var body string
	switch {
	case m.tickerNewsLoading:
		body = lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("Loading news…")
	case m.tickerNewsErr != "":
		body = lipgloss.NewStyle().Foreground(theme.ColorRed).Render("Error: " + m.tickerNewsErr)
	case len(m.tickerNews) == 0:
		body = lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("No news found.")
	default:
		item := m.tickerNews[m.tickerNewsScroll]
		sym := m.Holdings[m.selectedIdx].Symbol

		title := lipgloss.NewStyle().Foreground(theme.ColorPurple).Bold(true).Render("— " + sym + " News —")
		divider := lipgloss.NewStyle().Foreground(theme.ColorBorder).Render(strings.Repeat("─", 50))

		headline := lipgloss.NewStyle().Foreground(theme.ColorText).Bold(true).Render(item.Headline)
		meta := lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(item.Source + " · " + timeAgo(item.Datetime))

		summary := lipgloss.NewStyle().Foreground(theme.ColorText).
			Width(48).Render(truncate(item.Summary, 300))

		nav := ""
		if len(m.tickerNews) > 1 {
			nav = lipgloss.NewStyle().Foreground(theme.ColorMuted).Render(
				fmt.Sprintf("J/K next/prev  (%d/%d)  ", m.tickerNewsScroll+1, len(m.tickerNews)))
		}
		hint := lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("Esc / n close")

		body = lipgloss.JoinVertical(lipgloss.Left,
			title,
			divider,
			"",
			headline,
			meta,
			"",
			summary,
			"",
			nav+hint,
		)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorPurple).
		Padding(2, 3).
		Render(body)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceForeground(theme.ColorBg))
}

func currencySymbol(currency string) string {
	switch currency {
	case "USD":
		return "$"
	case "CAD":
		return "C$"
	case "INR":
		return "₹"
	case "GBP":
		return "£"
	case "EUR":
		return "€"
	case "JPY":
		return "¥"
	case "CNY":
		return "CN¥"
	case "AUD":
		return "A$"
	case "CHF":
		return "CHF"
	case "NZD":
		return "NZ$"
	case "KRW":
		return "₩"
	case "BRL":
		return "R$"
	default:
		return ""
	}
}

// formatMoney renders an amount with thousands separators and its currency
// symbol, e.g. "$12,345.67", "₹1,204.10", or "-C$500.00". When currency is
// empty no symbol is prefixed.
func formatMoney(v float64, currency string) string {
	symbol := currencySymbol(currency)
	neg := v < 0
	if neg {
		v = -v
	}
	whole := int64(v)
	frac := int(math.Round((v - float64(whole)) * 100))
	if frac == 100 {
		whole++
		frac = 0
	}

	ws := strconv.FormatInt(whole, 10)
	var grouped strings.Builder
	for i, d := range ws {
		if i > 0 && (len(ws)-i)%3 == 0 {
			grouped.WriteByte(',')
		}
		grouped.WriteRune(d)
	}

	sign := ""
	if neg {
		sign = "-"
	}
	return fmt.Sprintf("%s%s%s.%02d", sign, symbol, grouped.String(), frac)
}
