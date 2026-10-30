package animation

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
)

// Rubber is a value kept between two limits that gives past them, harder the
// farther it is pulled, and springs back when let go: a list scrolled past its
// end, a slider dragged off its track.
type Rubber struct {
	lo, hi float64
	give   float64 // the most it ever goes past a limit
	s      Spring
}

var _ bezel.Model[Rubber] = Rubber{}

// NewRubber holds a value in [lo, hi], at lo to begin with.
func NewRubber(lo, hi float64) Rubber {
	return Rubber{lo: lo, hi: max(hi, lo), give: 4, s: NewSpring(lo).WithFreq(16)}
}

// WithGive lets the value go at most g past either limit, however hard it is pulled.
func (r Rubber) WithGive(g float64) Rubber {
	if g > 0 {
		r.give = g
	}
	return r
}

// Pull puts the value where v asks, resisted past the limits: it never gets
// further than give beyond one.
func (r Rubber) Pull(v float64) Rubber {
	pos := v
	switch {
	case v > r.hi:
		d := v - r.hi
		pos = r.hi + r.give*d/(d+r.give)
	case v < r.lo:
		d := r.lo - v
		pos = r.lo - r.give*d/(d+r.give)
	}
	r.s = NewSpring(pos).WithFreq(r.s.freq()).WithDamping(r.s.damping())
	return r
}

// Release lets go: a value past a limit springs back onto it.
func (r Rubber) Release() (Rubber, tea.Cmd) {
	var cmd tea.Cmd
	r.s, cmd = r.s.To(min(max(r.s.pos, r.lo), r.hi))
	return r, cmd
}

// Value is the value now.
func (r Rubber) Value() float64 { return r.s.pos }

// Active reports whether the value is still springing back.
func (r Rubber) Active() bool { return r.s.Active() }

// Tick is the command for r's next frame, nil at rest.
func (r Rubber) Tick() tea.Cmd { return r.s.Tick() }

// FrameAt is r's frame for at, for a test or a host that delivers frames itself.
func (r Rubber) FrameAt(at time.Time) FrameMsg { return r.s.FrameAt(at) }

// Update springs back on its own frames and asks for the next until it rests.
func (r Rubber) Update(msg tea.Msg) (Rubber, bezel.Result) {
	var res bezel.Result
	r.s, res = r.s.Update(msg)
	return r, res
}

// Advance is r at now, and whether it rests.
func (r Rubber) Advance(now time.Time) (Rubber, bool) {
	var done bool
	r.s, done = r.s.Advance(now)
	return r, done
}
