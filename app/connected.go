package app

import (
	"dbsurf/db"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) filterDatabases() {
	query := strings.ToLower(a.dbSearchInput.Value())
	if query == "" {
		a.filteredDatabases = a.databases
		return
	}
	a.filteredDatabases = nil
	for _, db := range a.databases {
		if strings.Contains(strings.ToLower(db), query) {
			a.filteredDatabases = append(a.filteredDatabases, db)
		}
	}
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
		a.mode = modeList
		return a, nil
	case "/":
		a.dbSearching = true
		a.dbSearchInput.Focus()
		return a, textinput.Blink
	case "j", "down":
		if a.dbCursor < len(a.filteredDatabases)-1 {
			a.dbCursor++
		}
	case "k", "up":
		if a.dbCursor > 0 {
			a.dbCursor--
		}
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
			content += "Search: " + a.dbSearchInput.View() + "\n\n"
		}

		content += "Databases:\n"
		if len(a.filteredDatabases) > 0 {
			visibleCount := 10

			scrollOffset := 0
			if a.dbCursor >= visibleCount {
				scrollOffset = a.dbCursor - visibleCount + 1
			}

			start := scrollOffset
			end := scrollOffset + visibleCount
			if end > len(a.filteredDatabases) {
				end = len(a.filteredDatabases)
			}

			var lines []string
			for i := start; i < end; i++ {
				cursor := "  "
				line := a.filteredDatabases[i]
				if i == a.dbCursor {
					cursor = "> "
					line = selectedStyle.Render(line)
				}
				lines = append(lines, cursor+line)
			}
			content += strings.Join(lines, "\n")
		} else {
			content += dimStyle.Render("No databases found")
		}
	}

	controls := "j/k: navigate • /: search • enter: select • esc: back • q: quit"

	return a.renderFrame(content, controls)
}
