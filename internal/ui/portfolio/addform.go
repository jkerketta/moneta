package portfolio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

const (
	fieldTicker   = 0
	fieldShares   = 1
	fieldPrice    = 2
	fieldCurrency = 3
	fieldCount    = 4
)

// addForm is the single-window "Add Position" modal: one small dialog with
// fields for ticker, shares, average price, and currency.
type addForm struct {
	inputs   [3]textinput.Model // ticker, shares, price
	focus    int
	errMsg   string
	currency currencyPicker
}

var currencyOptions = []string{"USD", "CAD", "INR", "GBP", "EUR", "JPY", "CNY", "AUD", "CHF", "NZD", "KRW", "BRL"}

type currencyPicker struct {
	currencies  []string
	selectedIdx int
	focused     bool
}

func newCurrencyPicker() currencyPicker {
	return currencyPicker{
		currencies:  currencyOptions,
		selectedIdx: 0,
	}
}

func (p currencyPicker) Value() string {
	return p.currencies[p.selectedIdx]
}

func (p *currencyPicker) set(code string) {
	for i, c := range p.currencies {
		if c == code {
			p.selectedIdx = i
			return
		}
	}
}

func (p *currencyPicker) setFocused(f bool) {
	p.focused = f
}

func (p currencyPicker) Update(msg tea.KeyMsg) (currencyPicker, tea.Cmd) {
	switch msg.String() {
	case "left":
		if p.selectedIdx > 0 {
			p.selectedIdx--
		}
	case "right":
		if p.selectedIdx < len(p.currencies)-1 {
			p.selectedIdx++
		}
	}
	return p, nil
}

func (p currencyPicker) View() string {
	cur := p.currencies[p.selectedIdx]
	s := lipgloss.NewStyle().Width(20)
	if p.focused {
		s = s.Foreground(theme.Current().Accent).Bold(true)
		return s.Render(fmt.Sprintf("%s  ◀ ▶", cur))
	}
	return s.Foreground(theme.ColorText).Render(cur)
}

// detectCurrency infers a currency code from the ticker suffix conventions.
func detectCurrency(symbol string) string {
	s := strings.ToUpper(symbol)
	switch {
	case strings.HasSuffix(s, ".NS"), strings.HasSuffix(s, ".BO"):
		return "INR"
	case strings.HasSuffix(s, ".TO"), strings.HasSuffix(s, ".V"):
		return "CAD"
	case strings.HasSuffix(s, "-USD"):
		return "USD"
	case strings.HasSuffix(s, ".L"), strings.HasSuffix(s, ".LON"):
		return "GBP"
	case strings.HasSuffix(s, ".DE"), strings.HasSuffix(s, ".F"):
		return "EUR"
	case strings.HasSuffix(s, ".T"):
		return "JPY"
	case strings.HasSuffix(s, ".KS"), strings.HasSuffix(s, ".KQ"):
		return "KRW"
	default:
		return "USD"
	}
}

func newAddForm() (addForm, tea.Cmd) {
	labels := []struct {
		placeholder string
		charLimit   int
	}{
		{"AAPL", 12},
		{"10", 12},
		{"150.00", 12},
	}

	var f addForm
	for i, l := range labels {
		ti := textinput.New()
		ti.Placeholder = l.placeholder
		ti.CharLimit = l.charLimit
		ti.Width = 20
		ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.ColorMuted)
		ti.TextStyle = lipgloss.NewStyle().Foreground(theme.ColorText)
		ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.ColorBorder)
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(theme.Current().Accent)
		f.inputs[i] = ti
	}

	f.currency = newCurrencyPicker()
	cmd := f.focusField(fieldTicker)
	return f, cmd
}

func (f *addForm) focusField(i int) tea.Cmd {
	for j := range f.inputs {
		f.inputs[j].Blur()
	}
	f.currency.setFocused(false)
	f.focus = i
	if i == fieldCurrency {
		f.currency.setFocused(true)
		return nil
	}
	return f.inputs[i].Focus()
}

