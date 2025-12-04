package main

import (
	"fmt"
	"os"
	"strings"

	"dbsurf/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = `
     _ _                  __
  __| | |__  ___ _   _ _ _/ _|
 / _` + "`" + ` | '_ \/ __| | | | '__| |_
| (_| | |_) \__ \ |_| | |  |  _|
 \__,_|_.__/|___/\__,_|_|  |_|
`

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")).
			Padding(1, 2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

type mode int

const (
	modeList mode = iota
	modeInput
)

type model struct {
	config    *config.Config
	cursor    int
	mode      mode
	connInput textinput.Model
	nameInput textinput.Model
	inputStep int // 0 = entering conn string, 1 = entering name
	err       error
	width     int
	height    int
}

func initialModel() model {
	cfg, _ := config.Load()

	ci := textinput.New()
	ci.Placeholder = "postgres://user:pass@host:5432/dbname"
	ci.Width = 50

	ni := textinput.New()
	ni.Placeholder = "My Database"
	ni.Width = 30

	return model{
		config:    cfg,
		mode:      modeList,
		connInput: ci,
		nameInput: ni,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.mode == modeList {
			return m.updateList(msg)
		}
		return m.updateInput(msg)
	}
	return m, nil
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.config.Connections)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "n":
		m.mode = modeInput
		m.inputStep = 0
		m.connInput.Focus()
		return m, textinput.Blink
	case "d":
		if len(m.config.Connections) > 0 {
			m.config.Connections = append(
				m.config.Connections[:m.cursor],
				m.config.Connections[m.cursor+1:]...,
			)
			m.config.Save()
			if m.cursor >= len(m.config.Connections) && m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return m, nil
}

func (m model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
		m.connInput.Reset()
		m.nameInput.Reset()
		return m, nil
	case "enter":
		if m.inputStep == 0 {
			m.inputStep = 1
			m.connInput.Blur()
			m.nameInput.Focus()
			return m, textinput.Blink
		}
		m.config.AddConnection(
			m.nameInput.Value(),
			m.connInput.Value(),
			"postgres",
		)
		m.config.Save()
		m.mode = modeList
		m.connInput.Reset()
		m.nameInput.Reset()
		return m, nil
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.inputStep == 0 {
		m.connInput, cmd = m.connInput.Update(msg)
	} else {
		m.nameInput, cmd = m.nameInput.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// ASCII art logo
	logoRendered := titleStyle.Render(logo)

	var content string
	var controls string

	if m.mode == modeInput {
		if m.inputStep == 0 {
			content = "Connection string:\n\n" + m.connInput.View()
		} else {
			content = "Name for this connection:\n\n" + m.nameInput.View()
		}
		controls = "enter: submit • esc: cancel"
	} else {
		if len(m.config.Connections) == 0 {
			content = dimStyle.Render("No saved connections")
		} else {
			var lines []string
			for i, conn := range m.config.Connections {
				cursor := "  "
				line := fmt.Sprintf("%s (%s)", conn.Name, conn.DBType)
				if i == m.cursor {
					cursor = "> "
					line = selectedStyle.Render(line)
				}
				lines = append(lines, cursor+line)
			}
			content = strings.Join(lines, "\n")
		}
		controls = "j/k: navigate • n: new • d: delete • q: quit"
	}

	boxedContent := boxStyle.Render(content)

	logoRendered = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logoRendered)
	boxedContent = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, boxedContent)
	controlsRendered := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, dimStyle.Render(controls))

	contentHeight := strings.Count(logoRendered, "\n") + strings.Count(boxedContent, "\n") + 3
	bottomPadding := m.height - contentHeight - 2
	if bottomPadding < 0 {
		bottomPadding = 0
	}

	return logoRendered + "\n" +
		boxedContent + "\n" +
		strings.Repeat("\n", bottomPadding) +
		controlsRendered
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
