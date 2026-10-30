package text

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Caret blinks a cursor: shown for the first half of every period, hidden for
// the second, for as long as it lives. Still, it is always shown.
type Caret struct {
	glyph string
	p     float64
	run   frames.Run
}

var _ bezel.Model[Caret] = Caret{}

// NewCaret blinks glyph, "▌" or "_" say, once every period.
func NewCaret(glyph string, period time.Duration) Caret {
	c := Caret{glyph: glyph, run: frames.NewRun(period)}
	c.run.Repeat = animation.Forever
	return c
}

// View is the glyph, or as many blanks while it is hidden, so text after it
// holds still.
func (c Caret) View() string {
	if c.Active() && c.p >= 0.5 {
		return strings.Repeat(" ", ansi.StringWidth(c.glyph))
	}
	return c.glyph
}

// Active reports whether the caret is still blinking.
func (c Caret) Active() bool { return c.run.Active() }

// Tick is the command for c's next frame, nil once it stopped.
func (c Caret) Tick() tea.Cmd { return c.run.Tick() }

// FrameAt is c's frame for at, for a test or a host that delivers frames itself.
func (c Caret) FrameAt(at time.Time) animation.FrameMsg { return c.run.FrameAt(at) }

// Update blinks on its own frames and asks for the next while it runs.
func (c Caret) Update(msg tea.Msg) (Caret, bezel.Result) {
	r, p, ok, cmd := c.run.Update(msg)
	if ok {
		c.run, c.p = r, p
	}
	return c, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is c at now, and whether it stopped.
func (c Caret) Advance(now time.Time) (Caret, bool) {
	if !c.Active() {
		return c, true
	}
	c.run, c.p = c.run.Advance(now)
	return c, !c.Active()
}
