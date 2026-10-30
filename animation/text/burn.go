package text

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/internal/noise"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// burnWidth is the share of the run one letter takes from catching to gone.
const burnWidth = 0.3

// The fire a letter passes through, hottest first, and the ash it leaves.
var (
	flame = []color.Color{
		color.RGBA{255, 230, 120, 255},
		color.RGBA{255, 150, 30, 255},
		color.RGBA{215, 40, 20, 255},
	}
	ash = color.RGBA{110, 110, 110, 255}
)

// Burn sets text alight from the left, unevenly, each letter flaring through
// the colours of a flame and falling to ash before it is gone: something
// deleted for good, a message that self-destructs. Still, it is born gone.
type Burn struct {
	text []string
	p    float64
	run  frames.Run
}

var _ bezel.Model[Burn] = Burn{}

// NewBurn burns text away over dur.
func NewBurn(text string, dur time.Duration) Burn {
	b := Burn{text: edit.Graphemes(text), run: frames.NewRun(dur)}
	if !b.Active() {
		b.p = 1
	}
	return b
}

// catches is when grapheme i catches fire, in [0, 1-burnWidth].
func (b Burn) catches(i int) float64 {
	front := float64(i) / float64(max(len(b.text)-1, 1))
	jitter := (noise.Unit(noise.Hash(5, uint64(i))) - 0.5) * 0.15 // #nosec G115 -- a hash key
	return min(max(front*0.8+jitter, 0), 1) * (1 - burnWidth)
}

// View is the text burning, its width kept: blanks where letters are gone.
func (b Burn) View() string {
	var s strings.Builder
	for i, g := range b.text {
		q := (b.p - b.catches(i)) / burnWidth
		switch {
		case q <= 0:
			s.WriteString(g)
		case q < 0.6:
			c := flame[min(int(q/0.6*float64(len(flame))), len(flame)-1)]
			s.WriteString(lipgloss.NewStyle().Foreground(c).Bold(true).Render(g))
		case q < 0.9:
			s.WriteString(lipgloss.NewStyle().Foreground(ash).Render("·" + strings.Repeat(" ", max(ansi.StringWidth(g)-1, 0))))
		default:
			s.WriteString(strings.Repeat(" ", ansi.StringWidth(g)))
		}
	}
	return s.String()
}

// Active reports whether anything is still burning.
func (b Burn) Active() bool { return b.run.Active() }

// Tick is the command for b's next frame, nil once it is gone.
func (b Burn) Tick() tea.Cmd { return b.run.Tick() }

// FrameAt is b's frame for at, for a test or a host that delivers frames itself.
func (b Burn) FrameAt(at time.Time) animation.FrameMsg { return b.run.FrameAt(at) }

// Update burns on its own frames and asks for the next until it is gone.
func (b Burn) Update(msg tea.Msg) (Burn, bezel.Result) {
	r, p, ok, cmd := b.run.Update(msg)
	if ok {
		b.run, b.p = r, p
	}
	return b, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is b at now, and whether it is gone.
func (b Burn) Advance(now time.Time) (Burn, bool) {
	if !b.Active() {
		return b, true
	}
	b.run, b.p = b.run.Advance(now)
	return b, !b.Active()
}

// WithDelay holds the text whole for d after the first frame.
func (b Burn) WithDelay(d time.Duration) Burn {
	b.run.Delay = max(d, 0)
	return b
}
