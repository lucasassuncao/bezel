package shell

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/palette"
)

// statusExpiredMsg clears the status whose sequence it carries. A stale one,
// from a message already replaced, is dropped.
type statusExpiredMsg struct{ seq uint }

// Update handles what is the shell's: resize, status ticks, overlays, the
// status-row input, help, tabs and focus. handled false means the app's turn.
func (s Shell) Update(msg tea.Msg) (Shell, bool, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width, s.height = m.Width, m.Height
		return s.relayout(), true, nil

	case statusExpiredMsg:
		if m.seq == s.statusSeq {
			s.status = ""
		}
		return s.relayout(), true, nil

	case spinner.TickMsg:
		if m.ID == s.spin.ID() {
			next, cmd := s.tick(m)
			return next, true, cmd
		}

	case copiedMsg:
		next, cmd := s.copied(m)
		return next.relayout(), true, cmd

	case runMsg:
		next, cmd := s.run(m.action, m.arg, "")
		return next, true, cmd

	case palette.DoneMsg:
		// Popped here, not by a CloseMsg racing the command it just chose. With
		// no palette on top this is a second enter's echo: drop it.
		if _, ok := s.overlays.Top().(palette.Model); !ok {
			return s, true, nil
		}
		s = s.Pop()
		if m.Err == "" {
			return s.relayout(), true, m.Cmd
		}
		var cmd tea.Cmd
		s, cmd = s.SetStatus(m.Err, Error, s.cfg.StatusTTL)
		return s.relayout(), true, tea.Batch(cmd, m.Cmd)

	case overlay.CloseMsg, overlay.PushMsg:
		var cmd tea.Cmd
		s.overlays, _, cmd = s.overlays.Update(msg)
		return s.relayout(), true, cmd
	}

	if s.overlays.Len() > 0 {
		switch msg.(type) {
		case tea.KeyPressMsg, tea.PasteMsg:
			// Scopes are judged live: an answer may have changed them.
			if p, ok := s.overlays.Top().(palette.Model); ok {
				s.overlays = s.overlays.Pop().Push(p.SetItems(s.paletteItems()))
			}
			var cmd tea.Cmd
			s.overlays, _, cmd = s.overlays.Update(msg)
			return s.relayout(), true, cmd
		}
		// Anything else is an answer for the app, not the modal.
		return s, false, nil
	}

	k, isKey := msg.(tea.KeyPressMsg)
	if !isKey {
		return s, false, nil
	}

	if s.input != nil {
		updated, cmd := s.input.Update(k)
		*s.input = updated
		return s, true, cmd
	}

	if next, cmd, ok := s.runKey(k); ok {
		return next, true, cmd
	}
	return s, false, nil
}

// SetStatus shows text on the status row. ttl > 0 clears it later, unless a
// newer status has replaced it by then.
func (s Shell) SetStatus(text string, level Level, ttl time.Duration) (Shell, tea.Cmd) {
	s.statusSeq++
	s.status, s.level = text, level
	if ttl <= 0 {
		return s, nil
	}
	seq := s.statusSeq
	return s, tea.Tick(ttl, func(time.Time) tea.Msg { return statusExpiredMsg{seq: seq} })
}

// StatusLevel is how the current status is drawn.
func (s Shell) StatusLevel() Level { return s.level }

// StatusExpiry is the message that would clear the current status, for a
// test harness that drops timers and delivers them itself.
func (s Shell) StatusExpiry() tea.Msg { return statusExpiredMsg{seq: s.statusSeq} }

// IsStatusExpiry reports a status timer's message, for a harness that must
// not feed timers back into the loop.
func IsStatusExpiry(msg tea.Msg) bool {
	_, ok := msg.(statusExpiredMsg)
	return ok
}

func (s Shell) ClearStatus() Shell {
	s.statusSeq++
	s.status = ""
	return s
}

// SetInput puts a text input on the status row; while set it takes every
// key, and edits land in *m for the app to read back. nil removes it.
func (s Shell) SetInput(m *textinput.Model) Shell {
	s.input = m
	return s
}

func (s Shell) Input() *textinput.Model { return s.input }
