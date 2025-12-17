// query.go handles the query mode for executing SQL queries and navigating results.
// It supports running queries, viewing results in JSON format, inline editing,
// copying rows to clipboard, and navigating to column info.
package app

import (
	"dbsurf/db"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var valueStyle = lipgloss.NewStyle().Foreground(ColorText)

type clearCopyMsg struct{}

func parseTableName(query string) string {
	re := regexp.MustCompile(`(?i)\bFROM\s+(["\[\]?\w]+\.)?(["\[\]?\w]+)`)
	matches := re.FindStringSubmatch(query)
	if len(matches) >= 3 {
		schema := strings.Trim(matches[1], "\"[]`. ")
		table := strings.Trim(matches[2], "\"[]`")
		if schema != "" {
			return schema + "." + table
		}
		return table
	}
	return ""
}

func hasJoin(query string) bool {
	re := regexp.MustCompile(`(?i)\b(JOIN|,)\s+["\[\]?\w]+`)
	return re.MatchString(query)
}

func (a *App) updateQuery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	if a.editConfirming {
		return a.updateEditConfirm(msg)
	}

	if a.deleteConfirming {
		return a.updateDeleteConfirm(msg)
	}

	if a.fieldEditing {
		return a.updateFieldEdit(msg)
	}

	if a.resultSearching {
		return a.updateResultSearch(msg)
	}

	switch msg.String() {
	case "ctrl+d":
		if err := a.startRecordDelete(); err != nil {
			a.queryErr = err
		}
		return a, nil
	case "esc":
		if a.resultFilter != "" {
			a.clearResultFilter()
			return a, nil
		}
		a.mode = modeConnected
		return a, nil
	case "tab":
		if a.queryResult != nil && len(a.filteredResultRows) > 0 {
			a.queryFocused = !a.queryFocused
			if a.queryFocused {
				a.queryInput.Focus()
				return a, textinput.Blink
			} else {
				a.queryInput.Blur()
				a.fieldCursor = 0
			}
			return a, nil
		}
	case "/":
		if !a.queryFocused && a.queryResult != nil {
			return a, a.startResultSearch()
		}
	case "ctrl+t":
		tables, err := db.ListTables(a.db, a.selectedDatabase, a.dbType)
		if err != nil {
			a.queryErr = err
			return a, nil
		}
		a.tables = tables
		a.filteredTables = tables
		a.tableCursor = 0
		a.tableSearchInput.Reset()
		a.tableSearching = false
		a.mode = modeTableList
		return a, nil
	}

	if a.queryFocused {
		return a.updateQueryInput(msg)
	} else {
		return a.updateResultNavigation(msg)
	}
}

func (a *App) updateQueryInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		a.queryInput.SetValue("")
		return a, nil
	case "ctrl+e":
		return a, a.openAdvancedQueryEditor()
	case "enter":
		query := a.queryInput.Value()
		if query != "" {
			fullQuery := db.PrependUseDatabase(query, a.selectedDatabase, a.dbType)
			result, err := db.RunQuery(a.db, fullQuery)
			if err != nil {
				a.queryErr = err
				a.queryResult = nil
				a.filteredResultRows = nil
			} else {
				a.queryErr = nil
				a.queryResult = result
				a.resultFilter = ""
				a.resultSearchInput.SetValue("")
				a.filterResults()
				a.resultCursor = 0
				a.fieldCursor = 0
				a.queryTableName = parseTableName(query)
				a.queryPKColumns = nil
				if a.queryTableName != "" && !hasJoin(query) {
					var pkErr error
					a.queryPKColumns, pkErr = db.GetPrimaryKey(a.db, a.selectedDatabase, a.queryTableName, a.dbType)
					if pkErr != nil {
						// Show PK lookup error so user knows why edit/delete won't work
						a.queryErr = fmt.Errorf("PK lookup failed for %s: %v", a.queryTableName, pkErr)
					}
				}
			}
		}
		return a, nil
	}
	var cmd tea.Cmd
	a.queryInput, cmd = a.queryInput.Update(msg)
	return a, cmd
}

