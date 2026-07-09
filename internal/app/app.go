package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jkerketta/stocktui/internal/data"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/alerts"
	"github.com/jkerketta/stocktui/internal/ui/home"
	"github.com/jkerketta/stocktui/internal/ui/portfolio"
)

type navigateBack struct{}

func backCmd() tea.Msg { return navigateBack{} }

type screen int

const (
	screenHome screen = iota
	screenPortfolio
	screenAlerts
)

type AppModel struct {
	screen    screen
	home      home.Model
	portfolio portfolio.Model
	alerts    alerts.Model
	width     int
	height    int
}

func New() *AppModel {
	// Load Finnhub API key from various .env locations
	data.LoadEnv(".env")
	data.LoadEnv("../moneta/.env")
	data.LoadEnv("/Users/josekerketta/firstmate/projects/moneta/.env")

	m := &AppModel{
		screen:    screenHome,
		home:      home.New(),
		portfolio: portfolio.New(),
		alerts:    alerts.New(),
	}
	m.loadPortfolio()
	return m
}

func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.home.Init())
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size globally
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wm.Width
		m.height = wm.Height
		m.home.Width = wm.Width
		m.home.Height = wm.Height
		m.portfolio.Width = wm.Width
		m.portfolio.Height = wm.Height
		m.alerts.Width = wm.Width
		m.alerts.Height = wm.Height
	}

	var cmd tea.Cmd

	switch m.screen {
	case screenHome:
		m.home, cmd = m.home.Update(msg)
		// Check if user selected a menu item
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "enter" {
			sel := m.home.Selected()
			switch sel {
			case "View Portfolio":
				m.screen = screenPortfolio
				m.loadPortfolio()
			case "News":
				m.screen = screenAlerts
				symbols := make([]string, len(m.portfolio.Holdings))
				for i, h := range m.portfolio.Holdings {
					symbols[i] = h.Symbol
				}
				m.alerts = m.alerts.SetSymbols(symbols)
				return m, m.alerts.Init()
			case "Quit":
				return m, tea.Quit
			}
		}
		return m, cmd

	case screenPortfolio:
		m.portfolio, cmd = m.portfolio.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && (km.String() == "escape" || km.String() == "esc") {
			if !m.portfolio.InForm() {
				m.savePortfolio()
				m.screen = screenHome
				return m, nil
			}
		}
		return m, cmd

	case screenAlerts:
		m.alerts, cmd = m.alerts.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && (km.String() == "escape" || km.String() == "esc") {
			m.screen = screenHome
			return m, nil
		}
		return m, cmd
	}

	return m, cmd
}

func (m *AppModel) View() string {
	switch m.screen {
	case screenHome:
		return m.home.View()
	case screenPortfolio:
		return m.portfolio.View()
	case screenAlerts:
		return m.alerts.View()
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
