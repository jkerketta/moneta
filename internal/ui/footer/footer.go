package footer

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jkerketta/stocktui/internal/models"
)

type Model struct {
	width      int
	provider   string
	lastUpdate time.Time
	connected  bool
	err        error
	timeRange  models.TimeRange
}

func New(provider string) Model {
	return Model{
		provider:  provider,
		connected: true,
		timeRange: models.Range24H,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m *Model) SetSize(w, h int) {
	m.width = w
}

func (m *Model) SetStatus(lastUpdate time.Time, connected bool, err error) {
	m.lastUpdate = lastUpdate
	m.connected = connected
	m.err = err
}

func (m *Model) SetTimeRange(tr models.TimeRange) {
	m.timeRange = tr
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Background(lipgloss.Color("#1a1a2e"))

	accent := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#1a1a2e")).
		Bold(true)

	statusColor := lipgloss.Color("#04B575")
	statusText := "●"
	if !m.connected {
		statusColor = lipgloss.Color("#FF4C4C")
		statusText = "○"
	} else if m.err != nil {
		statusColor = lipgloss.Color("#FF4C4C")
		statusText = "○"
	}
	statusStyle := base.Copy().Foreground(statusColor)

	left := fmt.Sprintf(" %s %s ", statusStyle.Render(statusText), base.Render(m.provider))

	timeRanges := []models.TimeRange{models.Range1H, models.Range24H, models.Range7D, models.Range30D}
	var rangeStr string
	for _, tr := range timeRanges {
		if tr == m.timeRange {
			rangeStr += accent.Render(fmt.Sprintf(" [%s] ", tr))
		} else {
			rangeStr += base.Render(fmt.Sprintf(" %s ", tr))
		}
	}

	center := rangeStr

	timeStr := m.lastUpdate.Format("15:04:05")
	if m.err != nil {
		timeStr = "Error"
	}
	right := base.Render(fmt.Sprintf(" %s  / Search  s Sort  ? Help  q Quit ", timeStr))

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	centerW := m.width - leftW - rightW

	if centerW < 0 {
		centerW = 0
	}

	centeredCenter := lipgloss.PlaceHorizontal(centerW, lipgloss.Center, center)

	bar := lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a2e")).
		Width(m.width).
		Render(left + centeredCenter + right)

	return bar
}
