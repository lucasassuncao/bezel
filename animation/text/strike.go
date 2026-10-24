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

// Strike draws a line through text from left to right and leaves it: a task
// done, an option withdrawn.
type Strike struct {
	text []string // the graphemes
	ink  color.Color
	p    float64
	run  frames.Run
}

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
	n := int(math.Round(frames.Or(s.run.Ease)(s.p) * float64(len(s.text))))
	n = min(max(n, 0), len(s.text))
	st := lipgloss.NewStyle().Strikethrough(true)
	if s.ink != nil {
		st = st.Foreground(s.ink)
	}
	if n == 0 {
		return strings.Join(s.text, "")
	}
	return st.Render(strings.Join(s.text[:n], "")) + strings.Join(s.text[n:], "")
}

// Active reports whether the line is still being drawn.
func (s Strike) Active() bool { return s.run.Active() }

// Tick is the command for s's next frame, nil once it ended.
func (s Strike) Tick() tea.Cmd { return s.run.Tick() }

// FrameAt is s's frame for at, for a test or a host that delivers frames itself.
func (s Strike) FrameAt(at time.Time) animation.FrameMsg { return s.run.FrameAt(at) }

// Update draws on its own frames and asks for the next until it ends.
func (s Strike) Update(msg tea.Msg) (Strike, tea.Cmd) {
	r, p, ok, cmd := s.run.Update(msg)
	if ok {
		s.run, s.p = r, p
	}
	return s, cmd
}

// Advance is s at now, and whether it ended.
func (s Strike) Advance(now time.Time) (Strike, bool) {
	if !s.Active() {
		return s, true
	}
	s.run, s.p = s.run.Advance(now)
	return s, !s.Active()
}
