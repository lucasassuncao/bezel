package animation

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/edit"
)

// swellLength is how many graphemes one crest and trough of a Swell span.
const swellLength = 10

// Swell bobs a line of text letter by letter, a wave travelling along it: a
// title that breathes, a "listening…" that moves.
type Swell struct {
	text []string // the graphemes
	amp  int
	p    float64
	run  run
}

// NewSwell lifts text up to amp rows either way, a wave passing once every dur
// for as long as it lives; WithRepeat gives it an end. It draws 2·amp+1 rows.
func NewSwell(text string, amp int, dur time.Duration) Swell {
	w := Swell{text: edit.Graphemes(text), amp: max(amp, 1), run: newRun(dur)}
	w.run.ease, w.run.repeat = Linear, Forever
	return w
}

// WithRepeat passes the wave n more times after the first instead of forever.
func (w Swell) WithRepeat(n int) Swell {
	w.run.repeat = max(n, Forever)
	return w
}

// View is the text on 2·amp+1 rows, each grapheme on the row the wave puts it;
// at rest every grapheme sits on the middle one.
func (w Swell) View() string {
	rows := make([]strings.Builder, 2*w.amp+1)
	for i, g := range w.text {
		off := 0
		if w.Active() && !w.run.start.IsZero() {
			phase := 2 * math.Pi * (w.p - float64(i)/swellLength)
			off = int(math.Round(float64(w.amp) * math.Sin(phase)))
		}
		pad := strings.Repeat(" ", ansi.StringWidth(g))
		for r := range rows {
			if r == w.amp-off {
				rows[r].WriteString(g)
			} else {
				rows[r].WriteString(pad)
			}
		}
	}
	lines := make([]string, len(rows))
	for i := range rows {
		lines[i] = rows[i].String()
	}
	return strings.Join(lines, "\n")
}

// Active reports whether the wave is still passing.
func (w Swell) Active() bool { return w.run.active() }

// Tick is the command for w's next frame, nil once it ended.
func (w Swell) Tick() tea.Cmd { return w.run.tick() }

// FrameAt is w's frame for at, for a test or a host that delivers frames itself.
func (w Swell) FrameAt(at time.Time) FrameMsg { return w.run.frameAt(at) }

// Update moves on its own frames and asks for the next while it runs.
func (w Swell) Update(msg tea.Msg) (Swell, tea.Cmd) {
	r, p, ok, cmd := w.run.update(msg)
	if ok {
		w.run, w.p = r, p
	}
	return w, cmd
}

// Advance is w at now, and whether it ended.
func (w Swell) Advance(now time.Time) (Swell, bool) {
	if !w.Active() {
		return w, true
	}
	w.run, w.p = w.run.advance(now)
	return w, !w.Active()
}
