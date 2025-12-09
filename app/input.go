package app

import (
	"dbsurf/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.mode = modeList
		a.connInput.Reset()
		a.nameInput.Reset()
		return a, nil
	case "enter":
		if a.inputStep == 0 {
			a.inputStep = 1
			a.connInput.Blur()
			a.nameInput.Focus()
			return a, textinput.Blink
		}
		a.config.AddConnection(
			a.nameInput.Value(),
			a.connInput.Value(),
			db.DetectDBType(a.connInput.Value()),
		)
		a.config.Save()
		a.mode = modeList
		a.connInput.Reset()
		a.nameInput.Reset()
		return a, nil
	}

	var cmd tea.Cmd
	if a.inputStep == 0 {
		a.connInput, cmd = a.connInput.Update(msg)
	} else {
		a.nameInput, cmd = a.nameInput.Update(msg)
	}
	return a, cmd
}

func (a *App) viewInput() string {
	var content string

	if a.inputStep == 0 {
		content = "Connection string:\n\n" + a.connInput.View()
	} else {
		content = "Name for this connection:\n\n" + a.nameInput.View()
	}

	controls := "enter: submit • esc: cancel"

	return a.renderFrame(content, controls)
}
