// Package clip reads the system clipboard for the text fields, handing the
// text over as the paste a terminal sends, so a ctrl+v the terminal did not
// turn into one still pastes.
package clip

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

// timeout bounds a read that never comes back, as Shell.Copy bounds a write.
// A variable so a test can shrink it.
var timeout = 5 * time.Second

// Paste is the command that reads the clipboard through read, the system's
// when read is nil, and answers with its text as a tea.PasteMsg; with
// nothing when it is empty, cannot be read or does not answer in time.
func Paste(read func() (string, error)) tea.Cmd {
	if read == nil {
		read = clipboard.ReadAll
	}
	limit := timeout // read here, on the loop, never from the command
	return func() tea.Msg {
		type answer struct {
			s   string
			err error
		}
		done := make(chan answer, 1) // buffered: a late read still ends its goroutine
		go func() {
			s, err := read()
			done <- answer{s, err}
		}()
		select {
		case a := <-done:
			if a.err != nil || a.s == "" {
				return nil
			}
			return tea.PasteMsg{Content: a.s}
		case <-time.After(limit):
			return nil
		}
	}
}
