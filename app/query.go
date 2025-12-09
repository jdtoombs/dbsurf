package app

import (
	"dbsurf/db"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (a *App) buildResultTable() {
	if a.queryResult == nil || len(a.queryResult.Columns) == 0 {
		return
	}

	boxWidth := a.width - 4
	if boxWidth > 80 {
		boxWidth = 80
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	// Calculate column widths
	numCols := len(a.queryResult.Columns)
	colWidth := (boxWidth - 8) / numCols
	if colWidth < 8 {
		colWidth = 8
	}
	if colWidth > 20 {
		colWidth = 20
	}

	cols := make([]table.Column, numCols)
	for i, c := range a.queryResult.Columns {
		cols[i] = table.Column{Title: c, Width: colWidth}
	}

	rows := make([]table.Row, len(a.queryResult.Rows))
	for i, r := range a.queryResult.Rows {
		rows[i] = table.Row(r)
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(!a.queryFocused),
		table.WithHeight(8),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("6")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("2")).
		Bold(true)
	t.SetStyles(s)

	a.resultTable = t
}

func (a *App) updateQuery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.mode = modeConnected
		return a, nil
	case "tab":
		if a.queryResult != nil && len(a.queryResult.Rows) > 0 {
			a.queryFocused = !a.queryFocused
			if a.queryFocused {
				a.queryInput.Focus()
				a.resultTable.Blur()
			} else {
				a.queryInput.Blur()
				a.resultTable.Focus()
			}
			return a, nil
		}
	}

	if a.queryFocused {
		switch msg.String() {
		case "enter":
			query := a.queryInput.Value()
			if query != "" {
				// Prepend USE for SQL Server to ensure correct db context
				fullQuery := query
				if a.dbType == "sqlserver" {
					fullQuery = fmt.Sprintf("USE [%s]; %s", a.selectedDatabase, query)
				}
				result, err := db.RunQuery(a.db, fullQuery)
				if err != nil {
					a.queryErr = err
					a.queryResult = nil
				} else {
					a.queryErr = nil
					a.queryResult = result
					a.buildResultTable()
				}
			}
			return a, nil
		}
		var cmd tea.Cmd
		a.queryInput, cmd = a.queryInput.Update(msg)
		return a, cmd
	} else {
		var cmd tea.Cmd
		a.resultTable, cmd = a.resultTable.Update(msg)
		return a, cmd
	}
}

func (a *App) viewQuery() string {
	var content string

	content = "Database: " + selectedStyle.Render(a.selectedDatabase) + "\n\n"

	if a.queryFocused {
		content += "Query: " + a.queryInput.View() + "\n\n"
	} else {
		content += dimStyle.Render("Query: "+a.queryInput.Value()) + "\n\n"
	}

	if a.queryErr != nil {
		content += "Error: " + a.queryErr.Error()
	} else if a.queryResult != nil {
		content += a.resultTable.View() + "\n\n"
		content += dimStyle.Render(fmt.Sprintf("%d rows", len(a.queryResult.Rows)))
	} else {
		content += dimStyle.Render("Enter a query and press enter")
	}

	controls := "enter: execute • tab: switch focus • esc: back • q: quit"

	return a.renderFrame(content, controls)
}
