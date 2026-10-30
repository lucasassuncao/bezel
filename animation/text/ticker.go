package text

import (
	"fmt"
	"image/color"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Ticker is a number that rolls to each new value and flashes the colour of
// its direction, settling back to base: a price, a counter, a score.
type Ticker struct {
	from, to       int
	base, up, down color.Color
	dur            time.Duration
	p              float64
	run            frames.Run
}

var _ bezel.Model[Ticker] = Ticker{}

// NewTicker shows value; each change rolls over dur. It keeps the terminal's
// colour until WithBase, WithUp and WithDown give it some.
func NewTicker(value int, dur time.Duration) Ticker {
	return Ticker{from: value, to: value, dur: dur, p: 1}
}

// WithBase is the colour the number rests in.
func (t Ticker) WithBase(c color.Color) Ticker {
	t.base = c
	return t
}

// WithUp is the colour the number flashes when it rises.
func (t Ticker) WithUp(c color.Color) Ticker {
	t.up = c
	return t
}

// WithDown is the colour the number flashes when it falls.
func (t Ticker) WithDown(c color.Color) Ticker {
	t.down = c
	return t
}

// To rolls the number from where it is now to v.
func (t Ticker) To(v int) (Ticker, tea.Cmd) {
	t.from, t.to = t.Value(), v
	t.run = frames.NewRun(t.dur)
	t.run.Ease = animation.EaseOut
	t.p = 0
	if !t.Active() {
		t.p = 1
	}
	return t, t.Tick()
}

// Value is the number shown now.
func (t Ticker) Value() int {
	q := frames.Or(t.run.Ease)(t.p)
	if !t.Active() {
		q = 1
	}
	return t.from + int(math.Round(float64(t.to-t.from)*q))
}

// Color is the number's colour now: its direction's, fading back to base.
func (t Ticker) Color() color.Color {
	if !t.Active() || t.to == t.from {
		return t.base
	}
	dir := t.up
	if t.to < t.from {
		dir = t.down
	}
	return animation.Blend(dir, t.base, t.p)
}

// View is the number in its colour.
func (t Ticker) View() string {
	return lipgloss.NewStyle().Foreground(t.Color()).Render(fmt.Sprint(t.Value()))
}

// Active reports whether the number is still rolling.
func (t Ticker) Active() bool { return t.run.Active() }

// Tick is the command for t's next frame, nil at rest.
func (t Ticker) Tick() tea.Cmd { return t.run.Tick() }

// FrameAt is t's frame for at, for a test or a host that delivers frames itself.
func (t Ticker) FrameAt(at time.Time) animation.FrameMsg { return t.run.FrameAt(at) }

// Update rolls on its own frames and asks for the next until it settles.
func (t Ticker) Update(msg tea.Msg) (Ticker, bezel.Result) {
	r, p, ok, cmd := t.run.Update(msg)
	if ok {
		t.run, t.p = r, p
		t.settle()
	}
	return t, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is t at now, and whether it settled.
func (t Ticker) Advance(now time.Time) (Ticker, bool) {
	if !t.Active() {
		return t, true
	}
	t.run, t.p = t.run.Advance(now)
	t.settle()
	return t, !t.Active()
}

// settle lands t on its value once its run is over.
func (t *Ticker) settle() {
	if !t.Active() {
		t.from, t.p = t.to, 1
	}
}