func (a *App) updateResultNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left":
		a.resultCursor = moveCursor(a.resultCursor, -1, len(a.filteredResultRows))
	case "l", "right":
		a.resultCursor = moveCursor(a.resultCursor, 1, len(a.filteredResultRows))
	case "k", "up":
		if a.queryResult != nil && len(a.queryResult.Columns) > 0 {
			a.fieldCursor = moveCursor(a.fieldCursor, -1, len(a.queryResult.Columns))
		}
	case "j", "down":
		if a.queryResult != nil && len(a.queryResult.Columns) > 0 {
			a.fieldCursor = moveCursor(a.fieldCursor, 1, len(a.queryResult.Columns))
		}
	case "i":
		if cmd := a.startFieldEdit(); cmd != nil {
			return a, cmd
		}
	case "ctrl+c":
		if a.queryResult != nil && len(a.filteredResultRows) > 0 {
			row := a.filteredResultRows[a.resultCursor]
			var jsonParts []string
			for i, col := range a.queryResult.Columns {
				val := ""
				if i < len(row) {
					val = row[i]
				}
				jsonParts = append(jsonParts, fmt.Sprintf(`"%s": "%s"`, col, val))
			}
			jsonStr := "{" + strings.Join(jsonParts, ", ") + "}"
			if err := clipboard.WriteAll(jsonStr); err == nil {
				a.copySuccess = true
				return a, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
					return clearCopyMsg{}
				})
			}
		}
	case "?":
		if a.queryTableName != "" {
			info, err := db.GetColumnInfo(a.db, a.selectedDatabase, a.queryTableName, a.dbType)
			if err == nil && len(info) > 0 {
				a.columnInfoData = info
				a.columnInfoFilter = ""
				a.columnInfoSearchInput.SetValue("")
				a.filteredColumnInfo = info
				tableHeight := min(len(info), 15)
				a.columnInfoTable = buildColumnInfoTable(info, tableHeight)
				if a.queryResult != nil && a.fieldCursor < len(a.queryResult.Columns) {
					selectedCol := a.queryResult.Columns[a.fieldCursor]
					for i, col := range info {
						if col.Name == selectedCol {
							a.columnInfoTable.SetCursor(i)
							break
						}
					}
				}
				a.showingColumnInfo = true
			}
		}
	}
	return a, nil
}

