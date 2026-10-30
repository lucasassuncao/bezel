package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Float is Tween over a float64, for what moves by less than a cell: a
// progress bar drawn in eighths, a scrollbar's thumb, a colour's mix.
type Float struct {
	from, to float64
	cur      float64 // the value the last frame sampled
	run      frames.Run
}

var _ bezel.Model[Float] = Float{}

// NewFloat eases from from to to over dur, its clock started by its first
// frame. When dur <= 0, from == to or motion is Still it is inactive at to.
func NewFloat(from, to float64, dur time.Duration) Float {
	if dur <= 0 || from == to || Still() {
		return Float{from: to, to: to, cur: to}
	}
	return Float{from: from, to: to, cur: from, run: frames.NewRun(dur)}
}

// WithEase is f following e instead of InOut(Quad).
func (f Float) WithEase(e Ease) Float {
	f.run.Ease = e
	return f
}

// Active reports whether f is still moving.
func (f Float) Active() bool { return f.run.Active() }

// Int is Value rounded to the nearest whole number.
func (f Float) Int() int { return int(math.Round(f.cur)) }

// Tick is the command for f's next frame, nil once it landed.
func (f Float) Tick() tea.Cmd { return f.run.Tick() }

// FrameAt is f's frame for at, for a test or a host that delivers frames itself.
func (f Float) FrameAt(at time.Time) FrameMsg { return f.run.FrameAt(at) }

// Update samples f on its own frames and asks for the next until it lands.
func (f Float) Update(msg tea.Msg) (Float, bezel.Result) {
	fm, ok := f.run.Own(msg)
	if !ok {
		return f, bezel.Result{}
	}
	if f, _ = f.Advance(fm.At); !f.Active() {
		return f, bezel.Result{Handled: true}
	}
	var cmd tea.Cmd
	f.run, cmd = f.run.Next()
	return f, bezel.Result{Handled: true, Cmd: cmd}
}

// Advance is f sampled at now, and whether it landed: for a host driving
// several animations from one Clock.
func (f Float) Advance(now time.Time) (Float, bool) {
	if !f.Active() {
		return f, true
	}
	ease := frames.Or(f.run.Ease)
	var p float64
	f.run, p = f.run.Advance(now)
	switch {
	case f.Active():
		f.cur = f.from + (f.to-f.from)*ease(p)
	case p == 1:
		f.cur = f.to
	default:
		f.cur = f.from
	}
	return f, !f.Active()
}

// WithDelay holds f at From for d after its first frame.
func (f Float) WithDelay(d time.Duration) Float {
	f.run.Delay = max(d, 0)
	return f
}

// WithRepeat runs f n more times after the first, or without end for Forever.
func (f Float) WithRepeat(n int) Float {
	f.run.Repeat = max(n, Forever)
	return f
}

// WithYoyo sends every other run of f back the way it came.
func (f Float) WithYoyo(on bool) Float {
	f.run.Yoyo = on
	return f
}

// Value is where the last frame put the float.
func (f Float) Value() float64 { return f.cur }
