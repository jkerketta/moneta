package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jkerketta/stocktui/internal/data"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/home"
	"github.com/jkerketta/stocktui/internal/ui/portfolio"
	"github.com/jkerketta/stocktui/internal/ui/theme"
)

type screen int

const (
	screenHome screen = iota
	screenPortfolio
)

type AppModel struct {
	screen    screen
	home      home.Model
	portfolio portfolio.Model
	width     int
	height    int
}

func New() *AppModel {
	theme.LoadTheme()

	m := &AppModel{
		screen:    screenHome,
		home:      home.New(),
		portfolio: portfolio.New(),
	}
	m.loadPortfolio()
	return m
}

func (m *AppModel) Init() tea.Cmd {
	// Fetch live quotes as soon as the app opens so P&L is ready by the
	// time the user opens View Portfolio.
	var marketCmd tea.Cmd
	m.portfolio, marketCmd = m.portfolio.RefreshMarketData()
	return tea.Batch(tea.EnterAltScreen, m.home.Init(), marketCmd)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Window size changes must reach every screen's model immediately, not
	// just the currently active one - otherwise a screen sized only on
	// first visit (like the portfolio hub's chart overlay) stays 0x0 until
	// a resize happens to occur after navigating to it.
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wm.Width
		m.height = wm.Height
		m.home, _ = m.home.Update(msg)
		var pcmd tea.Cmd
		m.portfolio, pcmd = m.portfolio.Update(msg)
		return m, pcmd
	}

	// Apply quote/news results even while still on the home screen so the
	// startup RefreshMarketData() is not dropped.
	if updated, cmd, handled := m.portfolio.HandleAsyncMsg(msg); handled {
		m.portfolio = updated
		return m, cmd
	}

	var cmd tea.Cmd

	switch m.screen {
	case screenHome:
		m.home, cmd = m.home.Update(msg)
		// Check if user selected a menu item
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "enter" {
			switch m.home.Selected() {
			case "View Portfolio":
				m.screen = screenPortfolio
				m.loadPortfolio()
				var marketCmd tea.Cmd
				m.portfolio, marketCmd = m.portfolio.RefreshMarketData()
				return m, marketCmd
			case "Quit":
				return m, tea.Quit
			}
		}
		return m, cmd

	case screenPortfolio:
		// Capture InForm() before delegating: closing an overlay (chart,
		// add, remove) also flips InForm() false on the very same Esc
		// keypress, so checking afterward would make one Esc both close
		// the overlay AND immediately pop back to Home. Only navigate
		// home when Esc was pressed while already browsing.
		wasInForm := m.portfolio.InForm()
		m.portfolio, cmd = m.portfolio.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && (km.String() == "escape" || km.String() == "esc") && !wasInForm {
			m.savePortfolio()
			m.screen = screenHome
			return m, nil
		}
		return m, cmd
	}

	return m, cmd
}

func (m *AppModel) View() string {
	switch m.screen {
	case screenPortfolio:
		return m.portfolio.View()
	default:
		return m.home.View()
	}
}

func (m *AppModel) loadPortfolio() {
	p, err := data.LoadPortfolio(data.PortfolioPath)
	if err != nil {
		return
	}
	m.portfolio.Holdings = p.Holdings
}

func (m *AppModel) savePortfolio() {
	p := &models.Portfolio{Holdings: m.portfolio.Holdings}
	data.SavePortfolio(data.PortfolioPath, p)
}

func (m *AppModel) Close() {}
