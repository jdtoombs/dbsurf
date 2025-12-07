package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (a App) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "j", "down":
		if a.cursor < len(a.config.Connections)-1 {
			a.cursor++
		}
	case "k", "up":
		if a.cursor > 0 {
			a.cursor--
		}
	case "n":
		a.mode = modeInput
		a.inputStep = 0
		a.connInput.Focus()
		return a, textinput.Blink
	case "d":
		if len(a.config.Connections) > 0 {
			a.config.Connections = append(
				a.config.Connections[:a.cursor],
				a.config.Connections[a.cursor+1:]...,
			)
			a.config.Save()
			if a.cursor >= len(a.config.Connections) && a.cursor > 0 {
				a.cursor--
			}
		}
	}
	return a, nil
}

func (a App) viewList() string {
	var content string

	if len(a.config.Connections) == 0 {
		content = dimStyle.Render("No saved connections")
	} else {
		var lines []string
		for i, conn := range a.config.Connections {
			cursor := "  "
			line := fmt.Sprintf("%s (%s)", conn.Name, conn.DBType)
			if i == a.cursor {
				cursor = "> "
				line = selectedStyle.Render(line)
			}
			lines = append(lines, cursor+line)
		}
		content = strings.Join(lines, "\n")
	}

	controls := "j/k: navigate • n: new • d: delete • q: quit"

	return a.renderFrame(content, controls)
}

// renderFrame wraps content in the standard layout (logo + box + controls)
func (a App) renderFrame(content, controls string) string {
	logoRendered := titleStyle.Render(Logo)
	boxedContent := boxStyle.Render(content)

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
