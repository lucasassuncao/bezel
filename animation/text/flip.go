package text

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// flipping is what a letter shows as it turns over, after the old letter and
// before the new one: the card's top falling, its edge, its bottom landing.
var flipping = []string{"▀", "━", "▄"}

// flipTurn is the share of the run one letter takes to turn over; the rest
// is how far apart the columns start.
const flipTurn = 0.4

// Flip turns over each letter that changes, left to right, like the cards of
// a scoreboard: a score, a status word, a clock that changes in place. The
// letters that stay the same do not move.
type Flip struct {
	from, to []string // padded to the same width
	dur      time.Duration
	p        float64
	run      frames.Run
}

var _ bezel.Model[Flip] = Flip{}

// NewFlip turns from into to over dur.
func NewFlip(from, to string, dur time.Duration) Flip {
	a, b := edit.Graphemes(from), edit.Graphemes(to)
	for len(a) < len(b) {
		a = append(a, " ")
	}
	for len(b) < len(a) {
		b = append(b, " ")
	}
	f := Flip{from: a, to: b, dur: dur}
	if strings.Join(a, "") != strings.Join(b, "") {
		f.run = frames.NewRun(dur)
	}
	if !f.Active() {
		f.p = 1
	}
	return f
}

// To flips on from the text f is heading for to next, over the same duration.
func (f Flip) To(next string) (Flip, tea.Cmd) {
	n := NewFlip(strings.TrimRight(strings.Join(f.to, ""), " "), next, f.dur)
	return n, n.Tick()
}

// View is the text with each changing letter at its point in the turn.
func (f Flip) View() string {
	var b strings.Builder
	n := len(f.to)
	for i := range f.to {
		if f.from[i] == f.to[i] {
			b.WriteString(f.to[i])
			continue
		}
		start := 0.0
		if n > 1 {
			start = (1 - flipTurn) * float64(i) / float64(n-1)
		}
		q := (f.p - start) / flipTurn
		switch {
		case q <= 0:
			b.WriteString(f.from[i])
		case q >= 1:
			b.WriteString(f.to[i])
		default:
			b.WriteString(flipping[min(int(q*float64(len(flipping))), len(flipping)-1)])
		}
	}
	return b.String()
}

// Active reports whether any letter is still turning.
func (f Flip) Active() bool { return f.run.Active() }

// Tick is the command for f's next frame, nil once every letter landed.
func (f Flip) Tick() tea.Cmd { return f.run.Tick() }

// FrameAt is f's frame for at, for a test or a host that delivers frames itself.
func (f Flip) FrameAt(at time.Time) animation.FrameMsg { return f.run.FrameAt(at) }

// Update turns on its own frames and asks for the next until every letter lands.
func (f Flip) Update(msg tea.Msg) (Flip, bezel.Result) {
	r, p, ok, cmd := f.run.Update(msg)
	if ok {
		f.run, f.p = r, p
	}
	return f, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is f at now, and whether every letter landed.
func (f Flip) Advance(now time.Time) (Flip, bool) {
	if !f.Active() {
		return f, true
	}
	f.run, f.p = f.run.Advance(now)
	return f, !f.Active()
}

// WithDelay holds the old text for d after the first frame.
func (f Flip) WithDelay(d time.Duration) Flip {
	f.run.Delay = max(d, 0)
	return f
}
