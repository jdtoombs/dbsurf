// constants.go defines version info, layout dimensions, and color palette
// used throughout the application.
package app

import "github.com/charmbracelet/lipgloss"

// Version is set at build time via ldflags
var Version = "dev"
const (
	DefaultVisibleRows = 10
	// Layout dimensions for viewport calculation
	LogoHeight       = 10 // Logo art lines + version + newline after
	ControlsHeight   = 2  // Controls line + spacing
	BoxPadding       = 4  // Border (2) + padding (2) vertical
	BoxHeaderPadding = 5  // Header lines inside box (title + search + label)
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
