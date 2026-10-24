package text

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// waveLength is how many graphemes one crest and trough of a Wave span.
const waveLength = 10

// Wave bobs a line of text letter by letter, a wave travelling along it: a
// title that breathes, a "listening…" that moves.
type Wave struct {
	text []string // the graphemes
	amp  int
	p    float64
	run  frames.Run
}

// NewWave lifts text up to amp rows either way, a wave passing once every dur
// for as long as it lives; WithRepeat gives it an end. It draws 2·amp+1 rows.
func NewWave(text string, amp int, dur time.Duration) Wave {
	w := Wave{text: edit.Graphemes(text), amp: max(amp, 1), run: frames.NewRun(dur)}
	w.run.Ease, w.run.Repeat = animation.Linear, animation.Forever
	return w
}

// WithRepeat passes the wave n more times after the first instead of forever.
func (w Wave) WithRepeat(n int) Wave {
	w.run.Repeat = max(n, animation.Forever)
	return w
}

// View is the text on 2·amp+1 rows, each grapheme on the row the wave puts it;
// at rest every grapheme sits on the middle one.
func (w Wave) View() string {
	rows := make([]strings.Builder, 2*w.amp+1)
	for i, g := range w.text {
		off := 0
		if w.Active() && !w.run.Start.IsZero() {
			phase := 2 * math.Pi * (w.p - float64(i)/waveLength)
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
func (w Wave) Active() bool { return w.run.Active() }

// Tick is the command for w's next frame, nil once it ended.
func (w Wave) Tick() tea.Cmd { return w.run.Tick() }

// FrameAt is w's frame for at, for a test or a host that delivers frames itself.
func (w Wave) FrameAt(at time.Time) animation.FrameMsg { return w.run.FrameAt(at) }

// Update moves on its own frames and asks for the next while it runs.
func (w Wave) Update(msg tea.Msg) (Wave, tea.Cmd) {
	r, p, ok, cmd := w.run.Update(msg)
	if ok {
		w.run, w.p = r, p
	}
	return w, cmd
}

// Advance is w at now, and whether it ended.
func (w Wave) Advance(now time.Time) (Wave, bool) {
	if !w.Active() {
		return w, true
	}
	w.run, w.p = w.run.Advance(now)
	return w, !w.Active()
}
