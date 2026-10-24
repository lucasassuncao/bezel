package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Pendulum swings a value either side of 0, Cur cells from the middle: a
// "waiting" sign that sways, a needle settling, a notice that swings in.
type Pendulum struct {
	Cur    float64
	amp    float64
	period time.Duration
	decay  time.Duration // how long a swing takes to shrink to a third; 0 never
	run    frames.Run
}

// NewPendulum swings amp cells either way, once every period, for as long as
// it lives. WithDecay lets it come to rest.
func NewPendulum(amp float64, period time.Duration) Pendulum {
	p := Pendulum{amp: amp, period: period, run: frames.NewRun(time.Second)}
	p.run.Ease, p.run.Repeat = Linear, Forever
	if period <= 0 {
		p.run = frames.Run{}
	}
	return p
}

// WithDecay shrinks every swing, to about a third of its size each d, until
// it is too small to see and the pendulum rests in the middle.
func (p Pendulum) WithDecay(d time.Duration) Pendulum {
	p.decay = max(d, 0)
	return p
}

// at is the swing t seconds after the first frame, and whether it has died out.
func (p Pendulum) at(t float64) (float64, bool) {
	size := p.amp
	if p.decay > 0 {
		size *= math.Exp(-t / p.decay.Seconds())
		if math.Abs(size) < settleHop {
			return 0, true
		}
	}
	return size * math.Sin(2*math.Pi*t/p.period.Seconds()), false
}

// Active reports whether p is still swinging.
func (p Pendulum) Active() bool { return p.run.Active() }

// Tick is the command for p's next frame, nil at rest.
func (p Pendulum) Tick() tea.Cmd { return p.run.Tick() }

// FrameAt is p's frame for at, for a test or a host that delivers frames itself.
func (p Pendulum) FrameAt(at time.Time) FrameMsg { return p.run.FrameAt(at) }

// Update swings on its own frames and asks for the next until it rests.
func (p Pendulum) Update(msg tea.Msg) (Pendulum, tea.Cmd) {
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

// Advance is p at now, and whether it came to rest.
func (p Pendulum) Advance(now time.Time) (Pendulum, bool) {
	if !p.Active() {
		return p, true
	}
	p.run, _ = p.run.Advance(now)
	pos, rested := p.at(frames.Paced(now.Sub(p.run.Start)).Seconds())
	p.Cur = pos
	if rested {
		p.run = frames.Run{}
	}
	return p, rested
}
