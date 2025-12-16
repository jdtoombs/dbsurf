package app

import (
	"database/sql"
	"dbsurf/config"
	"dbsurf/db"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

const logoWidth = 56

var logoArt = `
       ▄▄  ▄▄                                         ▄▄▄▄
       ██  ██                                        ██▀▀▀
  ▄███▄██  ██▄███▄   ▄▄█████▄  ██    ██   ██▄████  ███████
 ██▀  ▀██  ██▀  ▀██  ██▄▄▄▄ ▀  ██    ██   ██▀        ██
 ██    ██  ██    ██   ▀▀▀▀██▄  ██    ██   ██         ██
 ▀██▄▄███  ███▄▄██▀  █▄▄▄▄▄██  ██▄▄▄███   ██         ██
   ▀▀▀ ▀▀  ▀▀ ▀▀▀     ▀▀▀▀▀▀    ▀▀▀▀ ▀▀   ▀▀         ▀▀
`

func centeredVersion() string {
	padding := (logoWidth - len(Version)) / 2
	if padding < 0 {
		padding = 0
	}
	return strings.Repeat(" ", padding) + Version
}

var Logo = logoArt + centeredVersion()

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	errorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	bracketFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)
)

type mode int

const (
	modeList mode = iota
	modeInput
	modeConnected
	modeQuery
	modeTableList
)

type App struct {
	config            *config.Config
	cursor            int
	mode              mode
	connInput         textinput.Model
	nameInput         textinput.Model
	inputStep         int
	inputErr          error
	inputTesting      bool
	inputSpinner      spinner.Model
	err               error
	width             int
	height            int
	db                *sql.DB
	dbErr             error
	dbType            string
	databases         []string
	filteredDatabases []string
	dbCursor          int
	dbSearching       bool
	dbSearchInput     textinput.Model
	// Query mode
	selectedDatabase   string
	queryInput         textinput.Model
	queryResult        *db.QueryResult
	queryErr           error
	queryFocused       bool
	resultCursor       int
	resultSearching    bool
	resultSearchInput  textinput.Model
	resultFilter       string
	filteredResultRows [][]string
	// Field editing
	fieldCursor        int
	fieldEditing       bool
	fieldEditInput     textinput.Model
	fieldOriginalValue string
	editConfirming     bool
	pendingUpdateSQL   string
	queryTableName        string
	queryPKColumns        []string
	advancedQueryTempFile string
	// Column info mode
	showingColumnInfo       bool
	columnInfoTable         table.Model
	columnInfoData          []db.ColumnInfo
	filteredColumnInfo      []db.ColumnInfo
	columnInfoSearching     bool
	columnInfoSearchInput   textinput.Model
	columnInfoFilter        string
	// Table list mode
	tables           []string
	filteredTables   []string
	tableCursor      int
	tableSearching   bool
	tableSearchInput textinput.Model
	// Copy mode
	copySuccess bool
	// Viewport for scrollable content
	viewport viewport.Model
}

func New() *App {
	cfg, _ := config.Load()

	ci := textinput.New()
	ci.Placeholder = "postgres://user:pass@host:5432/dbname"
	ci.Width = 50

	ni := textinput.New()
	ni.Placeholder = "My Database"
	ni.Width = 30

	si := textinput.New()
	si.Placeholder = "Search..."
	si.Width = 30

	editorName := "vim"
	if strings.Contains(os.Getenv("EDITOR"), "nvim") {
		editorName = "nvim"
	}

	qi := textinput.New()
	qi.Placeholder = "enter query here or ctrl+e for " + editorName
	qi.Width = 60

	ri := textinput.New()
	ri.Placeholder = "Filter results..."
	ri.Width = 30

	ti := textinput.New()
	ti.Placeholder = "Filter tables..."
	ti.Width = 30

	fi := textinput.New()
	fi.Placeholder = ""
	fi.Width = 50

	cfi := textinput.New()
	cfi.Placeholder = "Filter columns..."
	cfi.Width = 30

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	// Initialize viewport with disabled keybindings (we control scrolling programmatically)
	vp := viewport.New(80, 10)
	vp.KeyMap = viewport.KeyMap{}
	vp.MouseWheelEnabled = true

	return &App{
		config:                cfg,
		mode:                  modeList,
		connInput:             ci,
		nameInput:             ni,
		inputSpinner:          sp,
		dbSearchInput:         si,
		queryInput:            qi,
		queryFocused:          true,
		resultSearchInput:     ri,
		tableSearchInput:      ti,
		fieldEditInput:        fi,
		columnInfoSearchInput: cfi,
		viewport:              vp,
	}
}
