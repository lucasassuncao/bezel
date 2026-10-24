package frames

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Ease maps progress in [0,1] onto how far along the motion is.
type Ease func(p float64) float64

// InOut is the quadratic curve every motion follows unless told otherwise.
func InOut(p float64) float64 {
	if p < 0.5 {
		return 2 * p * p
	}
	q := -2*p + 2
	return 1 - q*q/2
}

// Or is e, or InOut for nil.
func Or(e Ease) Ease {
	if e == nil {
		return InOut
	}
	return e
}

// Run is the clock a timed animation keeps: when its first frame came, how
// long it lasts, its curve, and which frames are its own.
type Run struct {
	Loop
	Start   time.Time
	Dur     time.Duration
	Ease    Ease
	ID, Tag int64
}

// NewRun is a run of dur; inactive when dur <= 0 or motion is Still.
func NewRun(dur time.Duration) Run {
	if dur <= 0 || Still() {
		return Run{}
	}
	return Run{Dur: dur, ID: NextID()}
}

// Active reports whether r still runs.
func (r Run) Active() bool { return r.Dur > 0 }

// FrameAt is r's current frame for at.
func (r Run) FrameAt(at time.Time) Msg { return NewMsg(at, r.ID, r.Tag) }

// Tick is the command for r's next frame, nil once it stopped.
func (r Run) Tick() tea.Cmd {
	if !r.Active() {
		return nil
	}
	return Tick(Frame, r.ID, r.Tag)
}

// Own is the frame in msg when it is this run's current one.
func (r Run) Own(msg tea.Msg) (Msg, bool) {
	f, ok := Is(msg, r.ID, r.Tag)
	return f, ok && r.Active()
}

// Advance is r at now and its raw progress in [0,1]; the first call starts
// the clock, and once its runs are over it comes back spent.
func (r Run) Advance(now time.Time) (Run, float64) {
	if !r.Active() {
		return r, 1
	}
	if r.Start.IsZero() {
		r.Start = now
	}
	p, done := r.Phase(Paced(now.Sub(r.Start)), r.Dur)
	if done {
		return Run{}, p
	}
	return r, p
}

// Next is r ready for its following frame, with the command that brings it.
func (r Run) Next() (Run, tea.Cmd) {
	r.Tag++
	return r, r.Tick()
}

// Update advances r on msg when it is r's own frame: the run after it, its
// progress, whether msg was r's, and the command for the next frame.
func (r Run) Update(msg tea.Msg) (Run, float64, bool, tea.Cmd) {
	f, ok := r.Own(msg)
	if !ok {
		return r, 0, false, nil
	}
	r, p := r.Advance(f.At)
	if !r.Active() {
		return r, p, true, nil
	}
	r, cmd := r.Next()
	return r, p, true, cmd
}

// Forever is the repeat count of a motion that runs until it is replaced.
const Forever = -1

// Loop is when a timed motion begins and how often it runs again.
type Loop struct {
	Delay  time.Duration
	Repeat int  // runs after the first; Forever for no end
	Yoyo   bool // every other run goes back
}

// Phase is the raw progress after elapsed, and whether every run is over: 0
// through the delay, a yoyo's odd runs backwards, and an exact 0 or 1 at the end.
func (l Loop) Phase(elapsed, dur time.Duration) (float64, bool) {
	elapsed -= l.Delay
	if elapsed < 0 {
		return 0, false
	}
	n := int(elapsed / dur)
	back := func(n int) bool { return l.Yoyo && n%2 == 1 }
	if l.Repeat != Forever && n > l.Repeat {
		if back(l.Repeat) {
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
