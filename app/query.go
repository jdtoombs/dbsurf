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

var valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
var editingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)

type clearCopyMsg struct{}

func parseTableName(query string) string {
	re := regexp.MustCompile(`(?i)\bFROM\s+(["\[\]?\w]+\.)?(["\[\]?\w]+)`)
	matches := re.FindStringSubmatch(query)
	if len(matches) >= 3 {
		return strings.Trim(matches[2], "\"[]`")
	}
	return ""
}

func hasJoin(query string) bool {
	re := regexp.MustCompile(`(?i)\b(JOIN|,)\s+["\[\]?\w]+`)
	return re.MatchString(query)
}

func (a *App) generateUpdateSQL(tableName string, row []string, colIndex int, newValue string, pkColumns []string) string {
	var b strings.Builder
	b.WriteString("UPDATE ")
	b.WriteString(tableName)
	b.WriteString(" SET ")
	b.WriteString(a.queryResult.Columns[colIndex])
	b.WriteString(" = '")
	b.WriteString(strings.ReplaceAll(newValue, "'", "''"))
	b.WriteString("' WHERE ")

	first := true
	for _, pkCol := range pkColumns {
		for i, col := range a.queryResult.Columns {
			if col == pkCol {
				if !first {
					b.WriteString(" AND ")
				}
				first = false
				b.WriteString(col)
				if row[i] == "NULL" {
					b.WriteString(" IS NULL")
				} else {
					b.WriteString(" = '")
					b.WriteString(strings.ReplaceAll(row[i], "'", "''"))
					b.WriteString("'")
				}
				break
			}
		}
	}

	return b.String()
}

func (a *App) filterResults() {
	if a.queryResult == nil {
		a.filteredResultRows = nil
		return
	}

	if a.resultFilter == "" {
		a.filteredResultRows = a.queryResult.Rows
		return
	}

	filter := strings.ToLower(a.resultFilter)
	filtered := [][]string{}
	for _, row := range a.queryResult.Rows {
		for _, col := range row {
			if strings.Contains(strings.ToLower(col), filter) {
				filtered = append(filtered, row)
				break
			}
		}
	}
	a.filteredResultRows = filtered
	if a.resultCursor >= len(a.filteredResultRows) {
		a.resultCursor = max(0, len(a.filteredResultRows)-1)
	}
}

