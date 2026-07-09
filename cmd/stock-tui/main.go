package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jkerketta/stocktui/internal/app"
)

func main() {
	model := app.New()
	defer model.Close()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
