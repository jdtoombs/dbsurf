package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil
	case tea.KeyMsg:
		if msg.String() == "q" {
			return a, tea.Quit
		}
		switch a.mode {
		case modeList:
			return a.updateList(msg)
		case modeInput:
			return a.updateInput(msg)
		case modeConnected:
			return a.updateConnected(msg)
		}
	}
	return a, nil
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	switch a.mode {
	case modeList:
		return a.viewList()
	case modeInput:
		return a.viewInput()
	case modeConnected:
		return a.viewConnected()
	default:
		return a.viewList()
	}
}

func (a *App) renderFrame(content, controls string) string {
	logoRendered := titleStyle.Render(Logo)

	boxWidth := a.width - 4
	if boxWidth > 80 {
		boxWidth = 80
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	box := boxStyle.Width(boxWidth - 4) 
	boxedContent := box.Render(content)

	logoRendered = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, logoRendered)
	boxedContent = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, boxedContent)
	controlsRendered := lipgloss.PlaceHorizontal(a.width, lipgloss.Center, dimStyle.Render(controls))

	contentHeight := strings.Count(logoRendered, "\n") + strings.Count(boxedContent, "\n") + 3
	bottomPadding := a.height - contentHeight - 2
	if bottomPadding < 0 {
		bottomPadding = 0
	}

	return logoRendered + "\n" +
		boxedContent + "\n" +
		strings.Repeat("\n", bottomPadding) +
		controlsRendered
}
