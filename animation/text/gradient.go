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
)

// Gradient spreads colours round a text and lets them flow along it: a title
// or a banner that is alive while the screen waits.
type Gradient struct {
	text   []string // the graphemes
	colors []color.Color
	p      float64
	run    frames.Run
}

// NewGradient spreads colors evenly round text, back to the first at the
// end, and turns them once every dur for as long as it lives. Still, it is
// drawn without moving.
func NewGradient(text string, colors []color.Color, dur time.Duration) Gradient {
	g := Gradient{text: edit.Graphemes(text), colors: colors, run: frames.NewRun(dur)}
	g.run.Repeat = animation.Forever
	return g
}

// WithRepeat turns g n more times after the first instead of forever.
func (g Gradient) WithRepeat(n int) Gradient {
	g.run.Repeat = max(n, animation.Forever)
	return g
}

// WithYoyo sends every other turn back the way it came.
func (g Gradient) WithYoyo() Gradient {
	g.run.Yoyo = true
	return g
}

// View is the text, each grapheme in its colour now.
func (g Gradient) View() string {
	if len(g.colors) == 0 {
		return strings.Join(g.text, "")
	}
	var b strings.Builder
	for i, s := range g.text {
		b.WriteString(lipgloss.NewStyle().Foreground(g.colorAt(i)).Render(s))
	}
	return b.String()
}

// colorAt is grapheme i's colour: its place round the text, less how far the
// colours have flowed, read off the ring of stops.
func (g Gradient) colorAt(i int) color.Color {
	x := float64(i)/float64(len(g.text)) - g.p
	x -= math.Floor(x)
	pos := x * float64(len(g.colors))
	k := int(pos)
	return animation.Blend(g.colors[k%len(g.colors)], g.colors[(k+1)%len(g.colors)], pos-float64(k))
}

// Active reports whether the colours are still flowing.
func (g Gradient) Active() bool { return g.run.Active() }

// Tick is the command for g's next frame, nil once it stopped.
func (g Gradient) Tick() tea.Cmd { return g.run.Tick() }

// FrameAt is g's frame for at, for a test or a host that delivers frames itself.
func (g Gradient) FrameAt(at time.Time) animation.FrameMsg { return g.run.FrameAt(at) }

// Update flows on its own frames and asks for the next while it runs.
func (g Gradient) Update(msg tea.Msg) (Gradient, tea.Cmd) {
	f, ok := g.run.Own(msg)
	if !ok {
		return g, nil
	}
	if g, _ = g.Advance(f.At); !g.Active() {
		return g, nil
	}
	var cmd tea.Cmd
	g.run, cmd = g.run.Next()
	return g, cmd
}

// Advance is g at now, and whether it stopped.
func (g Gradient) Advance(now time.Time) (Gradient, bool) {
	if !g.Active() {
		return g, true
	}
	g.run, g.p = g.run.Advance(now)
	return g, !g.Active()
}
