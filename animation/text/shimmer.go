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

// Shimmer sweeps a glow across text over and over: a placeholder that says
// something is loading, livelier than a spinner beside it.
type Shimmer struct {
	text       []string // the graphemes
	base, glow color.Color
	band       int     // cells from the glow's centre to where it fades out
	p          float64 // raw progress of the current sweep
	run        frames.Run
}

// NewShimmer is text in base with a glow sweeping across it once every dur,
// for as long as it lives. Plain text only; WithRepeat gives it an end.
func NewShimmer(text string, base, glow color.Color, dur time.Duration) Shimmer {
	s := Shimmer{text: edit.Graphemes(text), base: base, glow: glow, band: 4, run: frames.NewRun(dur)}
	s.run.Ease, s.run.Repeat = animation.Linear, animation.Forever
	return s
}

// WithBand widens or narrows the glow to n cells either side of its centre.
func (s Shimmer) WithBand(n int) Shimmer {
	s.band = max(n, 1)
	return s
}

// WithDelay holds the glow off the text for d before each first sweep.
func (s Shimmer) WithDelay(d time.Duration) Shimmer {
	s.run.Delay = max(d, 0)
	return s
}

// WithRepeat sweeps n more times after the first instead of forever.
func (s Shimmer) WithRepeat(n int) Shimmer {
	s.run.Repeat = max(n, animation.Forever)
	return s
}

// WithYoyo sends every other sweep back from the right.
func (s Shimmer) WithYoyo() Shimmer {
	s.run.Yoyo = true
	return s
}

// View is the text, each grapheme coloured by how near the glow it is.
func (s Shimmer) View() string {
	var b strings.Builder
	for i, g := range s.text {
		b.WriteString(lipgloss.NewStyle().Foreground(s.colorAt(i)).Render(g))
	}
	return b.String()
}

// colorAt is grapheme i's colour: base, warmed towards glow inside the band.
func (s Shimmer) colorAt(i int) color.Color {
	if !s.Active() {
		return s.base
	}
	centre := float64(-s.band) + s.p*float64(len(s.text)+2*s.band)
	k := 1 - math.Abs(float64(i)-centre)/float64(s.band)
	if k <= 0 {
		return s.base
	}
	return animation.Blend(s.base, s.glow, k)
}

// Active reports whether the glow is still sweeping.
func (s Shimmer) Active() bool { return s.run.Active() }

// Tick is the command for s's next frame, nil once it ended.
func (s Shimmer) Tick() tea.Cmd { return s.run.Tick() }

// FrameAt is s's frame for at, for a test or a host that delivers frames itself.
func (s Shimmer) FrameAt(at time.Time) animation.FrameMsg { return s.run.FrameAt(at) }

// Update sweeps on its own frames and asks for the next while it runs.
func (s Shimmer) Update(msg tea.Msg) (Shimmer, tea.Cmd) {
	f, ok := s.run.Own(msg)
	if !ok {
		return s, nil
	}
	if s, _ = s.Advance(f.At); !s.Active() {
		return s, nil
	}
	var cmd tea.Cmd
	s.run, cmd = s.run.Next()
	return s, cmd
}

// Advance is s at now, and whether its sweeps are over.
func (s Shimmer) Advance(now time.Time) (Shimmer, bool) {
	if !s.Active() {
		return s, true
	}
	s.run, s.p = s.run.Advance(now)
	return s, !s.Active()
}
