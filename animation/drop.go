package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// settleHop is the height, in cells, below which a bounce is too small to draw.
const settleHop = 0.05

// Drop lets a value fall onto a floor under gravity and bounce until it
// settles: a notice that drops into place, a card that lands. Unlike
// Out(Bounce) it follows real motion: each hop lower and quicker than the last.
type Drop struct {
	Cur      float64
	from, to float64
	gravity  float64    // cells a second squared
	bounce   float64    // the share of its speed the floor gives back
	t        float64    // seconds since it let go
	run      frames.Run // its frames only; the motion is solved from the time
}

// NewDrop falls from from to the floor at to, either way up the screen.
func NewDrop(from, to float64) Drop {
	d := Drop{Cur: from, from: from, to: to, gravity: 60, bounce: 0.45, run: frames.NewRun(time.Second)}
	d.run.Repeat = Forever
	if !d.Active() {
		d.Cur = to
	}
	return d
}

// WithGravity is d pulled at g cells a second squared: 60 when unset.
func (d Drop) WithGravity(g float64) Drop {
	if g > 0 {
		d.gravity = g
	}
	return d
}

// WithBounce is d keeping e of its speed at each hit, in [0,1): 0 lands dead.
func (d Drop) WithBounce(e float64) Drop {
	d.bounce = min(max(e, 0), 0.95)
	return d
}

// at is where d is t seconds after it let go, and whether it has settled.
func (d Drop) at(t float64) (float64, bool) {
	height := math.Abs(d.to - d.from)
	sign := 1.0
	if d.to < d.from {
		sign = -1
	}
	fall := math.Sqrt(2 * height / d.gravity)
	if t < fall {
		return d.from + sign*d.gravity*t*t/2, false
	}
	t -= fall
	for v := d.gravity * fall * d.bounce; v*v/(2*d.gravity) >= settleHop; v *= d.bounce {
		hop := 2 * v / d.gravity
		if t < hop {
			return d.to - sign*(v*t-d.gravity*t*t/2), false
		}
		t -= hop
	}
	return d.to, true
}

// Active reports whether d is still falling or bouncing.
func (d Drop) Active() bool { return d.run.Active() }

// Tick is the command for d's next frame, nil once it settled.
func (d Drop) Tick() tea.Cmd { return d.run.Tick() }

// FrameAt is d's frame for at, for a test or a host that delivers frames itself.
func (d Drop) FrameAt(at time.Time) FrameMsg { return d.run.FrameAt(at) }

// Update falls on its own frames and asks for the next until it settles.
func (d Drop) Update(msg tea.Msg) (Drop, tea.Cmd) {
	f, ok := d.run.Own(msg)
	if !ok {
		return d, nil
	}
	if d, _ = d.Advance(f.At); !d.Active() {
		return d, nil
	}
	var cmd tea.Cmd
	d.run, cmd = d.run.Next()
	return d, cmd
}

// Advance is d at now, and whether it settled.
func (d Drop) Advance(now time.Time) (Drop, bool) {
	if !d.Active() {
		return d, true
	}
	d.run, _ = d.run.Advance(now)
	d.t = frames.Paced(now.Sub(d.run.Start)).Seconds()
	pos, settled := d.at(d.t)
	d.Cur = pos
	if settled {
		d.run = frames.Run{}
	}
	return d, settled
}
