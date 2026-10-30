package text

import (
	"image/color"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Rise floats a word up off its place and fades it out: "copied!", "+1", a
// "saved" said where it happened. Still, it is born gone.
type Rise struct {
	text   string
	rows   int
	fg, bg color.Color
	p      float64
	run    frames.Run
}

var _ bezel.Model[Rise] = Rise{}

// NewRise floats text rows rows up over dur, in the terminal's colour until
// WithFade.
func NewRise(text string, rows int, dur time.Duration) Rise {
	r := Rise{text: text, rows: max(rows, 1), run: frames.NewRun(dur)}
	r.run.Ease = animation.EaseOut
	if !r.Active() {
		r.p = 1
	}
	return r
}

// WithFade fades the text from fg into bg as it rises; bg is usually the
// background it rises over.
func (r Rise) WithFade(fg, bg color.Color) Rise {
	r.fg, r.bg = fg, bg
	return r
}

// View is rows+1 lines as wide as the text, the text on the line it has
// risen to; blank once it is gone.
func (r Rise) View() string {
	blank := strings.Repeat(" ", ansi.StringWidth(r.text))
	lines := make([]string, r.rows+1)
	for i := range lines {
		lines[i] = blank
	}
	if r.p < 1 {
		q := frames.Or(r.run.Ease)(r.p)
		line := r.rows - int(math.Round(q*float64(r.rows)))
		lines[line] = lipgloss.NewStyle().Foreground(animation.Blend(r.fg, r.bg, q)).Render(r.text)
	}
	return strings.Join(lines, "\n")
}

// Active reports whether the text is still rising.
func (r Rise) Active() bool { return r.run.Active() }

// Tick is the command for r's next frame, nil once it is gone.
func (r Rise) Tick() tea.Cmd { return r.run.Tick() }

// FrameAt is r's frame for at, for a test or a host that delivers frames itself.
func (r Rise) FrameAt(at time.Time) animation.FrameMsg { return r.run.FrameAt(at) }

// Update rises on its own frames and asks for the next until it is gone.
func (r Rise) Update(msg tea.Msg) (Rise, bezel.Result) {
	run, p, ok, cmd := r.run.Update(msg)
	if ok {
		r.run, r.p = run, p
	}
	return r, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is r at now, and whether it is gone.
func (r Rise) Advance(now time.Time) (Rise, bool) {
	if !r.Active() {
		return r, true
	}
	r.run, r.p = r.run.Advance(now)
	return r, !r.Active()
}
