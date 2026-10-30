package animation

import (
	"image/color"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
)

// Fade moves a colour from one to another: a status that dims once read, a
// border that warms when focused, a match whose highlight cools.
type Fade struct {
	from, to color.Color
	mix      Float
}

var _ bezel.Model[Fade] = Fade{}

// NewFade goes from from to to over dur. A nil colour is the terminal's
// own, which cannot be mixed: Blend switches to the other half way.
func NewFade(from, to color.Color, dur time.Duration) Fade {
	return Fade{from: from, to: to, mix: NewFloat(0, 1, dur)}
}

// WithEase is f following e instead of InOut(Quad).
func (f Fade) WithEase(e Ease) Fade {
	f.mix = f.mix.WithEase(e)
	return f
}

// Value is the colour this frame draws with.
func (f Fade) Value() color.Color { return Blend(f.from, f.to, f.mix.cur) }

// Active reports whether f is still changing.
func (f Fade) Active() bool { return f.mix.Active() }

// Tick is the command for f's next frame, nil once it landed.
func (f Fade) Tick() tea.Cmd { return f.mix.Tick() }

// FrameAt is f's frame for at, for a test or a host that delivers frames itself.
func (f Fade) FrameAt(at time.Time) FrameMsg { return f.mix.FrameAt(at) }

// Update samples f on its own frames and asks for the next until it lands.
func (f Fade) Update(msg tea.Msg) (Fade, bezel.Result) {
	var r bezel.Result
	f.mix, r = f.mix.Update(msg)
	return f, r
}

// Advance is f sampled at now, and whether it landed.
func (f Fade) Advance(now time.Time) (Fade, bool) {
	var done bool
	f.mix, done = f.mix.Advance(now)
	return f, done
}

// Blend is the colour p of the way from a to b, channel by channel. With
// either one nil there is nothing to mix, so it is a until half way, then b.
func Blend(a, b color.Color, p float64) color.Color {
	if a == nil || b == nil {
		if p < 0.5 {
			return a
		}
		return b
	}
	p = min(max(p, 0), 1)
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	mix := func(x, y uint32) uint8 {
		return uint8(math.Round((float64(x>>8)*(1-p) + float64(y>>8)*p)))
	}
	return color.RGBA{mix(ar, br), mix(ag, bg), mix(ab, bb), 255}
}

// WithDelay holds f at its first colour for d after its first frame.
func (f Fade) WithDelay(d time.Duration) Fade {
	f.mix = f.mix.WithDelay(d)
	return f
}

// WithRepeat runs f n more times after the first, or without end for Forever.
func (f Fade) WithRepeat(n int) Fade {
	f.mix = f.mix.WithRepeat(n)
	return f
}

// WithYoyo sends every other run of f back: with Forever, a pulse.
func (f Fade) WithYoyo(on bool) Fade {
	f.mix = f.mix.WithYoyo(on)
	return f
}
