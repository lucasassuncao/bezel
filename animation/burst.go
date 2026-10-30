package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/internal/noise"
)

// Spark is one particle of a Burst: where it is, and how much of its life is
// left, 1 when it flies out and 0 when it is gone.
type Spark struct {
	At   Point
	Life float64
}

// Burst throws sparks out of a point in every direction and lets gravity pull
// them down as they fade: a task ticked off, a level up, a "+1" that pops.
// Draw each of Sparks; Still, it is born spent.
type Burst struct {
	at      Point
	n       int
	speed   float64 // cells a second, the fastest spark
	gravity float64 // cells a second squared
	seed    uint64
	secs    float64 // seconds since it went off
	p       float64
	run     frames.Run
}

var _ bezel.Model[Burst] = Burst{}

// NewBurst throws n sparks out of at over dur.
func NewBurst(at Point, n int, dur time.Duration) Burst {
	b := Burst{at: at, n: max(n, 0), speed: 24, gravity: 30, seed: 1, run: frames.NewRun(dur)}
	if !b.Active() {
		b.p = 1
	}
	return b
}

// WithSpeed throws the fastest spark at v cells a second: 24 when unset.
func (b Burst) WithSpeed(v float64) Burst {
	if v > 0 {
		b.speed = v
	}
	return b
}

// WithGravity pulls the sparks down at g cells a second squared: 30 when
// unset, 0 for none.
func (b Burst) WithGravity(g float64) Burst {
	b.gravity = max(g, 0)
	return b
}

// WithSeed throws the sparks another way.
func (b Burst) WithSeed(seed uint64) Burst {
	b.seed = seed
	return b
}

// Sparks is every spark still alive, where it has flown. Rows count double,
// as a terminal cell is twice as tall as it is wide.
func (b Burst) Sparks() []Spark {
	if b.p >= 1 {
		return nil
	}
	out := make([]Spark, b.n)
	for i := range out {
		k := uint64(i) // #nosec G115 -- a hash key
		angle := noise.Unit(noise.Hash(b.seed, k*2)) * 2 * math.Pi
		v := b.speed * (0.35 + 0.65*noise.Unit(noise.Hash(b.seed, k*2+1)))
		x := b.at.X + v*math.Cos(angle)*b.secs
		y := b.at.Y + (v*math.Sin(angle)*b.secs+b.gravity*b.secs*b.secs/2)/2
		out[i] = Spark{At: Point{X: x, Y: y}, Life: 1 - b.p}
	}
	return out
}

// Active reports whether any spark is still flying.
func (b Burst) Active() bool { return b.run.Active() }

// Tick is the command for b's next frame, nil once it is spent.
func (b Burst) Tick() tea.Cmd { return b.run.Tick() }

// FrameAt is b's frame for at, for a test or a host that delivers frames itself.
func (b Burst) FrameAt(at time.Time) FrameMsg { return b.run.FrameAt(at) }

// Update flies on its own frames and asks for the next until it is spent.
func (b Burst) Update(msg tea.Msg) (Burst, bezel.Result) {
	f, ok := b.run.Own(msg)
	if !ok {
		return b, bezel.Result{Handled: ok}
	}
	if b, _ = b.Advance(f.At); !b.Active() {
		return b, bezel.Result{Handled: true}
	}
	var cmd tea.Cmd
	b.run, cmd = b.run.Next()
	return b, bezel.Result{Handled: true, Cmd: cmd}
}

// Advance is b at now, and whether it is spent.
func (b Burst) Advance(now time.Time) (Burst, bool) {
	if !b.Active() {
		return b, true
	}
	b.run, b.p = b.run.Advance(now)
	b.secs = frames.Paced(now.Sub(b.run.Start)).Seconds()
	if !b.Active() {
		b.p = 1
	}
	return b, !b.Active()
}

// WithDelay holds the sparks in for d after the first frame.
func (b Burst) WithDelay(d time.Duration) Burst {
	b.run.Delay = max(d, 0)
	return b
}
