package shell

import (
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
)

// spinFrames turn once a second, each wide enough to set the text apart.
var spinFrames = []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "}

const spinEvery = time.Second / 10

// shellIDs tells apart the spinners of two shells in one program.
var shellIDs atomic.Int64

// spinTickMsg advances the spinner of shell id, for the loop seq started:
// a tick from a loop Idle ended never doubles a newer one.
type spinTickMsg struct{ id, seq int64 }

// spinner is the busy mark: which frame shows, and which loop drives it.
type spinner struct {
	id, seq int64
	frame   int
}

func (sp spinner) tick() tea.Cmd {
	msg := spinTickMsg{id: sp.id, seq: sp.seq}
	return tea.Tick(spinEvery, func(time.Time) tea.Msg { return msg })
}

// Busy shows a spinner and text on the status row until Idle; a banner set
// meanwhile shows over it and busy returns when it expires. The command
// starts the spinner, and is nil when it is already spinning.
func (s Shell) Busy(text string) (Shell, tea.Cmd) {
	wasBusy := s.busy != ""
	s.busy = text
	if wasBusy || text == "" {
		return s, nil
	}
	s.spin.seq++
	return s, s.spin.tick()
}

// Idle ends Busy. A banner on the status row is left alone.
func (s Shell) Idle() Shell {
	s.busy = ""
	return s
}

// IsBusy reports whether Busy is in effect.
func (s Shell) IsBusy() bool { return s.busy != "" }

// Spinner is the current frame, unstyled, for a pane showing the same work.
func (s Shell) Spinner() string { return spinFrames[s.spin.frame] }

// tick advances the frame, re-arming only while busy so an idle shell stops.
func (s Shell) tick(msg spinTickMsg) (Shell, tea.Cmd) {
	if s.busy == "" || msg.seq != s.spin.seq {
		return s, nil
	}
	s.spin.frame = (s.spin.frame + 1) % len(spinFrames)
	return s, s.spin.tick()
}
