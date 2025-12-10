package app

import (
	"dbsurf/db"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// filterStrings filters a slice of strings by a query (case-insensitive)
func filterStrings(items []string, query string) []string {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	// Pre-allocate to avoid reallocations during append
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// moveCursor adjusts a cursor within bounds, returning the new position
func moveCursor(cursor, delta, max int) int {
	cursor += delta
	if cursor < 0 {
		return 0
	}
	if cursor >= max {
		return max - 1
	}
	return cursor
}

// buildColumnInfoTable creates a bubbles table for column info display
func buildColumnInfoTable(columns []db.ColumnInfo, height int) table.Model {
	// Define table columns
	cols := []table.Column{
		{Title: "Column", Width: 20},
		{Title: "Type", Width: 18},
		{Title: "Key", Width: 4},
		{Title: "Nullable", Width: 10},
		{Title: "Default", Width: 20},
	}

	// Build rows
	rows := make([]table.Row, len(columns))
	for i, col := range columns {
		typeStr := col.DataType
		if col.MaxLength != "" {
			typeStr += fmt.Sprintf("(%s)", col.MaxLength)
		}

		keyStr := ""
		if col.IsPrimary {
			keyStr = "PK"
		}

		nullStr := "NULL"
		if !col.IsNullable {
			nullStr = "NOT NULL"
		}

		defaultStr := col.Default

		rows[i] = table.Row{col.Name, typeStr, keyStr, nullStr, defaultStr}
	}

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		Foreground(lipgloss.Color("6")).
		Padding(0, 1)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("3")).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("7")).
		Padding(0, 1)

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	t.SetStyles(s)

	return t
}
