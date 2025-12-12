package app

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateViewportSize()
		return a, nil
	case clearCopyMsg:
		a.copySuccess = false
		return a, nil
	case spinner.TickMsg:
		if a.inputTesting {
			var cmd tea.Cmd
			a.inputSpinner, cmd = a.inputSpinner.Update(msg)
			return a, cmd
		}
		return a, nil
	case connectionTestMsg:
		return a.handleConnectionTestResult(msg)
	case tea.KeyMsg:
		if msg.String() == "q" {
			if a.db != nil {
				a.db.Close()
			}
			return a, tea.Quit
		}
		switch a.mode {
		case modeList:
			return a.updateList(msg)
		case modeInput:
			return a.updateInput(msg)
		case modeConnected:
			return a.updateConnected(msg)
		case modeQuery:
			return a.updateQuery(msg)
		case modeTableList:
			return a.updateTableList(msg)
		}
	}
	return a, nil
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	switch a.mode {
	case modeList:
		return a.viewList()
	case modeInput:
		return a.viewInput()
	case modeConnected:
		return a.viewConnected()
	case modeQuery:
		return a.viewQuery()
	case modeTableList:
		return a.viewTableList()
	default:
		return a.viewList()
	}
}

func (a *App) renderFrame(content, controls string) string {
	logoRendered := titleStyle.Render(Logo)

	boxWidth := a.width - 4
	if boxWidth > 80 {
		boxWidth = 80
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	box := boxStyle.Width(boxWidth - 4)
	boxedContent := box.Render(content)

	logoRendered = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, logoRendered)
	controlsRendered := lipgloss.PlaceHorizontal(a.width, lipgloss.Center, dimStyle.Render(controls))
	boxedContent = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, boxedContent)

	// Build frame with logo and box
	frame := logoRendered + "\n" + boxedContent

	// Place frame at top and controls at bottom of terminal
	frameHeight := a.height - 1 // Leave room for controls
	frame = lipgloss.Place(a.width, frameHeight, lipgloss.Center, lipgloss.Top, frame)

	return frame + "\n" + controlsRendered
}

// updateViewportSize recalculates viewport dimensions based on terminal size
func (a *App) updateViewportSize() {
	boxWidth := a.width - 4
	if boxWidth > 80 {
		boxWidth = 80
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	// Calculate available height for viewport content
	// Terminal height - logo - controls - box padding - header content inside box
	contentHeight := a.height - LogoHeight - ControlsHeight - BoxPadding - BoxHeaderPadding
	if contentHeight < 5 {
		contentHeight = 5
	}

	a.viewport.Width = boxWidth - 4 // account for box border/padding
	a.viewport.Height = contentHeight
}

// syncViewportToCursor adjusts viewport scroll position to keep cursor visible
func (a *App) syncViewportToCursor(cursor, totalItems int) {
	if totalItems == 0 {
		return
	}
	if cursor < a.viewport.YOffset {
		a.viewport.SetYOffset(cursor)
	} else if cursor >= a.viewport.YOffset+a.viewport.Height {
		a.viewport.SetYOffset(cursor - a.viewport.Height + 1)
	}
}
