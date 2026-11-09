// Package clip reads and writes the system clipboard for the text fields,
// handing a read over as the paste a terminal sends, so a ctrl+v the
// terminal did not turn into one still pastes.
package clip

import (
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

// Timeout bounds a read or write that never comes back, when the caller
// gives no limit of its own.
const Timeout = 5 * time.Second

// Paste is the command that reads the clipboard through read, the system's
// when nil, and answers with its text as a tea.PasteMsg; with nothing when it
// is empty, cannot be read or takes longer than limit, Timeout when zero.
func Paste(read func() (string, error), limit time.Duration) tea.Cmd {
	if read == nil {
		read = clipboard.ReadAll
	}
	if limit <= 0 {
		limit = Timeout
	}
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

// ErrTimeout is a write the clipboard never answered.
var ErrTimeout = errors.New("the clipboard did not answer")

// Copy is the command that writes s through write, the system's when nil,
// and answers with done of how it went: nil, the write's error, or ErrTimeout
// after limit, Timeout when zero.
func Copy(write func(string) error, s string, limit time.Duration, done func(error) tea.Msg) tea.Cmd {
	if write == nil {
		write = clipboard.WriteAll
	}
	if limit <= 0 {
		limit = Timeout
	}
	return func() tea.Msg {
		errc := make(chan error, 1) // buffered: a late write still ends its goroutine
		go func() { errc <- write(s) }()
		select {
		case err := <-errc:
			return done(err)
		case <-time.After(limit):
			return done(ErrTimeout)
		}
	}
}
