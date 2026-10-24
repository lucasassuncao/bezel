package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Path carries a point along a line of points at an even speed, however far
// apart they sit: a marker that follows a route, a cursor that curves.
type Path struct {
	pts  []Point
	lens []float64 // distance from the first point to each one
	Cur  Point     // where the last frame put it
	run  frames.Run
}

// NewPath travels points over dur. With fewer than two points, a path of no
// length, or motion Still, it is inactive at the last point.
func NewPath(dur time.Duration, points ...Point) Path {
	p := Path{pts: append([]Point(nil), points...)}
	if len(p.pts) == 0 {
		return p
	}
	p.lens = make([]float64, len(p.pts))
	for i := 1; i < len(p.pts); i++ {
		p.lens[i] = p.lens[i-1] + math.Hypot(p.pts[i].X-p.pts[i-1].X, p.pts[i].Y-p.pts[i-1].Y)
	}
	p.Cur = p.pts[len(p.pts)-1]
	if p.lens[len(p.lens)-1] == 0 {
		return p
	}
	if p.run = frames.NewRun(dur); p.run.Active() {
		p.Cur = p.pts[0]
	}
	return p
}

// Bezier is the cubic curve from a to d pulled towards b and c, as n+1 points
// for NewPath.
func Bezier(a, b, c, d Point, n int) []Point {
	n = max(n, 1)
	out := make([]Point, n+1)
	for i := range out {
		t := float64(i) / float64(n)
		u := 1 - t
		w0, w1, w2, w3 := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
		out[i] = Point{
			X: w0*a.X + w1*b.X + w2*c.X + w3*d.X,
			Y: w0*a.Y + w1*b.Y + w2*c.Y + w3*d.Y,
		}
	}
	return out
}

// at is the point a share q of the way along.
func (p Path) at(q float64) Point {
	total := p.lens[len(p.lens)-1]
	want := min(max(q, 0), 1) * total
	for i := 1; i < len(p.pts); i++ {
		if want > p.lens[i] && i < len(p.pts)-1 {
			continue
		}
		seg := p.lens[i] - p.lens[i-1]
		f := 0.0
		if seg > 0 {
			f = (want - p.lens[i-1]) / seg
		}
		a, b := p.pts[i-1], p.pts[i]
		return Point{X: a.X + (b.X-a.X)*f, Y: a.Y + (b.Y-a.Y)*f}
	}
	return p.pts[len(p.pts)-1]
}

// WithEase is p following e along its length instead of EaseInOut.
func (p Path) WithEase(e Ease) Path {
	p.run.Ease = e
	return p
}

// WithDelay holds p at its start for d after its first frame.
func (p Path) WithDelay(d time.Duration) Path {
	p.run.Delay = max(d, 0)
	return p
}

// WithRepeat travels p n more times after the first, or without end for Forever.
func (p Path) WithRepeat(n int) Path {
	p.run.Repeat = max(n, Forever)
	return p
}

// WithYoyo travels every other pass of p back.
func (p Path) WithYoyo() Path {
	p.run.Yoyo = true
	return p
}

// Active reports whether p is still travelling.
func (p Path) Active() bool { return p.run.Active() }

// Tick is the command for p's next frame, nil once it arrived.
func (p Path) Tick() tea.Cmd { return p.run.Tick() }

// FrameAt is p's frame for at, for a test or a host that delivers frames itself.
func (p Path) FrameAt(at time.Time) FrameMsg { return p.run.FrameAt(at) }

// Update travels on its own frames and asks for the next until it arrives.
func (p Path) Update(msg tea.Msg) (Path, tea.Cmd) {
	f, ok := p.run.Own(msg)
	if !ok {
		return p, nil
	}
	if p, _ = p.Advance(f.At); !p.Active() {
		return p, nil
	}
	var cmd tea.Cmd
	p.run, cmd = p.run.Next()
	return p, cmd
}

// Advance is p at now, and whether it arrived.
func (p Path) Advance(now time.Time) (Path, bool) {
	if !p.Active() {
		return p, true
	}
	ease := frames.Or(p.run.Ease)
	var q float64
	p.run, q = p.run.Advance(now)
	p.Cur = p.at(ease(q))
	return p, !p.Active()
}
