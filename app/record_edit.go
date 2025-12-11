package app

import (
	"dbsurf/db"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var editingStyle = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)

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

func (a *App) updateEditConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		fullQuery := db.PrependUseDatabase(a.pendingUpdateSQL, a.selectedDatabase, a.dbType)
		_, err := db.RunQuery(a.db, fullQuery)
		if err != nil {
			a.queryErr = err
		} else {
			query := db.PrependUseDatabase(a.queryInput.Value(), a.selectedDatabase, a.dbType)
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

func (a *App) updateFieldEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (a *App) startFieldEdit() tea.Cmd {
	if a.queryResult != nil && len(a.filteredResultRows) > 0 {
		row := a.filteredResultRows[a.resultCursor]
		if a.fieldCursor < len(row) {
			a.fieldOriginalValue = row[a.fieldCursor]
			a.fieldEditInput.SetValue(a.fieldOriginalValue)
			a.fieldEditInput.Focus()
			a.fieldEditing = true
			return textinput.Blink
		}
	}
	return nil
}

func (a *App) viewEditConfirm() string {
	var b strings.Builder
	b.WriteString(selectedStyle.Render("Confirm UPDATE"))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render(a.pendingUpdateSQL))
	b.WriteString("\n\n")
	b.WriteString("Execute this query? ")
	b.WriteString(selectedStyle.Render("(y/n)"))
	controls := "y: execute • n/esc: cancel"
	return a.renderFrame(b.String(), controls)
}
