package animation

import (
	"math"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Flock is a line of points that follow one another: the first springs after
// the target, each other one after the point ahead of it, so they trail like
// a tail behind a cursor.
type Flock struct {
	pts, vel      []Point
	target        Point
	freq, damping float64
	last          time.Time
	moving        bool
	id, tag       int64
}

// NewFlock is n points at rest at at, springs of stiffness 14 and damping 0.8.
func NewFlock(n int, at Point) Flock {
	f := Flock{pts: make([]Point, max(n, 1)), vel: make([]Point, max(n, 1)), target: at, freq: 14, damping: 0.8, id: frames.NextID()}
	for i := range f.pts {
		f.pts[i] = at
	}
	return f
}

// WithFreq is f's springs at stiffness w, in radians a second.
func (f Flock) WithFreq(w float64) Flock {
	if w > 0 {
		f.freq = w
	}
	return f
}

// WithDamping is f's springs at damping d: 1 settles, below it bounces.
func (f Flock) WithDamping(d float64) Flock {
	if d > 0 {
		f.damping = d
	}
	return f
}

// To aims the first point at p. When motion is Still every point jumps there.
func (f Flock) To(p Point) (Flock, tea.Cmd) {
	f.target = p
	f.pts, f.vel = slices.Clone(f.pts), slices.Clone(f.vel)
	if Still() {
		for i := range f.pts {
			f.pts[i], f.vel[i] = p, Point{}
		}
		f.moving = false
		return f, nil
	}
	if f.moving {
		return f, nil
	}
	f.moving, f.last = true, time.Time{}
	f.tag++
	return f, f.Tick()
}

// Points is where every point is now, the leader first.
func (f Flock) Points() []Point { return slices.Clone(f.pts) }

// Active reports whether any point is still moving.
func (f Flock) Active() bool { return f.moving }

// Tick is the command for f's next frame, nil at rest.
func (f Flock) Tick() tea.Cmd {
	if !f.moving {
		return nil
	}
	return frames.Tick(Frame, f.id, f.tag)
}

// FrameAt is f's frame for at, for a test or a host that delivers frames itself.
func (f Flock) FrameAt(at time.Time) FrameMsg { return frames.NewMsg(at, f.id, f.tag) }

// Update moves f on its own frames and asks for the next until it rests.
func (f Flock) Update(msg tea.Msg) (Flock, tea.Cmd) {
	fm, ok := frames.Is(msg, f.id, f.tag)
	if !ok || !f.moving {
		return f, nil
	}
	if f, _ = f.Advance(fm.At); !f.moving {
		return f, nil
	}
	f.tag++
	return f, f.Tick()
}

// Advance is f moved to now, and whether every point came to rest. Each point
// springs after where the one ahead of it has just been moved to.
func (f Flock) Advance(now time.Time) (Flock, bool) {
	if !f.moving {
		return f, true
	}
	if f.last.IsZero() {
		f.last = now
		return f, false
	}
	dt := max(frames.Paced(now.Sub(f.last)).Seconds(), 0)
	f.last = now
	f.pts, f.vel = slices.Clone(f.pts), slices.Clone(f.vel)
	still := true
	goal := f.target
	for i := range f.pts {
		x, vx := step(f.pts[i].X-goal.X, f.vel[i].X, f.freq, f.damping, dt)
		y, vy := step(f.pts[i].Y-goal.Y, f.vel[i].Y, f.freq, f.damping, dt)
		f.pts[i], f.vel[i] = Point{goal.X + x, goal.Y + y}, Point{vx, vy}
		if math.Abs(x) > rest || math.Abs(y) > rest || math.Abs(vx) > rest*10 || math.Abs(vy) > rest*10 {
			still = false
		}
		goal = f.pts[i]
	}
	if still {
		for i := range f.pts {
			f.pts[i], f.vel[i] = f.target, Point{}
		}
		f.moving = false
	}
	return f, !f.moving
}
