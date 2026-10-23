package animation

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Ease maps progress in [0,1] onto how far along the motion is, also in [0,1].
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
	start   time.Time
	dur     time.Duration
	ease    Ease
	id, tag int64
}

func newRun(dur time.Duration) run {
	if dur <= 0 {
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
// the clock, and at 1 the run comes back spent.
func (r run) advance(now time.Time) (run, float64) {
	if !r.active() {
		return r, 1
	}
	if r.start.IsZero() {
		r.start = now
	}
	p := float64(now.Sub(r.start)) / float64(r.dur)
	if p >= 1 {
		return run{}, 1
	}
	return r, max(p, 0)
}

// next is r ready for its following frame, with the command that brings it.
func (r run) next() (run, tea.Cmd) {
	r.tag++
	return r, r.tick()
}
