// advanced_query.go handles launching an external editor ($EDITOR) for writing
// longer SQL queries. Creates a temp .sql file, opens the editor, and reads
// back the content when the editor closes.
package app

import (
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type editorFinishedMsg struct {
	err      error
	tempFile string
}

func (a *App) openAdvancedQueryEditor() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vim"
	}

	tmpFile, err := os.CreateTemp("", "dbsurf-query-*.sql")
	if err != nil {
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}

	currentQuery := a.queryInput.Value()
	if _, err := tmpFile.WriteString(currentQuery); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}
	tmpFile.Close()

	a.advancedQueryTempFile = tmpFile.Name()

	c := exec.Command(editor, tmpFile.Name())
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{
			err:      err,
			tempFile: tmpFile.Name(),
		}
	})
}

func (a *App) handleEditorFinished(msg editorFinishedMsg) (tea.Model, tea.Cmd) {
	defer func() {
		if msg.tempFile != "" {
			os.Remove(msg.tempFile)
		}
		a.advancedQueryTempFile = ""
	}()

	if msg.err != nil {
		a.queryErr = msg.err
		return a, nil
	}

	content, err := os.ReadFile(msg.tempFile)
	if err != nil {
		a.queryErr = err
		return a, nil
	}

	query := strings.TrimRight(string(content), "\n\r")
	a.queryInput.SetValue(query)
	a.queryInput.Focus()

	return a, textinput.Blink
}
