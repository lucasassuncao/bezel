package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Orbit carries a point round a centre on an ellipse, once every period, for
// as long as it lives: a spinner with depth, a satellite round a title.
type Orbit struct {
	centre Point
	rx, ry float64
	p      float64
	run    frames.Run
}

var _ bezel.Model[Orbit] = Orbit{}

// NewOrbit circles centre at rx columns and ry rows; a terminal cell is about
// twice as tall as wide, so ry near rx/2 looks round.
func NewOrbit(centre Point, rx, ry float64, period time.Duration) Orbit {
	o := Orbit{centre: centre, rx: rx, ry: ry, run: frames.NewRun(period)}
	o.run.Repeat = Forever
	return o
}

// WithRepeat goes round n more times after the first instead of forever.
func (o Orbit) WithRepeat(n int) Orbit {
	o.run.Repeat = max(n, Forever)
	return o
}

// angle is how far round the point is, in radians, from the right going down.
func (o Orbit) angle() float64 { return 2 * math.Pi * o.p }

// Pos is where the point is now.
func (o Orbit) Pos() Point {
	a := o.angle()
	return Point{X: o.centre.X + o.rx*math.Cos(a), Y: o.centre.Y + o.ry*math.Sin(a)}
}

// Near reports whether the point is on the near half of its ellipse, the one
// below the centre: draw it over what it circles there, behind it on the far half.
func (o Orbit) Near() bool { return math.Sin(o.angle()) >= 0 }

// Active reports whether the point is still going round.
func (o Orbit) Active() bool { return o.run.Active() }

// Tick is the command for o's next frame, nil once it stopped.
func (o Orbit) Tick() tea.Cmd { return o.run.Tick() }

// FrameAt is o's frame for at, for a test or a host that delivers frames itself.
func (o Orbit) FrameAt(at time.Time) FrameMsg { return o.run.FrameAt(at) }

// Update goes round on its own frames and asks for the next while it runs.
func (o Orbit) Update(msg tea.Msg) (Orbit, bezel.Result) {
	r, p, ok, cmd := o.run.Update(msg)
	if ok {
		o.run, o.p = r, p
	}
	return o, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is o at now, and whether it stopped.
func (o Orbit) Advance(now time.Time) (Orbit, bool) {
	if !o.Active() {
		return o, true
	}
	o.run, o.p = o.run.Advance(now)
	return o, !o.Active()
}
