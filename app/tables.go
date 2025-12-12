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
	if a.showingColumnInfo {
		if a.columnInfoSearching {
			switch msg.String() {
			case "esc":
				a.columnInfoSearching = false
				a.columnInfoSearchInput.Blur()
				return a, nil
			case "enter":
				a.columnInfoSearching = false
				a.columnInfoSearchInput.Blur()
				return a, nil
			}
			var cmd tea.Cmd
			a.columnInfoSearchInput, cmd = a.columnInfoSearchInput.Update(msg)
			a.columnInfoFilter = a.columnInfoSearchInput.Value()
			a.filterAndRebuildColumnInfo()
			return a, cmd
		}

		switch msg.String() {
		case "esc":
			if a.columnInfoFilter != "" {
				a.columnInfoFilter = ""
				a.columnInfoSearchInput.SetValue("")
				a.filterAndRebuildColumnInfo()
				return a, nil
			}
			a.showingColumnInfo = false
			return a, nil
		case "?", "q":
			a.showingColumnInfo = false
			return a, nil
		case "/":
			a.columnInfoSearching = true
			a.columnInfoSearchInput.Focus()
			return a, textinput.Blink
		}
		var cmd tea.Cmd
		a.columnInfoTable, cmd = a.columnInfoTable.Update(msg)
		return a, cmd
	}

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
	case "?":
		if len(a.filteredTables) > 0 {
			tableName := a.filteredTables[a.tableCursor]
			cleanName := db.CleanTableName(tableName, a.dbType)
			info, err := db.GetColumnInfo(a.db, a.selectedDatabase, cleanName, a.dbType)
			if err == nil && len(info) > 0 {
				a.queryTableName = cleanName
				a.columnInfoData = info
				a.columnInfoFilter = ""
				a.columnInfoSearchInput.SetValue("")
				a.filteredColumnInfo = info
				tableHeight := min(len(info), 15)
				a.columnInfoTable = buildColumnInfoTable(info, tableHeight)
				a.showingColumnInfo = true
			}
		}
	case "enter":
		if len(a.filteredTables) > 0 {
			tableName := a.filteredTables[a.tableCursor]
			selectQuery := fmt.Sprintf("SELECT * FROM %s", tableName)
			fullQuery := db.PrependUseDatabase(selectQuery, a.selectedDatabase, a.dbType)
			result, err := db.RunQuery(a.db, fullQuery)
			if err != nil {
				a.queryErr = err
				a.queryResult = nil
				a.filteredResultRows = nil
			} else {
				a.queryErr = nil
				a.queryResult = result
				a.queryInput.SetValue(selectQuery)
				a.resultFilter = ""
				a.resultSearchInput.SetValue("")
				a.filterResults()
				a.resultCursor = 0
				a.fieldCursor = 0
				cleanName := db.CleanTableName(tableName, a.dbType)
				a.queryTableName = cleanName
				a.queryPKColumns, _ = db.GetPrimaryKey(a.db, a.selectedDatabase, cleanName, a.dbType)
			}
			a.mode = modeQuery
			return a, nil
		}
	}
	return a, nil
}

func (a *App) viewTableList() string {
	var content string

	if a.showingColumnInfo {
		content = selectedStyle.Render("Column Info: "+a.queryTableName) + "\n\n"

		if a.columnInfoSearching {
			content += inputLabelStyle.Render("Filter: ") + a.columnInfoSearchInput.View() + "\n\n"
		} else if a.columnInfoFilter != "" {
			content += dimStyle.Render("Filter: "+a.columnInfoFilter+" (esc to clear)") + "\n\n"
		} else {
			content += dimStyle.Render("Filter: press / to filter") + "\n\n"
		}

		if len(a.filteredColumnInfo) > 0 {
			content += a.columnInfoTable.View()
		} else {
			content += dimStyle.Render("No columns match filter")
		}

		return a.renderFrame(content, "j/k: navigate • /: filter • esc/?: close")
	}

	content = "Database: " + selectedStyle.Render(a.selectedDatabase) + "\n\n"

	if a.tableSearching {
		content += inputLabelStyle.Render("Search: ") + a.tableSearchInput.View() + "\n\n"
	} else if a.tableSearchInput.Value() != "" {
		content += dimStyle.Render("Search: "+a.tableSearchInput.Value()) + "\n\n"
	} else {
		content += dimStyle.Render("Search: press / to filter") + "\n\n"
	}

	content += "Tables:\n"
	if len(a.filteredTables) > 0 {
		var lines []string
		for i, tbl := range a.filteredTables {
			prefix := "  "
			line := tbl
			if i == a.tableCursor {
				prefix = "> "
				line = selectedStyle.Render(line)
			}
			lines = append(lines, prefix+line)
		}
		a.viewport.SetContent(strings.Join(lines, "\n"))
		a.syncViewportToCursor(a.tableCursor, len(a.filteredTables))
		content += a.viewport.View()
	} else {
		content += dimStyle.Render("No tables found")
	}

	controls := "j/k: navigate • /: search • ?: cols • enter: select • esc: back • q: quit"

	return a.renderFrame(content, controls)
}
