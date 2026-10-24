package animation

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Clock is one chain of frames for a screen running several animations at
// once: each frame, the host advances every one of them to its time, and keeps
// the clock going while any still moves. One tick a frame instead of one each.
type Clock struct {
	running bool
	id, tag int64
	every   time.Duration // the gap it asks between frames; Frame when zero
	last    time.Time     // the previous frame Take saw
}

// NewClock is a stopped clock.
func NewClock() Clock { return Clock{id: frames.NextID()} }

// Start sets c ticking; a running clock asks for nothing more.
func (c Clock) Start() (Clock, tea.Cmd) {
	if c.running {
		return c, nil
	}
	if c.id == 0 { // a Clock never built with NewClock
		c.id = frames.NextID()
	}
	c.running, c.last = true, time.Time{}
	return c.Next()
}

// Stop halts c; the frame already on its way is dropped.
func (c Clock) Stop() Clock {
	c.running, c.last = false, time.Time{}
	c.tag++
	return c
}

// Running reports whether c is ticking.
func (c Clock) Running() bool { return c.running }

// Frame is the time msg carries when it is c's current frame. Advance every
// animation to it, then call Next while any of them still moves.
func (c Clock) Frame(msg tea.Msg) (time.Time, bool) {
	f, ok := frames.Is(msg, c.id, c.tag)
	if !ok || !c.running {
		return time.Time{}, false
	}
	return f.At, true
}

// Take is Frame for a host that lets c pace itself: frames that arrive twice
// as far apart as asked mean the screen cannot keep up, so c asks half as
// often, down to a quarter of the usual rate, and faster again once they keep up.
func (c Clock) Take(msg tea.Msg) (Clock, time.Time, bool) {
	at, ok := c.Frame(msg)
	if !ok {
		return c, at, false
	}
	if gap := c.gap(); !c.last.IsZero() {
		switch took := at.Sub(c.last); {
		case took > 2*gap:
			c.every = min(2*gap, 4*Frame)
		case took < gap+gap/2 && gap > Frame:
			c.every = max(gap/2, Frame)
		}
	}
	c.last = at
	return c, at, true
}

// gap is how long c waits between frames.
func (c Clock) gap() time.Duration {
	if c.every <= 0 {
		return Frame
	}
	return c.every
}

// Next is c with the command for its following frame, nil when stopped.
func (c Clock) Next() (Clock, tea.Cmd) {
	if !c.running {
		return c, nil
	}
	c.tag++
	id, tag := c.id, c.tag
	return c, tea.Tick(c.gap(), func(at time.Time) tea.Msg { return frames.NewMsg(at, id, tag) })
}

// FrameAt is c's current frame for at, for a test that delivers frames itself.
func (c Clock) FrameAt(at time.Time) FrameMsg { return frames.NewMsg(at, c.id, c.tag) }
