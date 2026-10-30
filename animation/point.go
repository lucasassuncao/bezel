package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Point is a place on the screen in cells, fractional while it moves.
type Point struct{ X, Y float64 }

// Round is the cell p is drawn in.
func (p Point) Round() (x, y int) { return int(math.Round(p.X)), int(math.Round(p.Y)) }

// Move eases a point from one place to another over a duration: an overlay
// that drops in, a marker that travels.
type Move struct {
	From, To Point
	Cur      Point // where the last frame put it
	run      frames.Run
}

var _ bezel.Model[Move] = Move{}

// NewMove goes from from to to over dur. When dur <= 0, from == to or motion
// is Still it is inactive and already at to.
func NewMove(from, to Point, dur time.Duration) Move {
	if dur <= 0 || from == to || Still() {
		return Move{From: to, To: to, Cur: to}
	}
	return Move{From: from, To: to, Cur: from, run: frames.NewRun(dur)}
}

// WithEase is m following e instead of EaseInOut.
func (m Move) WithEase(e Ease) Move {
	m.run.Ease = e
	return m
}

// WithDelay holds m at From for d after its first frame.
func (m Move) WithDelay(d time.Duration) Move {
	m.run.Delay = max(d, 0)
	return m
}

// WithRepeat runs m n more times after the first, or without end for Forever.
func (m Move) WithRepeat(n int) Move {
	m.run.Repeat = max(n, Forever)
	return m
}

// WithYoyo sends every other run of m back the way it came.
func (m Move) WithYoyo(on bool) Move {
	m.run.Yoyo = on
	return m
}

// Active reports whether m is still moving.
func (m Move) Active() bool { return m.run.Active() }

// Tick is the command for m's next frame, nil once it landed.
func (m Move) Tick() tea.Cmd { return m.run.Tick() }

// FrameAt is m's frame for at, for a test or a host that delivers frames itself.
func (m Move) FrameAt(at time.Time) FrameMsg { return m.run.FrameAt(at) }

// Update moves on its own frames and asks for the next until it lands.
func (m Move) Update(msg tea.Msg) (Move, bezel.Result) {
	f, ok := m.run.Own(msg)
	if !ok {
		return m, bezel.Result{}
	}
	if m, _ = m.Advance(f.At); !m.Active() {
		return m, bezel.Result{Handled: true}
	}
	var cmd tea.Cmd
	m.run, cmd = m.run.Next()
	return m, bezel.Result{Handled: true, Cmd: cmd}
}

// Advance is m at now, and whether it landed.
func (m Move) Advance(now time.Time) (Move, bool) {
	if !m.Active() {
		return m, true
	}
	ease := frames.Or(m.run.Ease)
	var p float64
	m.run, p = m.run.Advance(now)
	switch {
	case m.Active():
		e := ease(p)
		m.Cur = Point{m.From.X + (m.To.X-m.From.X)*e, m.From.Y + (m.To.Y-m.From.Y)*e}
	case p == 1:
		m.Cur = m.To
	default:
		m.Cur = m.From
	}
	return m, !m.Active()
}

// Follow springs a point after a target on both axes, on one chain of frames.
type Follow struct {
	x, y    Spring
	id, tag int64
}

var _ bezel.Model[Follow] = Follow{}

// NewFollow is a point at rest at at, each axis a NewSpring.
func NewFollow(at Point) Follow {
	return Follow{x: NewSpring(at.X), y: NewSpring(at.Y), id: frames.NextID()}
}

// WithFreq is f with stiffness w on both axes, in radians a second.
func (f Follow) WithFreq(w float64) Follow {
	f.x, f.y = f.x.WithFreq(w), f.y.WithFreq(w)
	return f
}

// WithDamping is f with damping d on both axes: 1 settles, below it bounces.
func (f Follow) WithDamping(d float64) Follow {
	f.x, f.y = f.x.WithDamping(d), f.y.WithDamping(d)
	return f
}

// To aims f at p. A point at rest starts its frames; a moving one keeps them,
// and its speed. When motion is Still it jumps there.
func (f Follow) To(p Point) (Follow, tea.Cmd) {
	was := f.Active()
	f.x, _ = f.x.To(p.X) // the axes' own ticks go unused: f keeps one clock for both
	f.y, _ = f.y.To(p.Y)
	if was || !f.Active() {
		return f, nil
	}
	if f.id == 0 { // a Follow never built with NewFollow
		f.id = frames.NextID()
	}
	f.tag++
	return f, f.Tick()
}

// Pos is where the point is now.
func (f Follow) Pos() Point { return Point{f.x.Pos, f.y.Pos} }

// Active reports whether either axis is still moving.
func (f Follow) Active() bool { return f.x.Active() || f.y.Active() }

// Tick is the command for f's next frame, nil at rest.
func (f Follow) Tick() tea.Cmd {
	if !f.Active() {
		return nil
	}
	return frames.Tick(Frame, f.id, f.tag)
}

// FrameAt is f's frame for at, for a test or a host that delivers frames itself.
func (f Follow) FrameAt(at time.Time) FrameMsg { return frames.NewMsg(at, f.id, f.tag) }

// Update moves f on its own frames and asks for the next until it rests.
func (f Follow) Update(msg tea.Msg) (Follow, bezel.Result) {
	fm, ok := frames.Is(msg, f.id, f.tag)
	if !ok || !f.Active() {
		return f, bezel.Result{Handled: ok}
	}
	if f, _ = f.Advance(fm.At); !f.Active() {
		return f, bezel.Result{Handled: true}
	}
	f.tag++
	return f, bezel.Result{Handled: true, Cmd: f.Tick()}
}

// Advance is f moved to now, and whether it came to rest.
func (f Follow) Advance(now time.Time) (Follow, bool) {
	f.x, _ = f.x.Advance(now)
	f.y, _ = f.y.Advance(now)
	return f, !f.Active()
}
