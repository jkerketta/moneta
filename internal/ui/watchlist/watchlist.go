package watchlist

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/styles"
)

type SortMode int

const (
	SortByName SortMode = iota
	SortByPrice
	SortByChange
)

func (s SortMode) String() string {
	switch s {
	case SortByName:
		return "Name"
	case SortByPrice:
		return "Price"
	case SortByChange:
		return "Change%"
	default:
		return "Name"
	}
}

type Model struct {
	list        list.Model
	allItems    []item // Original unfiltered items
	width       int
	height      int
	searchMode  bool
	searchInput textinput.Model
	filterQuery string // Current active filter (persists after search closes)
	sortMode    SortMode
	sortAsc     bool // true = ascending, false = descending
}

type item struct {
	symbol    string
	price     float64
	changePct float64
}

func (i item) Title() string       { return i.symbol }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.symbol }

func New(symbols []string) Model {
	items := make([]item, len(symbols))
	for i, s := range symbols {
		items[i] = item{symbol: s}
	}

	l := list.New(toListItems(items), newDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowPagination(true)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.DisableQuitKeybindings()

	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	ti.CharLimit = 30
	ti.Width = 25

	return Model{
		list:        l,
		allItems:    items,
		searchInput: ti,
		sortMode:    SortByName,
		sortAsc:     true,
	}
}

func toListItems(items []item) []list.Item {
	result := make([]list.Item, len(items))
	for i, it := range items {
		result[i] = it
	}
	return result
}

type delegate struct{}

func newDelegate() delegate { return delegate{} }

func (d delegate) Height() int                               { return 1 }
func (d delegate) Spacing() int                              { return 0 }
func (d delegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it, ok := listItem.(item)
	if !ok {
		return
	}

	// Dynamic widths based on list width
	totalW := m.Width()
	symW := 14
	priceW := 12
	pctW := 9

	if totalW > 40 {
		symW = min(20, totalW-priceW-pctW-2)
	}

	// Symbol - truncate if needed
	sym := it.symbol
	if len(sym) > symW {
		sym = sym[:symW-1] + "…"
	}
	symStr := fmt.Sprintf("%-*s", symW, sym)

	// Price
	var priceStr string
	if it.price == 0 {
		priceStr = fmt.Sprintf("%*s", priceW, "—")
	} else if it.price >= 1000 {
		priceStr = fmt.Sprintf("%*.0f", priceW, it.price)
	} else {
		priceStr = fmt.Sprintf("%*.2f", priceW, it.price)
	}

	// Percent change
	var pctStr string
	if it.price == 0 {
		pctStr = fmt.Sprintf("%*s", pctW, "—")
	} else {
		pctStr = fmt.Sprintf("%+*.2f%%", pctW-1, it.changePct)
	}

	// Style based on selection and trend
	selected := index == m.Index()

	if selected {
		row := fmt.Sprintf("%s %s %s", symStr, priceStr, pctStr)
		fmt.Fprint(w, styles.SelectedItem.Render(row))
	} else {
		symStyled := lipgloss.NewStyle().Foreground(styles.ColorText).Render(symStr)
		priceStyled := lipgloss.NewStyle().Foreground(styles.ColorText).Render(priceStr)

		pctStyle := styles.PositiveChange
		if it.changePct < 0 {
			pctStyle = styles.NegativeChange
		}
		pctStyled := pctStyle.Render(pctStr)

		fmt.Fprint(w, fmt.Sprintf(" %s %s %s", symStyled, priceStyled, pctStyled))
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle search mode input
	if m.searchMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.searchInput.SetValue("")
				m.filterQuery = ""
				m.applyFilter("")
				return m, nil
			case "enter":
				m.searchMode = false
				m.filterQuery = m.searchInput.Value() // Save filter
				return m, nil
			}
		}
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
		m.filterQuery = m.searchInput.Value()
		m.applyFilter(m.filterQuery)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			m.searchMode = true
			m.searchInput.Focus()
			return m, textinput.Blink
		case "s":
			m.cycleSort()
			return m, nil
		case "S":
			m.sortAsc = !m.sortAsc
			m.applySorting()
			return m, nil
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Check if click is within bounds of the pane
			if msg.X >= 0 && msg.X < m.width && msg.Y >= 0 && msg.Y < m.height {
				listHeight := m.list.Height()
				if listHeight > 0 && listHeight <= m.height {
					topOffset := (m.height - listHeight) / 2
					if topOffset < 0 {
						topOffset = 0
					}
					// Account for search bar if visible
					if m.searchMode {
						topOffset += 2
					}
					if msg.Y >= topOffset && msg.Y < topOffset+listHeight {
						localIndex := msg.Y - topOffset
						index := localIndex + m.list.Paginator.Page*m.list.Paginator.PerPage
						if index >= 0 && index < len(m.list.Items()) {
							m.list.Select(index)
						}
					}
				}
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) cycleSort() {
	m.sortMode = (m.sortMode + 1) % 3
	m.applySorting()
}

func (m *Model) applySorting() {
	items := m.getFilteredItems()

	sort.SliceStable(items, func(i, j int) bool {
		var less bool
		switch m.sortMode {
		case SortByName:
			less = strings.ToLower(items[i].symbol) < strings.ToLower(items[j].symbol)
		case SortByPrice:
			less = items[i].price < items[j].price
		case SortByChange:
			less = items[i].changePct < items[j].changePct
		}
		if !m.sortAsc {
			return !less
		}
		return less
	})

	m.list.SetItems(toListItems(items))
}

func (m *Model) getFilteredItems() []item {
	listItems := m.list.Items()
	items := make([]item, len(listItems))
	for i, li := range listItems {
		items[i] = li.(item)
	}
	return items
}

func (m *Model) applyFilter(query string) {
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		m.list.SetItems(toListItems(m.allItems))
		m.applySorting()
		return
	}

	filtered := make([]item, 0)
	for _, it := range m.allItems {
		if strings.Contains(strings.ToLower(it.symbol), query) {
			filtered = append(filtered, it)
		}
	}
	m.list.SetItems(toListItems(filtered))
	m.applySorting()
}

