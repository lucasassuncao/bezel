package shell

import (
	"errors"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/clip"
)

// copiedMsg is a finished clipboard write, carrying the status to show.
type copiedMsg struct {
	done, failed string
	err          error
}

// errClipboardTimeout is worded for the status row: a write that hangs is
// almost always a missing or wedged helper, not anything about the text.
var errClipboardTimeout = errors.New("the clipboard did not answer: check that xclip, wl-copy or pbcopy is installed and working")

// Copy puts text on the clipboard off the event loop, then shows done on the
// status row, or "failed: why". A helper that hangs gives up after 5s
// instead of freezing the program.
func (s Shell) Copy(text, done, failed string) tea.Cmd {
	s = s.ready()
	return clip.Copy(s.cfg.Clipboard, text, func(err error) tea.Msg {
		if errors.Is(err, clip.ErrTimeout) {
			err = errClipboardTimeout
		}
		return copiedMsg{done: done, failed: failed, err: err}
	})
}

// copied shows how a copy went.
func (s Shell) copied(m copiedMsg) (Shell, tea.Cmd) {
	if m.err != nil {
		return s.SetStatus(draw.Sanitize(m.failed+": "+m.err.Error()), Error, s.cfg.StatusTTL)
	}
	return s.SetStatus(draw.Sanitize(m.done), OK, s.cfg.StatusTTL)
}
