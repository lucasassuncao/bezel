package animation

import (
	"cmp"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Key is one stop of Keyframes: the value at a share of the duration, and
// the curve that leads into it from the stop before.
type Key struct {
	At    float64 // in [0,1]
	Value float64
	Ease  Ease // nil is EaseInOut
}

// Keyframes moves a value through several stops instead of one: a bounce
// that lands, rises and lands again, or a dial that overshoots twice.
type Keyframes struct {
	keys []Key
	Cur  float64 // the value the last frame sampled
	run  frames.Run
}

var _ bezel.Model[Keyframes] = Keyframes{}

// NewKeyframes plays keys over dur, sorted by At. With motion Still, no
// time or fewer than two keys it is inactive at the last key's value.
func NewKeyframes(dur time.Duration, keys ...Key) Keyframes {
	k := Keyframes{keys: slices.Clone(keys)}
	slices.SortStableFunc(k.keys, func(a, b Key) int { return cmp.Compare(a.At, b.At) })
	if len(k.keys) == 0 {
		return k
	}
	k.Cur = k.keys[len(k.keys)-1].Value
	if len(k.keys) < 2 {
		return k
	}
	if k.run = frames.NewRun(dur); k.run.Active() {
		k.Cur = k.keys[0].Value
	}
	return k
}

// at is the value at progress p.
func (k Keyframes) at(p float64) float64 {
	if p <= k.keys[0].At {
		return k.keys[0].Value
	}
	for i := 1; i < len(k.keys); i++ {
		a, b := k.keys[i-1], k.keys[i]
		if p > b.At {
			continue
		}
		q := 1.0
		if b.At > a.At {
			q = (p - a.At) / (b.At - a.At)
		}
		return a.Value + (b.Value-a.Value)*frames.Or(b.Ease)(q)
	}
	return k.keys[len(k.keys)-1].Value
}

// WithDelay holds k at its first key for d after its first frame.
func (k Keyframes) WithDelay(d time.Duration) Keyframes {
	k.run.Delay = max(d, 0)
	return k
}

// WithRepeat plays k n more times after the first, or without end for Forever.
func (k Keyframes) WithRepeat(n int) Keyframes {
	k.run.Repeat = max(n, Forever)
	return k
}

// WithYoyo plays every other pass of k backwards.
func (k Keyframes) WithYoyo(on bool) Keyframes {
	k.run.Yoyo = on
	return k
}

// Active reports whether k is still moving.
func (k Keyframes) Active() bool { return k.run.Active() }

// Tick is the command for k's next frame, nil once it ended.
func (k Keyframes) Tick() tea.Cmd { return k.run.Tick() }

// FrameAt is k's frame for at, for a test or a host that delivers frames itself.
func (k Keyframes) FrameAt(at time.Time) FrameMsg { return k.run.FrameAt(at) }

// Update moves on its own frames and asks for the next until it ends.
func (k Keyframes) Update(msg tea.Msg) (Keyframes, bezel.Result) {
	f, ok := k.run.Own(msg)
	if !ok {
		return k, bezel.Result{Handled: ok}
	}
	if k, _ = k.Advance(f.At); !k.Active() {
		return k, bezel.Result{Handled: true}
	}
	var cmd tea.Cmd
	k.run, cmd = k.run.Next()
	return k, bezel.Result{Handled: true, Cmd: cmd}
}

// Advance is k at now, and whether it ended.
func (k Keyframes) Advance(now time.Time) (Keyframes, bool) {
	if !k.Active() {
		return k, true
	}
	var p float64
	k.run, p = k.run.Advance(now)
	k.Cur = k.at(p)
	return k, !k.Active()
}
