package shell

import (
	tea "charm.land/bubbletea/v2"
)

// Busy shows a spinner and text on the status row until Idle, which an empty
// text also is. A banner set meanwhile shows over it until it expires. The
// command starts the spinner; nil when it is already spinning.
func (s Shell) Busy(text string) (Shell, tea.Cmd) {
	s = s.ready()
	if text == "" {
		return s.Idle(), nil
	}
	wasBusy := s.busy != ""
	s.busy = text
	if wasBusy {
		return s, nil
	}
	var cmd tea.Cmd
	s.spin, cmd = s.spin.Start()
	return s, cmd
}

// Idle ends Busy. A banner on the status row is left alone.
func (s Shell) Idle() Shell {
	s = s.ready()
	s.busy = ""
	s.spin = s.spin.Stop()
	return s
}

// IsBusy reports whether Busy is in effect.
func (s Shell) IsBusy() bool { return s.busy != "" }

// Spinner is the current frame, unstyled, for a pane showing the same work.
func (s Shell) Spinner() string { return s.spin.Frame() }
