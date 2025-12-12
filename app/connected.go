package app

import (
	"dbsurf/db"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) filterDatabases() {
	a.filteredDatabases = filterStrings(a.databases, a.dbSearchInput.Value())
}

func (a *App) updateConnected(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.dbSearching {
		switch msg.String() {
		case "esc":
			a.dbSearching = false
			a.dbSearchInput.Reset()
			a.filteredDatabases = a.databases
			a.dbCursor = 0
			return a, nil
		case "enter":
			a.dbSearching = false
			a.dbSearchInput.Blur()
			a.dbCursor = 0
			return a, nil
		}
		var cmd tea.Cmd
		a.dbSearchInput, cmd = a.dbSearchInput.Update(msg)
		a.filterDatabases()
		a.dbCursor = 0
		return a, cmd
	}

	switch msg.String() {
	case "esc":
		if a.db != nil {
			a.db.Close()
			a.db = nil
		}
		a.mode = modeList
		return a, nil
	case "/":
		a.dbSearching = true
		a.dbSearchInput.Focus()
		return a, textinput.Blink
	case "j", "down":
		a.dbCursor = moveCursor(a.dbCursor, 1, len(a.filteredDatabases))
	case "k", "up":
		a.dbCursor = moveCursor(a.dbCursor, -1, len(a.filteredDatabases))
	case "enter":
		if len(a.filteredDatabases) > 0 {
			a.selectedDatabase = a.filteredDatabases[a.dbCursor]
			// Skip UseDatabase for SQL Server - handled via query prepend
			if a.dbType != "sqlserver" {
				if err := db.UseDatabase(a.db, a.selectedDatabase, a.dbType); err != nil {
					a.queryErr = err
				}
			}
			a.queryInput.Reset()
			a.queryInput.Focus()
			a.queryFocused = true
			a.queryResult = nil
			a.queryErr = nil
			a.mode = modeQuery
			return a, textinput.Blink
		}
	}
	return a, nil
}

func (a *App) viewConnected() string {
	var content string
	conn := a.config.Connections[a.cursor]

	if a.dbErr != nil {
		content = "Connection failed: " + a.dbErr.Error()
	} else {
		content = "Connected to " + conn.Name + "\n\n"

		if a.dbSearching {
			content += inputLabelStyle.Render("Search: ") + a.dbSearchInput.View() + "\n\n"
		} else if a.dbSearchInput.Value() != "" {
			content += dimStyle.Render("Search: "+a.dbSearchInput.Value()) + "\n\n"
		} else {
			content += dimStyle.Render("Search: press / to filter") + "\n\n"
		}

		content += "Databases:\n"
		if len(a.filteredDatabases) > 0 {
			var lines []string
			for i, db := range a.filteredDatabases {
				prefix := "  "
				line := db
				if i == a.dbCursor {
					prefix = "> "
					line = selectedStyle.Render(line)
				}
				lines = append(lines, prefix+line)
			}
			a.viewport.SetContent(strings.Join(lines, "\n"))
			a.syncViewportToCursor(a.dbCursor, len(a.filteredDatabases))
			content += a.viewport.View()
		} else {
			content += dimStyle.Render("No databases found")
		}
	}

	controls := "j/k: navigate • /: search • enter: select • esc: back • q: quit"

	return a.renderFrame(content, controls)
}
