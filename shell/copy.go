package shell

import (
	"errors"
	"time"

	"github.com/atotto/clipboard"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
)

// copiedMsg is a finished clipboard write, carrying the status to show.
type copiedMsg struct {
	done, failed string
	err          error
}

// clipboardTimeout bounds a write that never comes back. A variable so a
// test can shrink it.
var clipboardTimeout = 5 * time.Second

// errClipboardTimeout is worded for the status row: a write that hangs is
// almost always a missing or wedged helper, not anything about the text.
var errClipboardTimeout = errors.New("the clipboard did not answer - is xclip, wl-copy or pbcopy installed and working?")

// WithClipboard replaces the system clipboard with write, for a test that
// must not touch the machine's own.
func (s Shell) WithClipboard(write func(string) error) Shell {
	s.cfg.Clipboard = write
	return s
}

// Copy puts text on the clipboard off the event loop, then shows done on the
// status row, or "failed: why". A helper that hangs gives up after 5s
// instead of freezing the program.
func (s Shell) Copy(text, done, failed string) tea.Cmd {
	write := s.cfg.Clipboard // read here, on the loop, never from the command
	if write == nil {
		write = clipboard.WriteAll
	}
	return func() tea.Msg {
		errc := make(chan error, 1) // buffered: a late write must not leak a goroutine
		go func() { errc <- write(text) }()
		select {
		case err := <-errc:
			return copiedMsg{done: done, failed: failed, err: err}
		case <-time.After(clipboardTimeout):
			return copiedMsg{done: done, failed: failed, err: errClipboardTimeout}
		}
	}
}

// copied shows how a copy went.
func (s Shell) copied(m copiedMsg) (Shell, tea.Cmd) {
	if m.err != nil {
		return s.SetStatus(draw.Sanitize(m.failed+": "+m.err.Error()), Error, s.cfg.StatusTTL)
	}
	return s.SetStatus(draw.Sanitize(m.done), OK, s.cfg.StatusTTL)
}
