package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jkerketta/stocktui/internal/data"
	"github.com/jkerketta/stocktui/internal/models"
	"github.com/jkerketta/stocktui/internal/ui/alerts"
	"github.com/jkerketta/stocktui/internal/ui/home"
	"github.com/jkerketta/stocktui/internal/ui/portfolio"
)

type screen int

const (
	screenHome screen = iota
	screenPortfolio
	screenAlerts
	screenAdd
)

type AppModel struct {
	screen    screen
	home      home.Model
	portfolio portfolio.Model
	alerts    alerts.Model
	holdings  []models.Holding
	width     int
	height    int
	err       error
}

func New() *AppModel {
	return &AppModel{
		screen:    screenHome,
		home:      home.New(),
		portfolio: portfolio.New(),
		alerts:    alerts.New(),
	}
}

func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.home.Init())
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.screen == screenHome {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				sel := m.home.Selected()
				switch sel {
				case "View Portfolio":
					m.screen = screenPortfolio
					m.loadPortfolio()
				case "Alerts & News":
					m.screen = screenAlerts
				case "Add/Remove Position":
					m.screen = screenAdd
				case "Quit":
					return m, tea.Quit
				}
				return m, nil
			}
		}
		if msg.String() == "escape" || msg.String() == "esc" {
			m.screen = screenHome
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenHome:
		m.home, cmd = m.home.Update(msg)
	case screenPortfolio:
		m.portfolio, cmd = m.portfolio.Update(msg)
	case screenAlerts:
		m.alerts, cmd = m.alerts.Update(msg)
	case screenAdd:
		// TODO: handle add form
		m.screen = screenHome
		return m, nil
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
		m.err = err
		return
	}
	m.holdings = p.Holdings
	m.portfolio.Holdings = p.Holdings
}

func (m *AppModel) Close() {}
