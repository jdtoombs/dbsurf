package app

import (
	"dbsurf/db"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
)

func filterStrings(items []string, query string) []string {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterColumnInfo(items []db.ColumnInfo, query string) []db.ColumnInfo {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	filtered := make([]db.ColumnInfo, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (a *App) filterAndRebuildColumnInfo() {
	a.filteredColumnInfo = filterColumnInfo(a.columnInfoData, a.columnInfoFilter)
	if len(a.filteredColumnInfo) > 0 {
		tableHeight := min(len(a.filteredColumnInfo), 15)
		a.columnInfoTable = buildColumnInfoTable(a.filteredColumnInfo, tableHeight)
	}
}

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

func buildColumnInfoTable(columns []db.ColumnInfo, height int) table.Model {
	cols := []table.Column{
		{Title: "Column", Width: 20},
		{Title: "Type", Width: 18},
		{Title: "Key", Width: 4},
		{Title: "Nullable", Width: 10},
		{Title: "Default", Width: 20},
	}

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

	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		Foreground(ColorPrimary).
		Padding(0, 1)
	s.Selected = s.Selected.
		Foreground(ColorWarning).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(ColorText).
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
