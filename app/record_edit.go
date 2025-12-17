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
	b.WriteString(db.FormatTableName(tableName, a.dbType))
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

func (a *App) generateDeleteSQL(tableName string, row []string, pkColumns []string) string {
	var b strings.Builder
	b.WriteString("DELETE FROM ")
	b.WriteString(db.FormatTableName(tableName, a.dbType))
	b.WriteString(" WHERE ")

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

func (a *App) generateMultiDeleteSQL(tableName string, rows [][]string, pkColumns []string) string {
	if len(pkColumns) == 1 {
		// Simple IN clause for single PK
		pkCol := pkColumns[0]
		var pkIndex int
		for i, col := range a.queryResult.Columns {
			if col == pkCol {
				pkIndex = i
				break
			}
		}

		var values []string
		for _, row := range rows {
			if row[pkIndex] == "NULL" {
				continue // Skip NULL PKs
			}
			values = append(values, "'"+strings.ReplaceAll(row[pkIndex], "'", "''")+"'")
		}

		return fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)", db.FormatTableName(tableName, a.dbType), pkCol, strings.Join(values, ", "))
	}

	// Composite PK - use OR conditions
	var conditions []string
	for _, row := range rows {
		var parts []string
		for _, pkCol := range pkColumns {
			for i, col := range a.queryResult.Columns {
				if col == pkCol {
					if row[i] == "NULL" {
						parts = append(parts, col+" IS NULL")
					} else {
						parts = append(parts, col+" = '"+strings.ReplaceAll(row[i], "'", "''")+"'")
					}
					break
				}
			}
		}
		conditions = append(conditions, "("+strings.Join(parts, " AND ")+")")
	}

	return fmt.Sprintf("DELETE FROM %s WHERE %s", db.FormatTableName(tableName, a.dbType), strings.Join(conditions, " OR "))
}

func (a *App) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		fullQuery := db.PrependUseDatabase(a.pendingDeleteSQL, a.selectedDatabase, a.dbType)
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
				if a.resultCursor >= len(a.filteredResultRows) && a.resultCursor > 0 {
					a.resultCursor = len(a.filteredResultRows) - 1
				}
			}
		}
		a.deleteConfirming = false
		a.pendingDeleteSQL = ""
		a.fkDependencies = nil
		a.fkDependencyCounts = nil
		a.deletingMultiple = false
		return a, nil
	case "n", "N", "esc":
		a.deleteConfirming = false
		a.pendingDeleteSQL = ""
		a.fkDependencies = nil
		a.fkDependencyCounts = nil
		a.deletingMultiple = false
		return a, nil
	case "j", "down":
		if len(a.fkDependencies) > 0 {
			a.fkDepCursor = (a.fkDepCursor + 1) % len(a.fkDependencies)
		}
		return a, nil
	case "k", "up":
		if len(a.fkDependencies) > 0 {
			a.fkDepCursor = (a.fkDepCursor - 1 + len(a.fkDependencies)) % len(a.fkDependencies)
		}
		return a, nil
	case "enter":
		if len(a.fkDependencies) > 0 {
			a.queryFKDependency(a.fkDependencies[a.fkDepCursor])
		}
		return a, nil
	}
	return a, nil
}

func (a *App) startRecordDelete() error {
	if a.queryResult == nil || len(a.filteredResultRows) == 0 {
		return fmt.Errorf("no record selected")
	}
	if a.queryTableName == "" {
		return fmt.Errorf("could not determine table name from query")
	}
	if len(a.queryPKColumns) == 0 {
		return fmt.Errorf("table has no primary key or query contains JOIN")
	}

	// Check if multi-delete (more than 1 filtered row)
	if len(a.filteredResultRows) > 1 {
		a.deletingMultiple = true
		a.deleteRowCount = len(a.filteredResultRows)
		a.pendingDeleteSQL = a.generateMultiDeleteSQL(a.queryTableName, a.filteredResultRows, a.queryPKColumns)
	} else {
		a.deletingMultiple = false
		a.deleteRowCount = 1
		row := a.filteredResultRows[a.resultCursor]
		a.pendingDeleteSQL = a.generateDeleteSQL(a.queryTableName, row, a.queryPKColumns)
	}

	// Query FK dependencies
	a.fkDependencies = nil
	a.fkDependencyCounts = nil
	a.fkDepCursor = 0

	deps, err := db.GetReferencingFKs(a.db, a.selectedDatabase, a.queryTableName, a.dbType)
	if err == nil && len(deps) > 0 {
		a.fkDependencies = deps
		a.fkDependencyCounts = make(map[string]int)

		// Count referencing rows for each FK dependency
		for _, dep := range deps {
			count := a.countFKReferences(dep)
			a.fkDependencyCounts[dep.TableName+"."+dep.ColumnName] = count
		}
	}

	a.deleteConfirming = true
	return nil
}