func (a *App) updateQuery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.showingColumnInfo {
		switch msg.String() {
		case "esc", "?", "q":
			a.showingColumnInfo = false
			return a, nil
		}
		var cmd tea.Cmd
		a.columnInfoTable, cmd = a.columnInfoTable.Update(msg)
		return a, cmd
	}

	if a.editConfirming {
		switch msg.String() {
		case "y", "Y":
			fullQuery := a.pendingUpdateSQL
			if a.dbType == "sqlserver" {
				fullQuery = fmt.Sprintf("USE [%s]; %s", a.selectedDatabase, a.pendingUpdateSQL)
			}
			_, err := db.RunQuery(a.db, fullQuery)
			if err != nil {
				a.queryErr = err
			} else {
				query := a.queryInput.Value()
				if a.dbType == "sqlserver" {
					query = fmt.Sprintf("USE [%s]; %s", a.selectedDatabase, a.queryInput.Value())
				}
				result, err := db.RunQuery(a.db, query)
				if err != nil {
					a.queryErr = err
				} else {
					a.queryErr = nil
					a.queryResult = result
					a.filterResults()
				}
			}
			a.editConfirming = false
			a.pendingUpdateSQL = ""
			return a, nil
		case "n", "N", "esc":
			a.editConfirming = false
			a.pendingUpdateSQL = ""
			return a, nil
		}
		return a, nil
	}

	if a.fieldEditing {
		switch msg.String() {
		case "ctrl+c":
			a.fieldEditInput.SetValue("")
			return a, nil
		case "esc":
			a.fieldEditing = false
			a.fieldEditInput.Blur()
			return a, nil
		case "enter":
			newValue := a.fieldEditInput.Value()
			if newValue != a.fieldOriginalValue {
				if a.queryTableName == "" {
					a.queryErr = fmt.Errorf("could not determine table name from query")
					a.fieldEditing = false
					a.fieldEditInput.Blur()
					return a, nil
				}
				if len(a.queryPKColumns) == 0 {
					a.queryErr = fmt.Errorf("table has no primary key or query contains JOIN")
					a.fieldEditing = false
					a.fieldEditInput.Blur()
					return a, nil
				}
				row := a.filteredResultRows[a.resultCursor]
				a.pendingUpdateSQL = a.generateUpdateSQL(a.queryTableName, row, a.fieldCursor, newValue, a.queryPKColumns)
				a.editConfirming = true
			}
			a.fieldEditing = false
			a.fieldEditInput.Blur()
			return a, nil
		}
		var cmd tea.Cmd
		a.fieldEditInput, cmd = a.fieldEditInput.Update(msg)
		return a, cmd
	}

	if a.resultSearching {
		switch msg.String() {
		case "esc":
			a.resultSearching = false
			a.resultSearchInput.Blur()
			return a, nil
		case "enter":
			a.resultFilter = a.resultSearchInput.Value()
			a.filterResults()
			a.resultSearching = false
			a.resultSearchInput.Blur()
			return a, nil
		}
		var cmd tea.Cmd
		a.resultSearchInput, cmd = a.resultSearchInput.Update(msg)
		return a, cmd
	}

	switch msg.String() {
	case "esc":
		if a.resultFilter != "" {
			a.resultFilter = ""
			a.resultSearchInput.SetValue("")
			a.filterResults()
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
			a.resultSearching = true
			a.resultSearchInput.Focus()
			return a, textinput.Blink
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
		switch msg.String() {
		case "ctrl+c":
			a.queryInput.SetValue("")
			return a, nil
		case "enter":
			query := a.queryInput.Value()
			if query != "" {
				fullQuery := query
				if a.dbType == "sqlserver" {
					fullQuery = fmt.Sprintf("USE [%s]; %s", a.selectedDatabase, query)
				}
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
						a.queryPKColumns, _ = db.GetPrimaryKey(a.db, a.selectedDatabase, a.queryTableName, a.dbType)
					}
				}
			}
			return a, nil
		}
		var cmd tea.Cmd
		a.queryInput, cmd = a.queryInput.Update(msg)
		return a, cmd
	} else {
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
			if a.queryResult != nil && len(a.filteredResultRows) > 0 {
				row := a.filteredResultRows[a.resultCursor]
				if a.fieldCursor < len(row) {
					a.fieldOriginalValue = row[a.fieldCursor]
					a.fieldEditInput.SetValue(a.fieldOriginalValue)
					a.fieldEditInput.Focus()
					a.fieldEditing = true
					return a, textinput.Blink
				}
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
}

func (a *App) viewQuery() string {
	var b strings.Builder

	if a.showingColumnInfo {
		b.WriteString(selectedStyle.Render("Column Info: " + a.queryTableName))
		b.WriteString("\n\n")
		b.WriteString(a.columnInfoTable.View())
		return a.renderFrame(b.String(), "j/k: navigate • esc/?: close")
	}

	if a.editConfirming {
		b.WriteString(selectedStyle.Render("Confirm UPDATE"))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render(a.pendingUpdateSQL))
		b.WriteString("\n\n")
		b.WriteString("Execute this query? ")
		b.WriteString(selectedStyle.Render("(y/n)"))
		controls := "y: execute • n/esc: cancel"
		return a.renderFrame(b.String(), controls)
	}

	b.WriteString("Database: ")
	b.WriteString(selectedStyle.Render(a.selectedDatabase))
	b.WriteString("\n\n")

	if a.queryFocused {
		b.WriteString("Query: ")
		b.WriteString(a.queryInput.View())
		b.WriteString("\n\n")
	} else {
		b.WriteString(dimStyle.Render("Query: " + a.queryInput.Value()))
		b.WriteString("\n\n")
	}

	if a.resultSearching {
		b.WriteString("Filter: ")
		b.WriteString(a.resultSearchInput.View())
		b.WriteString("\n\n")
	} else if a.resultFilter != "" {
		b.WriteString(dimStyle.Render("Filter: " + a.resultFilter + " (esc to clear)"))
		b.WriteString("\n\n")
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
		for j, col := range a.queryResult.Columns {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			comma := ","
			if j == len(a.queryResult.Columns)-1 {
				comma = ""
			}
			b.WriteString("  ")

			isSelected := !a.queryFocused && j == a.fieldCursor
			if isSelected {
				b.WriteString(editingStyle.Render(fmt.Sprintf(`"%s"`, col)))
				b.WriteString(": ")
				if a.fieldEditing {
					b.WriteString(a.fieldEditInput.View())
				} else {
					b.WriteString(editingStyle.Render(fmt.Sprintf(`"%s"`, val)))
				}
			} else {
				b.WriteString(selectedStyle.Render(fmt.Sprintf(`"%s"`, col)))
				b.WriteString(": ")
				b.WriteString(valueStyle.Render(fmt.Sprintf(`"%s"`, val)))
			}
			b.WriteString(comma)
			b.WriteString("\n")
		}
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
		controls = "h/l: rows • j/k: fields • i: edit • ?: cols • ctrl+c: copy • /: filter • tab: query • esc: back"
	} else {
		controls = "enter: execute • tab: focus • /: filter • ctrl+c: clear • ctrl+t: tables • j/k: nav • esc: back"
	}

	return a.renderFrame(b.String(), controls)
}
