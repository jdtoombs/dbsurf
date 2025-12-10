package app

import (
	"dbsurf/db"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type connectionTestMsg struct {
	err error
}

func (a *App) testConnection(connString string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Connect(connString)
		if conn != nil {
			conn.Close()
		}
		return connectionTestMsg{err: err}
	}
}

func (a *App) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.mode = modeList
		a.connInput.Reset()
		a.nameInput.Reset()
		a.inputErr = nil
		a.inputStep = 0
		return a, nil
	case "enter":
		if a.inputStep == 0 {
			connString := strings.TrimSpace(a.connInput.Value())

			// Format validation
			if err := db.ValidateConnectionString(connString); err != nil {
				a.inputErr = err
				return a, nil
			}

			// Start live connection test
			a.inputTesting = true
			a.inputErr = nil
			return a, tea.Batch(a.testConnection(connString), a.inputSpinner.Tick)
		}
		// Step 1: save connection
		a.config.AddConnection(
			a.nameInput.Value(),
			a.connInput.Value(),
			db.DetectDBType(a.connInput.Value()),
		)
		a.config.Save()
		a.mode = modeList
		a.connInput.Reset()
		a.nameInput.Reset()
		a.inputErr = nil
		a.inputStep = 0
		return a, nil
	}

	// Clear error when user types (only in step 0)
	if a.inputStep == 0 {
		a.inputErr = nil
	}

	var cmd tea.Cmd
	if a.inputStep == 0 {
		a.connInput, cmd = a.connInput.Update(msg)
	} else {
		a.nameInput, cmd = a.nameInput.Update(msg)
	}
	return a, cmd
}

func (a *App) handleConnectionTestResult(msg connectionTestMsg) (tea.Model, tea.Cmd) {
	a.inputTesting = false
	if msg.err != nil {
		a.inputErr = fmt.Errorf("connection failed: %w", msg.err)
		return a, nil
	}
	// Connection successful - proceed to name input
	a.inputStep = 1
	a.connInput.Blur()
	a.nameInput.Focus()
	return a, textinput.Blink
}

func (a *App) viewInput() string {
	var content string

	if a.inputStep == 0 {
		content = "Connection string:\n\n" + a.connInput.View()

		if a.inputTesting {
			content += "\n\n" + a.inputSpinner.View() + " Testing connection..."
		}

		if a.inputErr != nil {
			content += "\n\n" + errorStyle.Render("Error: "+a.inputErr.Error())
		}
	} else {
		content = "Name for this connection:\n\n" + a.nameInput.View()
	}

	controls := "enter: submit • esc: cancel"

	return a.renderFrame(content, controls)
}
