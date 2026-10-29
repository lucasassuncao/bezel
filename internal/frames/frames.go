// Package frames is the clock every bezel animation keeps: the frame message,
// the run that counts a timed motion's progress, and the motion switches read
// from the environment.
package frames

import (
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Frame is the interval between animation frames (~60fps).
const Frame = 16 * time.Millisecond

// Msg is one tick of an animation's clock, at the time it fired; id and tag
// say whose frame it is and which step, so a twin from a second Tick is dropped.
type Msg struct {
	At      time.Time
	id, tag int64
}

// NewMsg is the frame for the animation id at step tag.
func NewMsg(at time.Time, id, tag int64) Msg { return Msg{At: at, id: id, tag: tag} }

// Is reports whether msg is the frame for id at step tag.
func Is(msg tea.Msg, id, tag int64) (Msg, bool) {
	f, ok := msg.(Msg)
	return f, ok && f.id == id && f.tag == tag
}

// Of reports whether msg is one of id's frames, of any step: a spent one is
// still id's to drop, never another's to take.
func Of(msg tea.Msg, id int64) bool {
	f, ok := msg.(Msg)
	return ok && f.id == id
}

// Tick is the command that delivers id's frame for step tag after every.
func Tick(every time.Duration, id, tag int64) tea.Cmd {
	return tea.Tick(every, func(at time.Time) tea.Msg { return Msg{At: at, id: id, tag: tag} })
}

var lastID atomic.Int64

// NextID numbers a new animation, so a replaced one's frames are told apart.
func NextID() int64 { return lastID.Add(1) }