func (m Model) View() string {
	var content string

	if m.searchMode {
		// Prominent search box with border
		searchBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorPrimary).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 1).
			Width(m.width - 6)

		labelStyle := lipgloss.NewStyle().
			Foreground(styles.ColorPrimary).
			Bold(true)

		typedTextStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

		cursorStyle := lipgloss.NewStyle().
			Foreground(styles.ColorPrimary).
			Bold(true).
			Blink(true)

		hintStyle := lipgloss.NewStyle().
			Foreground(styles.ColorSubtext).
			Italic(true)

		resultCountStyle := lipgloss.NewStyle().
			Foreground(styles.ColorSuccess)

		// Build search box content - show typed text directly
		searchLabel := labelStyle.Render("🔍 Search: ")
		typedText := m.searchInput.Value()
		if typedText == "" {
			typedText = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("type to filter...")
		} else {
			typedText = typedTextStyle.Render(typedText)
		}
		cursor := cursorStyle.Render("█")
		searchField := typedText + cursor

		// Show result count
		resultCount := len(m.list.Items())
		totalCount := len(m.allItems)
		countStr := resultCountStyle.Render(fmt.Sprintf(" (%d/%d)", resultCount, totalCount))

		searchLine := searchLabel + searchField + countStr
		hintLine := hintStyle.Render("Enter to confirm • Esc to cancel")

		searchBoxContent := searchLine + "\n" + hintLine
		searchBox := searchBoxStyle.Render(searchBoxContent)

		content = searchBox + "\n" + m.list.View()
	} else {
		// Show sort indicator in header
		sortIndicator := ""
		if m.sortMode != SortByName || !m.sortAsc {
			arrow := "↑"
			if !m.sortAsc {
				arrow = "↓"
			}
			sortIndicator = lipgloss.NewStyle().
				Foreground(styles.ColorSubtext).
				Render(fmt.Sprintf(" [%s %s]", m.sortMode.String(), arrow))
		}

		if sortIndicator != "" {
			content = sortIndicator + "\n" + m.list.View()
		} else {
			content = m.list.View()
		}
	}

	return styles.Pane.
		Width(m.width).
		Height(m.height).
		Render(content)
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.list.SetSize(w-4, h-4)
	m.searchInput.Width = w - 8
}

func (m *Model) UpdateQuotes(quotes []models.Quote) {
	qmap := make(map[string]models.Quote, len(quotes))
	for _, q := range quotes {
		qmap[q.Symbol] = q
	}

	// Update allItems with new data
	for i, it := range m.allItems {
		if q, ok := qmap[it.symbol]; ok {
			m.allItems[i].price = q.Price
			m.allItems[i].changePct = q.ChangePct
		}
	}

	// Re-apply filter and sort to update the visible list
	m.applyFilter(m.filterQuery)
}

// UpdatePriceChange updates change % for a symbol based on historical data
func (m *Model) UpdatePriceChange(symbol string, currentPrice, startPrice float64) {
	// Update in allItems
	for i, it := range m.allItems {
		if it.symbol == symbol {
			m.allItems[i].price = currentPrice
			if startPrice > 0 {
				m.allItems[i].changePct = ((currentPrice - startPrice) / startPrice) * 100
			}
			break
		}
	}

	// Re-apply filter and sort to update the visible list
	m.applyFilter(m.filterQuery)
}

func (m Model) SelectedSymbol() string {
	if it, ok := m.list.SelectedItem().(item); ok {
		return it.symbol
	}
	return ""
}

func (m Model) IsSearching() bool {
	return m.searchMode
}

// SortInfo returns current sort mode and direction
func (m Model) SortInfo() (SortMode, bool) {
	return m.sortMode, m.sortAsc
}
