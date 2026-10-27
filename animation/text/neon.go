package text

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/internal/noise"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// neonSlot is how long one flicker lasts, and how often the sign rolls for one.
const neonSlot = 50 * time.Millisecond

// Neon lights text like a neon sign: steady, but now and then the whole sign
// stutters, and one weak letter flickers more often than the rest. The same
// seed flickers the same way every time; Still, it never flickers.
type Neon struct {
	text    []string
	on, off color.Color
	weak    int // the letter with the failing tube
	seed    uint64
	slot    int64 // how many neonSlots since the first frame
	run     frames.Run
}

// NewNeon lights text in on, dropping towards off when it flickers.
func NewNeon(text string, on, off color.Color) Neon {
	n := Neon{text: edit.Graphemes(text), on: on, off: off, seed: 1, run: frames.NewRun(time.Second)}
	n.run.Repeat = animation.Forever
	return n.WithSeed(1)
}

// WithSeed flickers another way, with another weak letter.
func (n Neon) WithSeed(seed uint64) Neon {
	n.seed = seed
	n.weak = -1
	if len(n.text) > 0 {
		n.weak = int(noise.Hash(seed, 0) % uint64(len(n.text))) // #nosec G115 -- below the text's length
	}
	return n
}

// lit is how bright grapheme i is now, 1 fully lit.
func (n Neon) lit(i int) float64 {
	if !n.Active() {
		return 1
	}
	s := uint64(n.slot) // #nosec G115 -- a non-negative count
	if noise.Unit(noise.Hash(n.seed, s*2+1)) < 0.03 {
		return 0.2 // the whole sign stutters
	}
	if i == n.weak && noise.Unit(noise.Hash(n.seed, s*2+2)) < 0.18 {
		return 0.35
	}
	return 1
}

// View is the text, each letter in its brightness now.
func (n Neon) View() string {
	var b strings.Builder
	for i, g := range n.text {
		b.WriteString(lipgloss.NewStyle().Foreground(animation.Blend(n.off, n.on, n.lit(i))).Render(g))
	}
	return b.String()
}

// Active reports whether the sign is still flickering.
func (n Neon) Active() bool { return n.run.Active() }

// Tick is the command for n's next frame, nil once it stopped.
func (n Neon) Tick() tea.Cmd { return n.run.Tick() }

// FrameAt is n's frame for at, for a test or a host that delivers frames itself.
func (n Neon) FrameAt(at time.Time) animation.FrameMsg { return n.run.FrameAt(at) }

// Update flickers on its own frames and asks for the next while it lives.
func (n Neon) Update(msg tea.Msg) (Neon, tea.Cmd) {
	f, ok := n.run.Own(msg)
	if !ok {
		return n, nil
	}
	n, _ = n.Advance(f.At)
	var cmd tea.Cmd
	n.run, cmd = n.run.Next()
	return n, cmd
}

// Advance is n at now; a sign never ends on its own, so done is false while it lives.
func (n Neon) Advance(now time.Time) (Neon, bool) {
	if !n.Active() {
		return n, true
	}
	n.run, _ = n.run.Advance(now)
	n.slot = int64(frames.Paced(now.Sub(n.run.Start)) / neonSlot)
	return n, false
}
