package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Ease maps progress in [0,1] onto how far along the motion is: 0 at the start
// and 1 at the end, and between them a lively one, Back or Elastic, may leave [0,1].
type Ease func(p float64) float64

// Linear moves at one speed throughout: a counter, a typewriter.
func Linear(p float64) float64 { return p }

// EaseIn starts slow and ends fast, for something leaving.
func EaseIn(p float64) float64 { return p * p }

// EaseOut starts fast and settles, for something arriving.
func EaseOut(p float64) float64 { return 1 - (1-p)*(1-p) }

// EaseInOut is quadratic, gentler than a cubic on purpose: on whole cells a
// cubic freezes, then skips rows, while this never repeats a height more
// than once over 180ms. Every tween follows it unless told otherwise.
func EaseInOut(p float64) float64 {
	if p < 0.5 {
		return 2 * p * p
	}
	q := -2*p + 2
	return 1 - q*q/2
}

// or is e, or EaseInOut for nil.
func (e Ease) or() Ease {
	if e == nil {
		return EaseInOut
	}
	return e
}

// run is the clock a timed animation keeps: when its first frame came, how
// long it lasts, its curve, and which frames are its own.
type run struct {
	loop
	start   time.Time
	dur     time.Duration
	ease    Ease
	id, tag int64
}

func newRun(dur time.Duration) run {
	if dur <= 0 || Still() {
		return run{}
	}
	return run{dur: dur, id: lastID.Add(1)}
}

func (r run) active() bool { return r.dur > 0 }

func (r run) frameAt(at time.Time) FrameMsg { return FrameMsg{At: at, id: r.id, tag: r.tag} }

func (r run) tick() tea.Cmd {
	if !r.active() {
		return nil
	}
	id, tag := r.id, r.tag
	return tea.Tick(Frame, func(at time.Time) tea.Msg { return FrameMsg{At: at, id: id, tag: tag} })
}

// own is the frame in msg when it is this run's current one.
func (r run) own(msg tea.Msg) (FrameMsg, bool) {
	f, ok := msg.(FrameMsg)
	return f, ok && r.active() && f.id == r.id && f.tag == r.tag
}

// advance is r at now and its raw progress in [0,1]; the first call starts
// the clock, and once its runs are over it comes back spent.
func (r run) advance(now time.Time) (run, float64) {
	if !r.active() {
		return r, 1
	}
	if r.start.IsZero() {
		r.start = now
	}
	p, done := r.phase(now.Sub(r.start), r.dur)
	if done {
		return run{}, p
	}
	return r, p
}

// next is r ready for its following frame, with the command that brings it.
func (r run) next() (run, tea.Cmd) {
	r.tag++
	return r, r.tick()
}

// Forever is the repeat count of a motion that runs until it is replaced.
const Forever = -1

// loop is when a timed motion begins and how often it runs again.
type loop struct {
	delay  time.Duration
	repeat int  // runs after the first; Forever for no end
	yoyo   bool // every other run goes back
}

// phase is the raw progress after elapsed, and whether every run is over: 0
// through the delay, a yoyo's odd runs backwards, and an exact 0 or 1 at the end.
func (l loop) phase(elapsed, dur time.Duration) (float64, bool) {
	elapsed -= l.delay
	if elapsed < 0 {
		return 0, false
	}
	n := int(elapsed / dur)
	back := func(n int) bool { return l.yoyo && n%2 == 1 }
	if l.repeat != Forever && n > l.repeat {
		if back(l.repeat) {
			return 0, true
		}
		return 1, true
	}
	p := float64(elapsed%dur) / float64(dur)
	if back(n) {
		p = 1 - p
	}
	return p, false
}

// The curves below ease in, slow first; Out and InOut turn any one round.

// Sine eases in along a quarter of a sine wave, the gentlest of them.
func Sine(p float64) float64 { return 1 - math.Cos(p*math.Pi/2) }

// Cubic is p³, slower to start than EaseIn's p².
func Cubic(p float64) float64 { return p * p * p }

// Quart is p⁴.
func Quart(p float64) float64 { return p * p * p * p }

// Quint is p⁵.
func Quint(p float64) float64 { return p * p * p * p * p }

// Expo doubles its pace every tenth of the way: almost still, then gone.
func Expo(p float64) float64 {
	if p <= 0 {
		return 0
	}
	return math.Pow(2, 10*p-10)
}

// Circ eases along a quarter circle.
func Circ(p float64) float64 { return 1 - math.Sqrt(1-p*p) }

// backPull is how far Back pulls away, the common 10% or so.
const backPull = 1.70158

// Back pulls back a little before it goes; Out(Back) overshoots and settles.
func Back(p float64) float64 { return (backPull+1)*p*p*p - backPull*p*p }

// Elastic winds up like a spring before it lets go; Out(Elastic) rings past
// its end and settles.
func Elastic(p float64) float64 {
	if p <= 0 || p >= 1 {
		return math.Round(p)
	}
	return -math.Pow(2, 10*p-10) * math.Sin((10*p-10.75)*2*math.Pi/3)
}

// Bounce hops off the start a few times before it goes; Out(Bounce) drops
// onto its end like a ball, and never leaves [0,1].
func Bounce(p float64) float64 { return 1 - bounceOut(1-p) }

// bounceOut is the ball falling onto its end: four arcs, each lower.
func bounceOut(p float64) float64 {
	const n, d = 7.5625, 2.75
	switch {
	case p < 1/d:
		return n * p * p
	case p < 2/d:
		p -= 1.5 / d
		return n*p*p + 0.75
	case p < 2.5/d:
		p -= 2.25 / d
		return n*p*p + 0.9375
	default:
		p -= 2.625 / d
		return n*p*p + 0.984375
	}
}

// Out is e played backwards from the end: fast first, then settling.
func Out(e Ease) Ease { return func(p float64) float64 { return 1 - e(1-p) } }

// InOut is e easing in over the first half and out over the second.
func InOut(e Ease) Ease {
	return func(p float64) float64 {
		if p < 0.5 {
			return e(2*p) / 2
		}
		return 1 - e(2-2*p)/2
	}
}

// Steps moves in n equal jumps, like a clock's hand; none is one jump to the end.
func Steps(n int) Ease {
	return func(p float64) float64 {
		if n <= 0 || p >= 1 {
			return 1
		}
		return math.Floor(p*float64(n)) / float64(n)
	}
}
