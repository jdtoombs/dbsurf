package app

import "github.com/charmbracelet/lipgloss"

// Version is set at build time via ldflags
var Version = "dev"

// UI constants
const (
	DefaultVisibleRows = 10
)

// Color palette
const (
	ColorPrimary   = lipgloss.Color("6") // cyan
	ColorSuccess   = lipgloss.Color("2") // green
	ColorWarning   = lipgloss.Color("3") // yellow
	ColorError     = lipgloss.Color("1") // red
	ColorMuted     = lipgloss.Color("8") // gray
	ColorText      = lipgloss.Color("7") // white
)
