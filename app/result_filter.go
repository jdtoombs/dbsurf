package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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

func (a *App) updateResultSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.resultSearching = false
		a.resultSearchInput.Blur()
		return a, nil
	case "enter":
		a.resultSearching = false
		a.resultSearchInput.Blur()
		return a, nil
	}
	var cmd tea.Cmd
	a.resultSearchInput, cmd = a.resultSearchInput.Update(msg)
	a.resultFilter = a.resultSearchInput.Value()
	a.filterResults()
	return a, cmd
}

func (a *App) startResultSearch() tea.Cmd {
	a.resultSearching = true
	a.resultSearchInput.Focus()
	return textinput.Blink
}

func (a *App) clearResultFilter() {
	a.resultFilter = ""
	a.resultSearchInput.SetValue("")
	a.filterResults()
}
