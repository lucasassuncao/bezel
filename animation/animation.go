// Package animation moves things on frames that arrive as messages: eased
// tweens, springs, colour fades, typed and scrolling text, a shake, and a
// Clock that drives several of them on one chain of ticks.
package animation

import (
	"math"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Frame is the interval between animation frames (~60fps). Terminal
// rows are coarse, so a tween resolves to only a handful of distinct heights;
// the frame rate exists to keep the easing curve smooth in time, not in space.
const Frame = 16 * time.Millisecond

// Tween eases an int between two values over a duration. Adapted from the
// minimize animation in github.com/Gaurav-Gosain/tuios (MIT), reduced to one
// axis. The zero value is inactive: At reports it finished, at To.
type Tween struct {
	From, To int
	Cur      int       // value sampled by the last advance; the height the frame draws
	Start    time.Time // zero until the first frame, which starts the clock
	Dur      time.Duration
	ease     Ease  // nil is EaseInOut
	id       int64 // which FrameMsg values are this tween's
	tag      int64 // which step: a twin frame from a second Tick is dropped
}

// FrameMsg is one tick of a tween's clock, at the time it fired. Hand it to
// every running tween's Update: each one takes only its own.
type FrameMsg struct {
	At      time.Time
	id, tag int64
}

// lastID numbers the tweens New builds, so a replaced one's frames are told apart.
var lastID atomic.Int64

// Tick is the command for t's next frame, nil for a tween with nothing to do.
func (t Tween) Tick() tea.Cmd {
	if !t.Active() {
		return nil
	}
	id, tag := t.id, t.tag
	return tea.Tick(Frame, func(at time.Time) tea.Msg { return FrameMsg{At: at, id: id, tag: tag} })
}

// FrameAt is t's frame for at, for a test harness that delivers frames itself.
func (t Tween) FrameAt(at time.Time) FrameMsg { return FrameMsg{At: at, id: t.id, tag: t.tag} }

// Update samples t on one of its own frames and asks for the next until it
// lands. Other messages, frames meant for another tween and a twin frame from
// a second Tick leave it alone.
func (t Tween) Update(msg tea.Msg) (Tween, tea.Cmd) {
	f, ok := msg.(FrameMsg)
	if !ok || f.id != t.id || f.tag != t.tag || !t.Active() {
		return t, nil
	}
	next, done := t.Advance(f.At)
	if done {
		return next, nil
	}
	next.tag++
	return next, next.Tick()
}

// Active reports whether t is currently in flight.
func (t Tween) Active() bool { return t.Dur > 0 }

// Advance is t sampled at now into Cur, and whether it finished, coming back
// inactive at To. The renderer reads Cur, not At, so one frame never mixes two
// samples and lays the hint panel out a row off from the space the preview left.
func (t Tween) Advance(now time.Time) (Tween, bool) {
	if t.Start.IsZero() {
		t.Start = now
	}
	var done bool
	t.Cur, done = t.At(now)
	if done {
		return Tween{From: t.To, To: t.To, Cur: t.To}, true
	}
	return t, false
}

// At returns the interpolated value at now, and whether the tween has finished.
// A finished tween reports its end value exactly, so the panel always lands on
// the height the static layout would have chosen. Before its clock starts it is at From.
func (t Tween) At(now time.Time) (int, bool) {
	if t.Dur <= 0 {
		return t.To, true
	}
	if t.Start.IsZero() {
		return t.From, false
	}
	p := float64(now.Sub(t.Start)) / float64(t.Dur)
	if p >= 1 {
		return t.To, true
	}
	if p < 0 {
		p = 0
	}
	return t.From + int(math.Round(float64(t.To-t.From)*t.ease.or()(p))), false
}

// New builds the tween taking a panel from its on-screen height to target
// over dur, its clock started by the first frame Tick delivers. When dur <= 0
// or from == target it is inactive and already at target: the toggle is instant.
func New(from, target int, dur time.Duration) Tween {
	if dur <= 0 || from == target {
		return Tween{From: target, To: target, Cur: target}
	}
	return Tween{From: from, To: target, Cur: from, Dur: dur, id: lastID.Add(1)}
}

// WithEase is t following e instead of EaseInOut.
func (t Tween) WithEase(e Ease) Tween {
	t.ease = e
	return t
}