func (f *addForm) autoDetectCurrency() {
	ticker := strings.TrimSpace(f.inputs[fieldTicker].Value())
	if ticker == "" {
		return
	}
	f.currency.set(detectCurrency(ticker))
}

// validate parses the current field values into a Holding, setting errMsg
// and reporting failure when something isn't fillable yet.
func (f *addForm) validate() (models.Holding, bool) {
	ticker := strings.ToUpper(strings.TrimSpace(f.inputs[fieldTicker].Value()))
	if ticker == "" {
		f.errMsg = "Ticker is required"
		return models.Holding{}, false
	}

	shares, err := strconv.ParseFloat(strings.TrimSpace(f.inputs[fieldShares].Value()), 64)
	if err != nil || shares <= 0 {
		f.errMsg = "Shares must be a positive number"
		return models.Holding{}, false
	}

	price, err := strconv.ParseFloat(strings.TrimSpace(f.inputs[fieldPrice].Value()), 64)
	if err != nil || price <= 0 {
		f.errMsg = "Avg price must be a positive number"
		return models.Holding{}, false
	}

	f.errMsg = ""
	return models.Holding{
		Symbol:   ticker,
		Shares:   shares,
		AvgPrice: price,
		Currency: f.currency.Value(),
	}, true
}

// Update handles the form's own navigation (tab/enter/esc) and otherwise
// delegates to the focused text input. It reports whether the form was
// submitted or cancelled this step.
func (f addForm) Update(msg tea.Msg) (addForm, tea.Cmd, bool, bool) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc", "escape":
			return f, nil, false, true
		case "tab", "down":
			if f.focus == fieldTicker {
				f.autoDetectCurrency()
			}
			cmd := f.focusField((f.focus + 1) % fieldCount)
			return f, cmd, false, false
		case "shift+tab", "up":
			cmd := f.focusField((f.focus - 1 + fieldCount) % fieldCount)
			return f, cmd, false, false
		case "enter":
			if f.focus < fieldCount-1 {
				if f.focus == fieldTicker {
					f.autoDetectCurrency()
				}
				cmd := f.focusField(f.focus + 1)
				return f, cmd, false, false
			}
			if _, ok := f.validate(); ok {
				return f, nil, true, false
			}
			return f, nil, false, false
		}

		if f.focus == fieldCurrency && (km.String() == "left" || km.String() == "right") {
			var cmd tea.Cmd
			f.currency, cmd = f.currency.Update(km)
			return f, cmd, false, false
		}
	}

	if f.focus == fieldCurrency {
		return f, nil, false, false
	}

	var cmd tea.Cmd
	f.inputs[f.focus], cmd = f.inputs[f.focus].Update(msg)
	return f, cmd, false, false
}

func (f addForm) View() string {
	title := lipgloss.NewStyle().Foreground(theme.Current().Accent).Bold(true).Render("Add Position")
	labelStyle := lipgloss.NewStyle().Foreground(theme.ColorMuted).Width(11)

	row := func(label string, content string) string {
		return labelStyle.Render(label) + content
	}

	rows := lipgloss.JoinVertical(lipgloss.Left,
		row("Ticker", f.inputs[fieldTicker].View()),
		row("Shares", f.inputs[fieldShares].View()),
		row("Avg Price", f.inputs[fieldPrice].View()),
		row("Currency", f.currency.View()),
	)

	help := lipgloss.NewStyle().Foreground(theme.ColorMuted).Render("Tab/↑↓ move · ←/→ cycle currency · Enter save · Esc cancel")

	parts := []string{title, "", rows}
	if f.errMsg != "" {
		parts = append(parts, "", lipgloss.NewStyle().Foreground(theme.ColorRed).Render(f.errMsg))
	}
	parts = append(parts, "", help)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
