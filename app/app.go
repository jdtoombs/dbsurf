package app

import (
	"dbsurf/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const Logo = `
     _ _                  __
  __| | |__  ___ _   _ _ _/ _|
 / ` + "`" + `| '_ \/ __| | | | '__| |_
| (_| | |_) \__ \ |_| | |  |  _|
 \__,_|_.__/|___/\__,_|_|  |_|
`

// Styles
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

// Modes
type mode int

const (
	modeList mode = iota
	modeInput
)

// App is the main application state
type App struct {
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

func New() App {
	cfg, _ := config.Load()

	ci := textinput.New()
	ci.Placeholder = "postgres://user:pass@host:5432/dbname"
	ci.Width = 50

	ni := textinput.New()
	ni.Placeholder = "My Database"
	ni.Width = 30

	return App{
		config:    cfg,
		mode:      modeList,
		connInput: ci,
		nameInput: ni,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil
	case tea.KeyMsg:
		switch a.mode {
		case modeList:
			return a.updateList(msg)
		case modeInput:
			return a.updateInput(msg)
		}
	}
	return a, nil
}

func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	switch a.mode {
	case modeList:
		return a.viewList()
	case modeInput:
		return a.viewInput()
	default:
		return a.viewList()
	}
}
