package text

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/internal/noise"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// fallShare is the share of the run one letter takes to drop and settle; the
// rest is spread over when the letters let go.
const fallShare = 0.5

// Fall drops the letters of a word from above, each letting go at its own
// moment, and lands them in place with a bounce: a title that arrives, a
// word that piles up out of nothing. Still, it is born landed.
type Fall struct {
	text []string
	rows int
	p    float64
	run  frames.Run
}

// NewFall drops text from rows rows above its line over dur.
func NewFall(text string, rows int, dur time.Duration) Fall {
	f := Fall{text: edit.Graphemes(text), rows: max(rows, 1), run: frames.NewRun(dur)}
	f.run.Ease = animation.Linear
	if !f.Active() {
		f.p = 1
	}
	return f
}

// height is how many rows above its line grapheme i is now; -1 before it
// has let go.
func (f Fall) height(i int) int {
	letGo := noise.Unit(noise.Hash(11, uint64(i))) * (1 - fallShare) // #nosec G115 -- a hash key
	q := (f.p - letGo) / fallShare
	switch {
	case q < 0:
		return -1
	case q >= 1:
		return 0
	}
	return int(math.Round(float64(f.rows) * (1 - animation.Out(animation.Bounce)(q))))
}

// View is rows+1 lines as wide as the text, each letter on the row it has
// fallen to; the last line is where they land.
func (f Fall) View() string {
	lines := make([][]string, f.rows+1)
	for y := range lines {
		lines[y] = make([]string, len(f.text))
		for i, g := range f.text {
			lines[y][i] = strings.Repeat(" ", ansi.StringWidth(g))
		}
	}
	for i, g := range f.text {
		if h := f.height(i); h >= 0 {
			lines[f.rows-h][i] = g
		}
	}
	out := make([]string, len(lines))
	for y, l := range lines {
		out[y] = strings.Join(l, "")
	}
	return strings.Join(out, "\n")
}

// Active reports whether any letter is still falling.
func (f Fall) Active() bool { return f.run.Active() }

// Tick is the command for f's next frame, nil once every letter landed.
func (f Fall) Tick() tea.Cmd { return f.run.Tick() }

// FrameAt is f's frame for at, for a test or a host that delivers frames itself.
func (f Fall) FrameAt(at time.Time) animation.FrameMsg { return f.run.FrameAt(at) }

// Update drops on its own frames and asks for the next until every letter lands.
func (f Fall) Update(msg tea.Msg) (Fall, tea.Cmd) {
	r, p, ok, cmd := f.run.Update(msg)
	if ok {
		f.run, f.p = r, p
	}
	return f, cmd
}

// Advance is f at now, and whether every letter landed.
func (f Fall) Advance(now time.Time) (Fall, bool) {
	if !f.Active() {
		return f, true
	}
	f.run, f.p = f.run.Advance(now)
	return f, !f.Active()
}

// WithDelay holds every letter up for d after the first frame.
func (f Fall) WithDelay(d time.Duration) Fall {
	f.run.Delay = max(d, 0)
	return f
}
