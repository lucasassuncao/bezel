package text

import (
	"image/color"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Strike draws a line through text from left to right and leaves it: a task
// done, an option withdrawn.
type Strike struct {
	text []string // the graphemes
	ink  color.Color
	p    float64
	run  frames.Run
}

var _ bezel.Model[Strike] = Strike{}

// NewStrike strikes text through over dur. Still, it is born struck.
func NewStrike(text string, dur time.Duration) Strike {
	s := Strike{text: edit.Graphemes(text), run: frames.NewRun(dur)}
	if !s.Active() {
		s.p = 1
	}
	return s
}

// WithInk draws the struck graphemes in c, a muted grey for something done.
func (s Strike) WithInk(c color.Color) Strike {
	s.ink = c
	return s
}

// View is the text with the struck graphemes crossed out.
func (s Strike) View() string {
	st := lipgloss.NewStyle().Strikethrough(true)
	if s.ink != nil {
		st = st.Foreground(s.ink)
	}
	return swept(s.text, frames.Or(s.run.Ease)(s.p), st)
}

// swept is text with its first q of graphemes in st, the rest as they are:
// the line a strike or a marker has reached.
func swept(text []string, q float64, st lipgloss.Style) string {
	n := min(max(int(math.Round(q*float64(len(text)))), 0), len(text))
	if n == 0 {
		return strings.Join(text, "")
	}
	return st.Render(strings.Join(text[:n], "")) + strings.Join(text[n:], "")
}

// Active reports whether the line is still being drawn.
func (s Strike) Active() bool { return s.run.Active() }

// Tick is the command for s's next frame, nil once it ended.
func (s Strike) Tick() tea.Cmd { return s.run.Tick() }

// FrameAt is s's frame for at, for a test or a host that delivers frames itself.
func (s Strike) FrameAt(at time.Time) animation.FrameMsg { return s.run.FrameAt(at) }

// Update draws on its own frames and asks for the next until it ends.
func (s Strike) Update(msg tea.Msg) (Strike, bezel.Result) {
	r, p, ok, cmd := s.run.Update(msg)
	if ok {
		s.run, s.p = r, p
	}
	return s, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is s at now, and whether it ended.
func (s Strike) Advance(now time.Time) (Strike, bool) {
	if !s.Active() {
		return s, true
	}
	s.run, s.p = s.run.Advance(now)
	return s, !s.Active()
}
