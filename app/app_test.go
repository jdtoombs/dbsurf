package app

import (
	"dbsurf/config"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	app := New()

	if app == nil {
		t.Fatal("New() returned nil")
	}
	if app.mode != modeList {
		t.Errorf("New() mode = %v, want modeList", app.mode)
	}
	if app.cursor != 0 {
		t.Errorf("New() cursor = %d, want 0", app.cursor)
	}
	if app.config == nil {
		t.Error("New() config is nil")
	}
	if app.queryFocused != true {
		t.Error("New() queryFocused should default to true")
	}
}

func TestMoveCursor(t *testing.T) {
	tests := []struct {
		name     string
		current  int
		delta    int
		max      int
		expected int
	}{
		{"move down from start", 0, 1, 5, 1},
		{"move down middle", 2, 1, 5, 3},
		{"move down at end stays", 4, 1, 5, 4},
		{"move up from middle", 2, -1, 5, 1},
		{"move up at start stays", 0, -1, 5, 0},
		{"empty list", 0, 1, 0, -1}, // edge case: max-1 when max=0
		{"single item down", 0, 1, 1, 0},
		{"single item up", 0, -1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := moveCursor(tt.current, tt.delta, tt.max)
			if got != tt.expected {
				t.Errorf("moveCursor(%d, %d, %d) = %d, want %d",
					tt.current, tt.delta, tt.max, got, tt.expected)
			}
		})
	}
}

func TestModeTransition_ListToInput(t *testing.T) {
	app := New()
	app.config = &config.Config{Connections: []config.Connection{}}
	app.mode = modeList

	// Simulate pressing 'n' to add new connection
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	model, _ := app.updateList(msg)
	updated := model.(*App)

	if updated.mode != modeInput {
		t.Errorf("After 'n' key, mode = %v, want modeInput", updated.mode)
	}
	if updated.inputStep != 0 {
		t.Errorf("After 'n' key, inputStep = %d, want 0", updated.inputStep)
	}
}

func TestModeTransition_InputEscape(t *testing.T) {
	app := New()
	app.mode = modeInput
	app.inputStep = 0

	// Simulate pressing 'esc' to go back
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	model, _ := app.updateInput(msg)
	updated := model.(*App)

	if updated.mode != modeList {
		t.Errorf("After 'esc' key in input mode, mode = %v, want modeList", updated.mode)
	}
}


func TestFilterStrings(t *testing.T) {
	items := []string{"users", "orders", "user_sessions", "products"}

	tests := []struct {
		filter   string
		expected int
	}{
		{"", 4},
		{"user", 2},     // "users" and "user_sessions"
		{"orders", 1},
		{"xyz", 0},
	}

	for _, tt := range tests {
		filtered := filterStrings(items, tt.filter)
		if len(filtered) != tt.expected {
			t.Errorf("filterStrings with filter %q = %d results, want %d",
				tt.filter, len(filtered), tt.expected)
		}
	}
}

func TestApp_Init(t *testing.T) {
	app := New()
	cmd := app.Init()

	// Init should return nil (no initial command)
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestApp_WindowSizeMsg(t *testing.T) {
	app := New()

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.width != 120 {
		t.Errorf("After WindowSizeMsg, width = %d, want 120", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("After WindowSizeMsg, height = %d, want 40", updated.height)
	}
}
