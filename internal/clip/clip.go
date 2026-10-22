// Package clip reads the system clipboard for the text fields, handing the
// text over as the paste a terminal sends, so a ctrl+v the terminal did not
// turn into one still pastes.
package clip

import (
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

// Paste is the command that reads the clipboard through read, the system's
// when read is nil, and answers with its text as a tea.PasteMsg; with
// nothing when it is empty or cannot be read.
func Paste(read func() (string, error)) tea.Cmd {
	if read == nil {
		read = clipboard.ReadAll
	}
	return func() tea.Msg {
		s, err := read()
		if err != nil || s == "" {
			return nil
		}
		return tea.PasteMsg{Content: s}
	}
}