func (a *App) viewQuery() string {
	var b strings.Builder

	if a.showingColumnInfo {
		b.WriteString(selectedStyle.Render("Column Info: " + a.queryTableName))
		b.WriteString("\n\n")

		if a.columnInfoSearching {
			b.WriteString(inputLabelStyle.Render("Filter: "))
			b.WriteString(a.columnInfoSearchInput.View())
			b.WriteString("\n\n")
		} else if a.columnInfoFilter != "" {
			b.WriteString(dimStyle.Render("Filter: " + a.columnInfoFilter + " (esc to clear)"))
			b.WriteString("\n\n")
		} else {
			b.WriteString(dimStyle.Render("Filter: press / to filter"))
			b.WriteString("\n\n")
		}

		if len(a.filteredColumnInfo) > 0 {
			b.WriteString(a.columnInfoTable.View())
		} else {
			b.WriteString(dimStyle.Render("No columns match filter"))
		}

		controls := "j/k: navigate • /: filter • esc/?: close"
		return a.renderFrame(b.String(), controls)
	}

	if a.editConfirming {
		return a.viewEditConfirm()
	}

	if a.deleteConfirming {
		return a.viewDeleteConfirm()
	}

	b.WriteString("Database: ")
	b.WriteString(selectedStyle.Render(a.selectedDatabase))
	b.WriteString("\n\n")

	if a.queryFocused {
		b.WriteString(inputLabelStyle.Render("> Query: "))
		b.WriteString(focusedInputStyle.Render(a.queryInput.View()))
		b.WriteString("\n\n")
	} else {
		b.WriteString(dimStyle.Render("  Query: " + a.queryInput.Value()))
		b.WriteString("\n\n")
	}

	if a.queryResult != nil {
		if a.resultSearching {
			b.WriteString(inputLabelStyle.Render("Filter: "))
			b.WriteString(a.resultSearchInput.View())
			b.WriteString("\n\n")
		} else if a.resultFilter != "" {
			b.WriteString(dimStyle.Render("Filter: " + a.resultFilter + " (esc to clear)"))
			b.WriteString("\n\n")
		} else if !a.queryFocused {
			b.WriteString(dimStyle.Render("Filter: press / to filter"))
			b.WriteString("\n\n")
		} else {
			b.WriteString(dimStyle.Render("Filter: focus results to filter"))
			b.WriteString("\n\n")
		}
	}

	if a.queryErr != nil {
		b.WriteString("Error: ")
		b.WriteString(a.queryErr.Error())
	} else if a.queryResult != nil && len(a.filteredResultRows) > 0 {
		row := a.filteredResultRows[a.resultCursor]
		bracketStyle := dimStyle
		if !a.queryFocused {
			bracketStyle = bracketFocusedStyle
		}
		b.WriteString(bracketStyle.Render("{"))
		b.WriteString("\n")

		var fieldLines []string
		for j, col := range a.queryResult.Columns {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			comma := ","
			if j == len(a.queryResult.Columns)-1 {
				comma = ""
			}

			var line strings.Builder
			line.WriteString("  ")
			isSelected := !a.queryFocused && j == a.fieldCursor
			if isSelected {
				line.WriteString(editingStyle.Render(fmt.Sprintf(`"%s"`, col)))
				line.WriteString(": ")
				if a.fieldEditing {
					line.WriteString(a.fieldEditInput.View())
				} else {
					line.WriteString(editingStyle.Render(fmt.Sprintf(`"%s"`, val)))
				}
			} else {
				line.WriteString(selectedStyle.Render(fmt.Sprintf(`"%s"`, col)))
				line.WriteString(": ")
				line.WriteString(valueStyle.Render(fmt.Sprintf(`"%s"`, val)))
			}
			line.WriteString(comma)
			fieldLines = append(fieldLines, line.String())
		}

		a.viewport.SetContent(strings.Join(fieldLines, "\n"))
		a.syncViewportToCursor(a.fieldCursor, len(a.queryResult.Columns))
		b.WriteString(strings.TrimRight(a.viewport.View(), "\n "))
		b.WriteString("\n")
		b.WriteString(bracketStyle.Render("}"))
		b.WriteString("\n")

		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Row "))
		b.WriteString(selectedStyle.Render(fmt.Sprintf("%d/%d", a.resultCursor+1, len(a.filteredResultRows))))
		if a.resultFilter != "" {
			b.WriteString(dimStyle.Render(" (filtered from "))
			b.WriteString(selectedStyle.Render(fmt.Sprintf("%d", len(a.queryResult.Rows))))
			b.WriteString(dimStyle.Render(")"))
		}
		if a.copySuccess {
			b.WriteString(selectedStyle.Render(" Copied!"))
		}
	} else if a.queryResult != nil {
		b.WriteString(dimStyle.Render("No results"))
	} else {
		b.WriteString(dimStyle.Render("Enter a query and press enter"))
	}

	var controls string
	if a.fieldEditing {
		controls = "enter: save • esc: cancel"
	} else if !a.queryFocused && a.queryResult != nil && len(a.filteredResultRows) > 0 {
		controls = "h/l: rows • j/k: fields • i: edit • ctrl+d: delete • ?: cols • ctrl+c: copy • /: filter • tab: query • esc: back"
	} else {
		controls = "enter: run • ctrl+e: editor • tab: results • ctrl+c: clear • ctrl+d: delete • ctrl+t: tables • esc: back"
	}

	return a.renderFrame(b.String(), controls)
}
