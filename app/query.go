package app

import (
	"dbsurf/db"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

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
	// Handle search mode
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
			} else {
				a.queryInput.Blur()
			}
			return a, nil
		}
	case "/":
		if !a.queryFocused && a.queryResult != nil {
			a.resultSearching = true
			a.resultSearchInput.Focus()
			return a, nil
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
				}
			}
			return a, nil
		}
		var cmd tea.Cmd
		a.queryInput, cmd = a.queryInput.Update(msg)
		return a, cmd
	} else {
		// Navigate results
		switch msg.String() {
		case "j", "down":
			if a.resultCursor < len(a.filteredResultRows)-1 {
				a.resultCursor++
			}
		case "k", "up":
			if a.resultCursor > 0 {
				a.resultCursor--
			}
		}
		return a, nil
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

	if a.resultSearching {
		content += "Filter: " + a.resultSearchInput.View() + "\n\n"
	} else if a.resultFilter != "" {
		content += dimStyle.Render("Filter: "+a.resultFilter+" (esc to clear)") + "\n\n"
	}

	if a.queryErr != nil {
		content += "Error: " + a.queryErr.Error()
	} else if a.queryResult != nil && len(a.filteredResultRows) > 0 {
		// Show current row in JSON-like format with highlighted keys
		row := a.filteredResultRows[a.resultCursor]
		content += selectedStyle.Render("{") + "\n"
		for j, col := range a.queryResult.Columns {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			comma := ","
			if j == len(a.queryResult.Columns)-1 {
				comma = ""
			}
			content += "  " + selectedStyle.Render(fmt.Sprintf(`"%s"`, col)) + ": " + valueStyle.Render(fmt.Sprintf(`"%s"`, val)) + comma + "\n"
		}
		content += selectedStyle.Render("}") + "\n"

		content += "\n" + dimStyle.Render(fmt.Sprintf("Row %d/%d", a.resultCursor+1, len(a.filteredResultRows)))
		if a.resultFilter != "" {
			content += dimStyle.Render(fmt.Sprintf(" (filtered from %d)", len(a.queryResult.Rows)))
		}
	} else if a.queryResult != nil {
		content += dimStyle.Render("No results")
	} else {
		content += dimStyle.Render("Enter a query and press enter")
	}

	controls := "enter: execute • tab: focus • /: filter • ctrl+t: tables • j/k: nav • esc: back"

	return a.renderFrame(content, controls)
}
