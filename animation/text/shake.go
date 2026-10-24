package text

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// swings is how many times a Shake goes left and back before it settles.
const swings = 3

// Shake sways a block side to side and dies away: a refusal without words,
// for a wrong password or a value the form will not take.
type Shake struct {
	amp int
	p   float64 // raw progress, 1 at rest
	run frames.Run
}

// NewShake swings up to amp cells either way over dur.
func NewShake(amp int, dur time.Duration) Shake {
	return Shake{amp: max(amp, 0), p: 1, run: frames.NewRun(dur)}
}

// Offset is how many cells right of centre the block sits now, negative for left.
func (s Shake) Offset() int {
	if !s.run.Active() {
		return 0
	}
	return int(math.Round(float64(s.amp) * math.Sin(2*math.Pi*swings*s.p) * (1 - s.p)))
}

// View is block moved by Offset, every line padded so it is always 2·amp
// cells wider than it was: what sits around it never moves.
func (s Shake) View(block string) string {
	off := s.Offset()
	left, right := strings.Repeat(" ", s.amp+off), strings.Repeat(" ", s.amp-off)
	lines := strings.Split(block, "\n")
	for i, l := range lines {
		lines[i] = left + l + right
	}
	return strings.Join(lines, "\n")
}

// Active reports whether s is still swinging.
func (s Shake) Active() bool { return s.run.Active() }

// Tick is the command for s's next frame, nil once it settled.
func (s Shake) Tick() tea.Cmd { return s.run.Tick() }

// FrameAt is s's frame for at, for a test or a host that delivers frames itself.
func (s Shake) FrameAt(at time.Time) animation.FrameMsg { return s.run.FrameAt(at) }

// Update samples s on its own frames and asks for the next until it settles.
func (s Shake) Update(msg tea.Msg) (Shake, tea.Cmd) {
	f, ok := s.run.Own(msg)
	if !ok {
		return s, nil
	}
	if s, _ = s.Advance(f.At); !s.Active() {
		return s, nil
	}
	var cmd tea.Cmd
	s.run, cmd = s.run.Next()
	return s, cmd
}

// Advance is s sampled at now, and whether it settled.
func (s Shake) Advance(now time.Time) (Shake, bool) {
	if !s.Active() {
		return s, true
	}
	s.run, s.p = s.run.Advance(now)
	return s, !s.run.Active()
}

// WithDelay holds s still for d after its first frame.
func (s Shake) WithDelay(d time.Duration) Shake {
	s.run.Delay = max(d, 0)
	return s
}

// WithRepeat shakes s n more times after the first, or without end for Forever.
func (s Shake) WithRepeat(n int) Shake {
	s.run.Repeat = max(n, animation.Forever)
	return s
}
