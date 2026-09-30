package shell

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// Busy shows a spinner and text on the status row until Idle; a banner set
// meanwhile shows over it and busy returns when it expires. The command
// starts the spinner, and is nil when it is already spinning.
func (s Shell) Busy(text string) (Shell, tea.Cmd) {
	wasBusy := s.busy != ""
	s.busy = text
	if wasBusy || text == "" {
		return s, nil
	}
	return s, s.spin.Tick
}

// Idle ends Busy. A banner on the status row is left alone.
func (s Shell) Idle() Shell {
	s.busy = ""
	return s
}

func (s Shell) IsBusy() bool { return s.busy != "" }

// Spinner is the current frame, unstyled, for a pane showing the same work.
func (s Shell) Spinner() string { return s.spin.View() }

// tick advances the frame, re-arming only while busy so an idle shell stops.
func (s Shell) tick(msg spinner.TickMsg) (Shell, tea.Cmd) {
	if s.busy == "" {
		return s, nil
	}
	var cmd tea.Cmd
	s.spin, cmd = s.spin.Update(msg)
	return s, cmd
}
