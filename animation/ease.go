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