func (a *App) countFKReferences(dep db.FKDependency) int {
	// Get the value of the referenced column from the current row(s)
	var refColIndex int
	for i, col := range a.queryResult.Columns {
		if col == dep.ReferencedColumn {
			refColIndex = i
			break
		}
	}

	// Collect unique values to check
	valueSet := make(map[string]bool)
	for _, row := range a.filteredResultRows {
		if refColIndex < len(row) && row[refColIndex] != "NULL" {
			valueSet[row[refColIndex]] = true
		}
	}

	if len(valueSet) == 0 {
		return 0
	}

	// Build query to count references
	var values []string
	for v := range valueSet {
		values = append(values, "'"+strings.ReplaceAll(v, "'", "''")+"'")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IN (%s)",
		dep.TableName, dep.ColumnName, strings.Join(values, ", "))
	query = db.PrependUseDatabase(query, a.selectedDatabase, a.dbType)

	result, err := db.RunQuery(a.db, query)
	if err != nil || len(result.Rows) == 0 {
		return 0
	}

	var count int
	fmt.Sscanf(result.Rows[0][0], "%d", &count)
	return count
}

func (a *App) queryFKDependency(dep db.FKDependency) {
	// Get the value of the referenced column from the current row(s)
	var refColIndex int
	for i, col := range a.queryResult.Columns {
		if col == dep.ReferencedColumn {
			refColIndex = i
			break
		}
	}

	// Collect unique values
	valueSet := make(map[string]bool)
	for _, row := range a.filteredResultRows {
		if refColIndex < len(row) && row[refColIndex] != "NULL" {
			valueSet[row[refColIndex]] = true
		}
	}

	if len(valueSet) == 0 {
		return
	}

	var values []string
	for v := range valueSet {
		values = append(values, "'"+strings.ReplaceAll(v, "'", "''")+"'")
	}

	// Generate and run SELECT query
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s IN (%s)",
		dep.TableName, dep.ColumnName, strings.Join(values, ", "))

	a.queryInput.SetValue(query)
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
		a.queryTableName = dep.TableName
		a.queryPKColumns, _ = db.GetPrimaryKey(a.db, a.selectedDatabase, dep.TableName, a.dbType)
	}

	// Exit delete confirmation
	a.deleteConfirming = false
	a.pendingDeleteSQL = ""
	a.fkDependencies = nil
	a.fkDependencyCounts = nil
}

func (a *App) viewDeleteConfirm() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("Confirm DELETE"))
	b.WriteString("\n\n")

	// Show row count info for multi-delete
	if a.deletingMultiple {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Deleting %d of %d filtered rows",
			a.deleteRowCount, a.deleteRowCount)))
		if len(a.queryResult.Rows) != a.deleteRowCount {
			b.WriteString(dimStyle.Render(fmt.Sprintf(" (from %d total)", len(a.queryResult.Rows))))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(dimStyle.Render(a.pendingDeleteSQL))
	b.WriteString("\n")

	// Show FK dependencies if any have referencing rows
	hasRefs := false
	for _, dep := range a.fkDependencies {
		count := a.fkDependencyCounts[dep.TableName+"."+dep.ColumnName]
		if count > 0 {
			hasRefs = true
			break
		}
	}

	if hasRefs {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Warning: This record is referenced by other tables:"))
		b.WriteString("\n")

		for i, dep := range a.fkDependencies {
			count := a.fkDependencyCounts[dep.TableName+"."+dep.ColumnName]
			if count == 0 {
				continue
			}

			if i == a.fkDepCursor {
				b.WriteString(editingStyle.Render(fmt.Sprintf("> %s: %d rows (via %s)",
					dep.TableName, count, dep.ColumnName)))
			} else {
				b.WriteString(dimStyle.Render(fmt.Sprintf("  %s: %d rows (via %s)",
					dep.TableName, count, dep.ColumnName)))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(dimStyle.Render("To delete, first remove rows from the tables above"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("Execute this query? ")
	b.WriteString(errorStyle.Render("(y/n)"))

	var controls string
	if hasRefs {
		controls = "y: execute • n/esc: cancel • j/k: navigate • enter: query dependency"
	} else {
		controls = "y: execute • n/esc: cancel"
	}
	return a.renderFrame(b.String(), controls)
}
