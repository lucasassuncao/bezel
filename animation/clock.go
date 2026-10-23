package animation

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Clock is one chain of frames for a screen running several animations at
// once: each frame, the host advances every one of them to its time, and keeps
// the clock going while any still moves. One tick a frame instead of one each.
type Clock struct {
	running bool
	id, tag int64
}

// NewClock is a stopped clock.
func NewClock() Clock { return Clock{id: lastID.Add(1)} }

// Start sets c ticking; a running clock asks for nothing more.
func (c Clock) Start() (Clock, tea.Cmd) {
	if c.running {
		return c, nil
	}
	if c.id == 0 { // a Clock never built with NewClock
		c.id = lastID.Add(1)
	}
	c.running = true
	return c.Next()
}

// Stop halts c; the frame already on its way is dropped.
func (c Clock) Stop() Clock {
	c.running = false
	c.tag++
	return c
}

// Running reports whether c is ticking.
func (c Clock) Running() bool { return c.running }

// Frame is the time msg carries when it is c's current frame. Advance every
// animation to it, then call Next while any of them still moves.
func (c Clock) Frame(msg tea.Msg) (time.Time, bool) {
	f, ok := msg.(FrameMsg)
	if !ok || !c.running || f.id != c.id || f.tag != c.tag {
		return time.Time{}, false
	}
	return f.At, true
}

// Next is c with the command for its following frame, nil when stopped.
func (c Clock) Next() (Clock, tea.Cmd) {
	if !c.running {
		return c, nil
	}
	c.tag++
	id, tag := c.id, c.tag
	return c, tea.Tick(Frame, func(at time.Time) tea.Msg { return FrameMsg{At: at, id: id, tag: tag} })
}

// FrameAt is c's current frame for at, for a test that delivers frames itself.
func (c Clock) FrameAt(at time.Time) FrameMsg { return FrameMsg{At: at, id: c.id, tag: c.tag} }
