// Package animation is the eased integer tween that drives panel
// transitions: a hint pane sliding open, a column growing.
package animation

import (
	"math"
	"time"
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
	Cur      int // value sampled by the last advance; the height the frame draws
	Start    time.Time
	Dur      time.Duration
}

// Active reports whether t is currently in flight.
func (t Tween) Active() bool { return t.Dur > 0 }

// Advance samples the curve at now into Cur and reports whether it finished.
// The renderer reads Cur, not At, so one frame never mixes two samples and
// lays the hint panel out a row off from the space the preview left.
func (t *Tween) Advance(now time.Time) bool {
	v, done := t.At(now)
	t.Cur = v
	return done
}

// At returns the interpolated value at now, and whether the tween has finished.
// A finished tween reports its end value exactly, so the panel always lands on
// the height the static layout would have chosen.
func (t Tween) At(now time.Time) (int, bool) {
	if t.Dur <= 0 {
		return t.To, true
	}
	p := float64(now.Sub(t.Start)) / float64(t.Dur)
	if p >= 1 {
		return t.To, true
	}
	if p < 0 {
		p = 0
	}
	return t.From + int(math.Round(float64(t.To-t.From)*easeInOutQuad(p))), false
}

// easeInOutQuad maps progress in [0,1] onto a quadratic ease-in-out, gentler
// than tuios' cubic on purpose: on whole cells a cubic freezes, then skips
// rows, while this never repeats a height more than once over 180ms.
func easeInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	p := -2*t + 2
	return 1 - p*p/2
}

// New builds the tween taking a panel from its on-screen height to target
// over dur. It returns the zero tween when dur <= 0 or from == target, so the
// caller falls through to an instant toggle.
func New(from, target int, dur time.Duration) Tween {
	if dur <= 0 || from == target {
		return Tween{}
	}
	return Tween{From: from, To: target, Cur: from, Start: time.Now(), Dur: dur}
}
