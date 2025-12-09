package app

import (
	"dbsurf/db"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) filterTables() {
	a.filteredTables = filterStrings(a.tables, a.tableSearchInput.Value())
}

func (a *App) updateTableList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.tableSearching {
		switch msg.String() {
		case "esc":
			a.tableSearching = false
			a.tableSearchInput.Reset()
			a.filteredTables = a.tables
			a.tableCursor = 0
			return a, nil
		case "enter":
			a.tableSearching = false
			a.tableSearchInput.Blur()
			a.tableCursor = 0
			return a, nil
		}
		var cmd tea.Cmd
		a.tableSearchInput, cmd = a.tableSearchInput.Update(msg)
		a.filterTables()
		a.tableCursor = 0
		return a, cmd
	}

	switch msg.String() {
	case "esc":
		a.mode = modeQuery
		return a, nil
	case "/":
		a.tableSearching = true
		a.tableSearchInput.Focus()
		return a, textinput.Blink
	case "j", "down":
		a.tableCursor = moveCursor(a.tableCursor, 1, len(a.filteredTables))
	case "k", "up":
		a.tableCursor = moveCursor(a.tableCursor, -1, len(a.filteredTables))
	case "enter":
		if len(a.filteredTables) > 0 {
			tableName := a.filteredTables[a.tableCursor]
			query := fmt.Sprintf("SELECT * FROM %s", tableName)
			if a.dbType == "sqlserver" {
				query = fmt.Sprintf("USE [%s]; SELECT * FROM %s", a.selectedDatabase, tableName)
			}
			result, err := db.RunQuery(a.db, query)
			if err != nil {
				a.queryErr = err
				a.queryResult = nil
				a.filteredResultRows = nil
			} else {
				a.queryErr = nil
				a.queryResult = result
				a.queryInput.SetValue(fmt.Sprintf("SELECT * FROM %s", tableName))
				a.resultFilter = ""
				a.resultSearchInput.SetValue("")
				a.filterResults()
				a.resultCursor = 0
			}
			a.mode = modeQuery
			return a, nil
		}
	}
	return a, nil
}

func (a *App) viewTableList() string {
	var content string

	content = "Database: " + selectedStyle.Render(a.selectedDatabase) + "\n\n"

	if a.tableSearching {
		content += "Search: " + a.tableSearchInput.View() + "\n\n"
	} else if a.tableSearchInput.Value() != "" {
		content += dimStyle.Render("Search: "+a.tableSearchInput.Value()) + "\n\n"
	}

	content += "Tables:\n"
	if len(a.filteredTables) > 0 {
		visibleCount := 10

		scrollOffset := 0
		if a.tableCursor >= visibleCount {
			scrollOffset = a.tableCursor - visibleCount + 1
		}

		start := scrollOffset
		end := scrollOffset + visibleCount
		if end > len(a.filteredTables) {
			end = len(a.filteredTables)
		}

		var lines []string
		for i := start; i < end; i++ {
			cursor := "  "
			line := a.filteredTables[i]
			if i == a.tableCursor {
				cursor = "> "
				line = selectedStyle.Render(line)
			}
			lines = append(lines, cursor+line)
		}
		content += strings.Join(lines, "\n")
	} else {
		content += dimStyle.Render("No tables found")
	}

	controls := "j/k: navigate • /: search • enter: select • esc: back • q: quit"

	return a.renderFrame(content, controls)
}
