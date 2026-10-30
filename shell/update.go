package shell

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/component/input"
	"github.com/lucasassuncao/bezel/component/toast"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/palette"
)

// statusExpiredMsg clears the status whose sequence it carries. A stale one,
// from a message already replaced, is dropped.
type statusExpiredMsg struct{ seq uint }

// toastMsg carries a toast's frame or expiry, so the shell knows its own.
type toastMsg struct{ tea.Msg }

// toastCmd wraps what cmd sends in a toastMsg; nil stays nil.
func toastCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg { return toastMsg{cmd()} }
}

// Update handles what is the shell's: resize, status ticks, overlays, the
// status-row input, help, tabs and focus, and a left click on a dialog, a
// tab or a pane, which comes back as a ClickMsg. false means the app's turn.
// A resize is both: the shell takes the size and hands it on, false, so the
// app sizes its own models from Rect on the same message.
func (s Shell) Update(msg tea.Msg) (Shell, bool, tea.Cmd) {
	s = s.ready()
	next, handled, cmd := s.update(msg)
	if !handled {
		return s, false, nil
	}
	// The one place a message re-resolves the layout, so no branch can forget to.
	_, resize := msg.(tea.WindowSizeMsg)
	return next.relayout(), !resize, cmd
}

func (s Shell) update(msg tea.Msg) (Shell, bool, tea.Cmd) {
	switch m := msg.(type) {
	case tea.MouseClickMsg:
		if m.Button == tea.MouseLeft {
			return s.click(m)
		}

	case animation.FrameMsg:
		if s.trans.Owns(m) {
			var cmd tea.Cmd
			s.trans, cmd = s.trans.Update(m)
			return s, true, cmd
		}
		if s.spin.Owns(m) {
			var cmd tea.Cmd
			s.spin, cmd = s.spin.Update(m)
			return s, true, cmd
		}
		if run, p, ok, cmd := s.glow.run.Update(m); ok {
			s.glow.run, s.glow.p = run, p
			return s, true, cmd
		}

	case tea.WindowSizeMsg:
		s.width, s.height = m.Width, m.Height
		return s, true, nil

	case statusExpiredMsg:
		if m.seq == s.statusSeq {
			s.status = ""
		}
		return s, true, nil

	case toastMsg:
		var cmd tea.Cmd
		s.toasts, cmd = s.toasts.Update(m.Msg)
		return s, true, toastCmd(cmd)

	case copiedMsg:
		next, cmd := s.copied(m)
		return next, true, cmd

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
			return s, true, m.Cmd
		}
		var cmd tea.Cmd
		s, cmd = s.SetStatus(m.Err, Error, s.cfg.StatusTTL)
		return s, true, tea.Batch(cmd, m.Cmd)

	case overlay.CloseMsg, overlay.PushMsg, overlay.TopMsg:
		var cmd tea.Cmd
		s.overlays, _, cmd = s.overlays.Update(msg)
		return s, true, cmd
	}

	if b, ok := msg.(input.BlinkMsg); ok {
		return s.blink(b)
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
			return s, true, cmd
		}
		// Anything else is an answer for the app, not the modal.
		return s, false, nil
	}

	if s.typing {
		switch msg.(type) {
		case tea.KeyPressMsg, tea.PasteMsg:
			var cmd tea.Cmd
			s.input, cmd = s.input.Update(msg)
			return s, true, cmd
		}
	}

	k, isKey := msg.(tea.KeyPressMsg)
	if !isKey {
		return s, false, nil
	}

	if next, cmd, ok := s.runKey(k); ok {
		return next, true, cmd
	}
	return s, false, nil
}

// SetStatus shows text on the status row, or as a toast in the Toast style.
// ttl > 0 clears it later, unless a newer status has replaced it by then; a
// toast always leaves, after StatusTTL when ttl is not above zero.
func (s Shell) SetStatus(text string, level Level, ttl time.Duration) (Shell, tea.Cmd) {
	s = s.ready()
	s.statusSeq++
	s.status, s.level = text, level
	var shown tea.Cmd
	if s.cfg.StatusStyle == Toast {
		life := ttl
		if life <= 0 {
			life = s.cfg.StatusTTL
		}
		s.toasts, shown = s.toasts.WithTTL(life).Push(text, toastKind(level))
		shown = toastCmd(shown)
	}
	if ttl <= 0 {
		return s, shown
	}
	seq := s.statusSeq
	return s, tea.Batch(shown, tea.Tick(ttl, func(time.Time) tea.Msg { return statusExpiredMsg{seq: seq} }))
}

// toastKind is the toast that reports a status at level.
func toastKind(level Level) toast.Kind {
	switch level {
	case OK:
		return toast.Success
	case Error:
		return toast.Error
	}
	return toast.Info
}

// StatusLevel is how the current status is drawn.
func (s Shell) StatusLevel() Level { return s.level }

// StatusExpiry is the message that would clear the current status, for a
// test harness that drops timers and delivers them itself.
func (s Shell) StatusExpiry() tea.Msg { return statusExpiredMsg{seq: s.statusSeq} }

// IsStatusExpiry reports a status timer's message, or a toast's frame, for a
// harness that must not feed timers back into the loop.
func IsStatusExpiry(msg tea.Msg) bool {
	switch msg.(type) {
	case statusExpiredMsg, toastMsg:
		return true
	}
	return false
}

// ClearStatus removes the banner now, and makes its pending timer a no-op.
func (s Shell) ClearStatus() Shell {
	s = s.ready()
	s.statusSeq++
	s.status = ""
	return s
}

// SetInput puts a text input on the status row. While it is there it takes
// every key; the app reads the edits back with Input after each Update.
func (s Shell) SetInput(m input.Model) Shell {
	s = s.ready()
	s.input, s.typing = m, true
	return s
}

// ClearInput takes the input off the status row and gives the keys back.
func (s Shell) ClearInput() Shell {
	s = s.ready()
	s.input, s.typing = input.Model{}, false
	return s
}

// Input is the status row's input as edited so far, and false when none is set.
func (s Shell) Input() (input.Model, bool) { return s.input, s.typing }

// blink hands a caret's blink to the field that armed it: an open dialog's
// or the status row's, which re-arm it. A blink neither knows is the app's,
// for a field it drives itself, so handled is false.
func (s Shell) blink(msg input.BlinkMsg) (Shell, bool, tea.Cmd) {
	if s.overlays.Len() > 0 {
		if next, _, cmd := s.overlays.Update(msg); cmd != nil {
			s.overlays = next
			return s, true, cmd
		}
	}
	if s.typing {
		if in, cmd := s.input.Update(msg); cmd != nil {
			s.input = in
			return s, true, cmd
		}
	}
	return s, false, nil
}
