// Package animation moves values on frames that arrive as messages: tweens,
// springs, fades and physics, alone or on a shared Clock. Its subpackages
// animate text and scenes. BEZEL_MOTION sets the speed.
package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Frame is the interval between animation frames (~60fps). Terminal
// rows are coarse, so a tween resolves to only a handful of distinct heights;
// the frame rate exists to keep the easing curve smooth in time, not in space.
const Frame = frames.Frame

// Tween eases an int between two values over a duration. Adapted from the
// minimize animation in github.com/Gaurav-Gosain/tuios (MIT), reduced to one
// axis. The zero value is inactive: At reports it finished, at To.
type Tween struct {
	From, To int
	Cur      int       // value sampled by the last advance; the height the frame draws
	Start    time.Time // zero until the first frame, which starts the clock
	Dur      time.Duration
	loop     frames.Loop
	ease     Ease  // nil is EaseInOut
	id       int64 // which FrameMsg values are this tween's
	tag      int64 // which step: a twin frame from a second Tick is dropped
}

// FrameMsg is one tick of an animation's clock, at the time it fired. Hand it
// to every running animation's Update: each one takes only its own.
type FrameMsg = frames.Msg

// Tick is the command for t's next frame, nil for a tween with nothing to do.
func (t Tween) Tick() tea.Cmd {
	if !t.Active() {
		return nil
	}
	return frames.Tick(Frame, t.id, t.tag)
}

// FrameAt is t's frame for at, for a test harness that delivers frames itself.
func (t Tween) FrameAt(at time.Time) FrameMsg { return frames.NewMsg(at, t.id, t.tag) }

// Update samples t on one of its own frames and asks for the next until it
// lands. Other messages, frames meant for another tween and a twin frame from
// a second Tick leave it alone.
func (t Tween) Update(msg tea.Msg) (Tween, tea.Cmd) {
	f, ok := frames.Is(msg, t.id, t.tag)
	if !ok || !t.Active() {
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
// inactive where it landed. The renderer reads Cur, not At, so one frame never mixes two
// samples and lays the hint panel out a row off from the space the preview left.
func (t Tween) Advance(now time.Time) (Tween, bool) {
	if t.Start.IsZero() {
		t.Start = now
	}
	var done bool
	t.Cur, done = t.At(now)
	if done {
		return Tween{From: t.Cur, To: t.Cur, Cur: t.Cur}, true
	}
	return t, false
}

// At returns the interpolated value at now, and whether the tween has finished.
// A finished tween reports its end exactly, To or a yoyo's From, so a panel
// lands on the height the static layout would choose. Unstarted, it is at From.
func (t Tween) At(now time.Time) (int, bool) {
	if t.Dur <= 0 {
		return t.To, true
	}
	if t.Start.IsZero() {
		return t.From, false
	}
	p, done := t.loop.Phase(frames.Paced(now.Sub(t.Start)), t.Dur)
	switch {
	case done && p == 1:
		return t.To, true
	case done:
		return t.From, true
	}
	return t.From + int(math.Round(float64(t.To-t.From)*frames.Or(t.ease)(p))), false
}

// NewTween builds the tween taking a panel from its on-screen height to target
// over dur, its clock started by the first frame Tick delivers. When dur <= 0
// or from == target, or motion is Still, it is inactive and already at target.
func NewTween(from, target int, dur time.Duration) Tween {
	if dur <= 0 || from == target || Still() {
		return Tween{From: target, To: target, Cur: target}
	}
	return Tween{From: from, To: target, Cur: from, Dur: dur, id: frames.NextID()}
}

// WithEase is t following e instead of EaseInOut.
func (t Tween) WithEase(e Ease) Tween {
	t.ease = e
	return t
}

// WithDelay holds t at From for d after its first frame.
func (t Tween) WithDelay(d time.Duration) Tween {
	t.loop.Delay = max(d, 0)
	return t
}

// WithRepeat runs t n more times after the first, or without end for Forever.
func (t Tween) WithRepeat(n int) Tween {
	t.loop.Repeat = max(n, Forever)
	return t
}

// WithYoyo sends every other run of t back the way it came.
func (t Tween) WithYoyo(on bool) Tween {
	t.loop.Yoyo = on
	return t
}
