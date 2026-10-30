package text

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Ellipsis grows a row of dots and starts over: "loading", "saving", waiting
// without a spinner. It is always as wide as its dots, so what follows holds
// still. Still, every dot is shown.
type Ellipsis struct {
	dots int
	p    float64
	run  frames.Run
}

var _ bezel.Model[Ellipsis] = Ellipsis{}

// NewEllipsis grows up to dots dots, one round every period.
func NewEllipsis(dots int, period time.Duration) Ellipsis {
	e := Ellipsis{dots: max(dots, 1), run: frames.NewRun(period)}
	e.run.Repeat = animation.Forever
	return e
}

// View is the dots shown so far, padded to the full width.
func (e Ellipsis) View() string {
	n := e.dots
	if e.Active() {
		n = min(int(e.p*float64(e.dots+1)), e.dots)
	}
	return strings.Repeat(".", n) + strings.Repeat(" ", e.dots-n)
}

// Active reports whether the dots are still growing.
func (e Ellipsis) Active() bool { return e.run.Active() }

// Tick is the command for e's next frame, nil once it stopped.
func (e Ellipsis) Tick() tea.Cmd { return e.run.Tick() }

// FrameAt is e's frame for at, for a test or a host that delivers frames itself.
func (e Ellipsis) FrameAt(at time.Time) animation.FrameMsg { return e.run.FrameAt(at) }

// Update grows on its own frames and asks for the next while it runs.
func (e Ellipsis) Update(msg tea.Msg) (Ellipsis, bezel.Result) {
	r, p, ok, cmd := e.run.Update(msg)
	if ok {
		e.run, e.p = r, p
	}
	return e, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is e at now, and whether it stopped.
func (e Ellipsis) Advance(now time.Time) (Ellipsis, bool) {
	if !e.Active() {
		return e, true
	}
	e.run, e.p = e.run.Advance(now)
	return e, !e.Active()
}
