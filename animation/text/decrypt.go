package text

import (
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// flaps is the order a split-flap letter turns through, blank first.
const flaps = " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// flip is the share of the run one flap takes to turn.
const flip = 0.04

// Decrypt turns each letter through the flaps of a departures board until it
// shows its own, the leftmost landing first: a status that arrives, a code
// revealed. Unlike Scramble the letters roll in order, not at random.
type Decrypt struct {
	text []string // the graphemes
	p    float64
	run  frames.Run
}

var _ bezel.Model[Decrypt] = Decrypt{}

// NewDecrypt lands every letter of text within dur.
func NewDecrypt(text string, dur time.Duration) Decrypt {
	d := Decrypt{text: edit.Graphemes(text), run: frames.NewRun(dur)}
	if !d.Active() {
		d.p = 1
	}
	return d
}

// WithDelay holds every letter blank for d before the first starts to turn.
func (d Decrypt) WithDelay(delay time.Duration) Decrypt {
	d.run.Delay = max(delay, 0)
	return d
}

// View is the text with every letter still turning on the flap it has reached.
func (d Decrypt) View() string {
	var b strings.Builder
	for i, g := range d.text {
		b.WriteString(d.flapAt(i, g))
	}
	return b.String()
}

// flapAt is what grapheme i shows now: itself once landed, else the flap a
// few turns before it. Graphemes off the board show as they are.
func (d Decrypt) flapAt(i int, g string) string {
	r, size := utf8.DecodeRuneInString(g)
	target := strings.IndexRune(flaps, unicode.ToUpper(r))
	if size != len(g) || target < 0 {
		return g
	}
	lands := 0.3 + 0.7*float64(i)/float64(max(len(d.text)-1, 1))
	if d.p >= lands {
		return g
	}
	turns := int(math.Ceil((lands - d.p) / flip))
	shown := rune(flaps[((target-turns)%len(flaps)+len(flaps))%len(flaps)])
	if unicode.IsLower(r) {
		shown = unicode.ToLower(shown)
	}
	return string(shown)
}

// Active reports whether any letter is still turning.
func (d Decrypt) Active() bool { return d.run.Active() }

// Tick is the command for d's next frame, nil once every letter landed.
func (d Decrypt) Tick() tea.Cmd { return d.run.Tick() }

// FrameAt is d's frame for at, for a test or a host that delivers frames itself.
func (d Decrypt) FrameAt(at time.Time) animation.FrameMsg { return d.run.FrameAt(at) }

// Update turns on its own frames and asks for the next until every letter lands.
func (d Decrypt) Update(msg tea.Msg) (Decrypt, bezel.Result) {
	r, p, ok, cmd := d.run.Update(msg)
	if ok {
		d.run, d.p = r, p
	}
	return d, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is d at now, and whether every letter landed.
func (d Decrypt) Advance(now time.Time) (Decrypt, bool) {
	if !d.Active() {
		return d, true
	}
	d.run, d.p = d.run.Advance(now)
	return d, !d.Active()
}
