package text

import (
	"image/color"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/internal/hue"
)

// Rainbow spreads the colour wheel along text and turns it round, over and
// over: a celebration, a title that will not sit still.
type Rainbow struct {
	text   []string // the graphemes
	spread float64  // degrees of hue from one grapheme to the next
	s, l   float64  // the saturation and lightness every hue is drawn at
	p      float64
	run    frames.Run
}

// NewRainbow spreads the wheel along text, turning it once every dur for as
// long as it lives. Still, the colours stay where they start.
func NewRainbow(text string, dur time.Duration) Rainbow {
	r := Rainbow{text: edit.Graphemes(text), spread: 24, s: 0.75, l: 0.65, run: frames.NewRun(dur)}
	r.run.Ease, r.run.Repeat = animation.Linear, animation.Forever
	return r
}

// WithTone draws every hue at c's saturation and lightness, so the wheel
// matches a theme's accent instead of its own bright default.
func (r Rainbow) WithTone(c color.Color) Rainbow {
	_, r.s, r.l = hue.ToHSL(c)
	return r
}

// WithSpread puts degrees of hue between one grapheme and the next.
func (r Rainbow) WithSpread(degrees float64) Rainbow {
	r.spread = degrees
	return r
}

// View is the text, each grapheme its hue on the turning wheel.
func (r Rainbow) View() string {
	var b strings.Builder
	for i, g := range r.text {
		b.WriteString(lipgloss.NewStyle().Foreground(r.colorAt(i)).Render(g))
	}
	return b.String()
}

// colorAt is grapheme i's colour this frame.
func (r Rainbow) colorAt(i int) color.Color {
	h := math.Mod(float64(i)*r.spread-r.p*360, 360)
	if h < 0 {
		h += 360
	}
	return hue.FromHSL(h, r.s, r.l)
}

// Active reports whether the wheel is still turning.
func (r Rainbow) Active() bool { return r.run.Active() }

// Tick is the command for r's next frame, nil once it ended.
func (r Rainbow) Tick() tea.Cmd { return r.run.Tick() }

// FrameAt is r's frame for at, for a test or a host that delivers frames itself.
func (r Rainbow) FrameAt(at time.Time) animation.FrameMsg { return r.run.FrameAt(at) }

// Update turns on its own frames and asks for the next while it runs.
func (r Rainbow) Update(msg tea.Msg) (Rainbow, tea.Cmd) {
	run, p, ok, cmd := r.run.Update(msg)
	if ok {
		r.run, r.p = run, p
	}
	return r, cmd
}

// Advance is r at now, and whether it ended.
func (r Rainbow) Advance(now time.Time) (Rainbow, bool) {
	if !r.Active() {
		return r, true
	}
	r.run, r.p = r.run.Advance(now)
	return r, !r.Active()
}
