package app

import (
	"dbsurf/db"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		a.cursor = moveCursor(a.cursor, 1, len(a.config.Connections))
	case "k", "up":
		a.cursor = moveCursor(a.cursor, -1, len(a.config.Connections))
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
	case "enter":
		if len(a.config.Connections) > 0 {
			if a.db != nil {
				a.db.Close()
				a.db = nil
			}
			conn := a.config.Connections[a.cursor]
			a.dbType = conn.DBType
			a.db, a.dbErr = db.Connect(conn.ConnString)
			if a.dbErr == nil {
				if a.dbType == "postgres" {
					a.selectedDatabase = ""
					a.queryInput.Reset()
					a.queryInput.Focus()
					a.queryFocused = true
					a.queryResult = nil
					a.queryErr = nil
					a.mode = modeQuery
					return a, textinput.Blink
				}
				a.databases, _ = db.ListDatabases(a.db, a.dbType)
			}
			a.filteredDatabases = a.databases
			a.dbCursor = 0
			a.dbSearching = false
			a.dbSearchInput.Reset()
			a.mode = modeConnected
		}
	}
	return a, nil
}

func (a *App) viewList() string {
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
