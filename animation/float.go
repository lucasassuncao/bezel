package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Float is Tween over a float64, for what moves by less than a cell: a
// progress bar drawn in eighths, a scrollbar's thumb, a colour's mix.
type Float struct {
	From, To float64
	Cur      float64 // the value the last frame sampled
	run      run
}

// NewFloat eases from from to to over dur, its clock started by its first
// frame. When dur <= 0 or from == to it is inactive and already at to.
func NewFloat(from, to float64, dur time.Duration) Float {
	if dur <= 0 || from == to {
		return Float{From: to, To: to, Cur: to}
	}
	return Float{From: from, To: to, Cur: from, run: newRun(dur)}
}

// WithEase is f following e instead of EaseInOut.
func (f Float) WithEase(e Ease) Float {
	f.run.ease = e
	return f
}

// Active reports whether f is still moving.
func (f Float) Active() bool { return f.run.active() }

// Int is Cur rounded to the nearest whole number.
func (f Float) Int() int { return int(math.Round(f.Cur)) }

// Tick is the command for f's next frame, nil once it landed.
func (f Float) Tick() tea.Cmd { return f.run.tick() }

// FrameAt is f's frame for at, for a test or a host that delivers frames itself.
func (f Float) FrameAt(at time.Time) FrameMsg { return f.run.frameAt(at) }

// Update samples f on its own frames and asks for the next until it lands.
func (f Float) Update(msg tea.Msg) (Float, tea.Cmd) {
	fm, ok := f.run.own(msg)
	if !ok {
		return f, nil
	}
	if f, _ = f.Advance(fm.At); !f.Active() {
		return f, nil
	}
	var cmd tea.Cmd
	f.run, cmd = f.run.next()
	return f, cmd
}

// Advance is f sampled at now, and whether it landed: for a host driving
// several animations from one Clock.
func (f Float) Advance(now time.Time) (Float, bool) {
	ease := f.run.ease.or()
	var p float64
	f.run, p = f.run.advance(now)
	if !f.run.active() {
		f.Cur = f.To
		return f, true
	}
	f.Cur = f.From + (f.To-f.From)*ease(p)
	return f, false
}
