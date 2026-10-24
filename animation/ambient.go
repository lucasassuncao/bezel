package animation

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

type ambientKind int

const (
	rain ambientKind = iota
	snow
	confetti
	fireworks
	starfield
	static
	plasma
)

// Ambient is a scene that plays behind content, or in place of it: rain,
// snow, stars, fireworks. It is drawn from the time alone, at any size, for as
// long as it lives; Still, it holds its first frame.
type Ambient struct {
	kind   ambientKind
	colors []color.Color
	t      float64 // seconds since the first frame, at Speed
	run    run
}

func newAmbient(kind ambientKind, colors ...color.Color) Ambient {
	a := Ambient{kind: kind, colors: colors, run: newRun(time.Second)}
	a.run.ease, a.run.repeat = Linear, Forever
	return a
}

// View is the scene w columns wide and h rows tall.
func (a Ambient) View(w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	cv := newCanvas(w, h)
	a.paint(cv, w, h)
	return cv.grid().String()
}

// Under is the scene with block on top of it: every cell of block that draws
// something, a grapheme or a background, hides the scene there.
func (a Ambient) Under(block string) string {
	g := parse(block)
	w, h := g.width(), len(g)
	cv := newCanvas(w, h)
	a.paint(cv, w, h)
	for y, row := range g {
		x := 0
		for _, c := range row {
			if strings.TrimSpace(c.text) != "" || c.bg != nil {
				cv.put(x, y, c)
			}
			x += c.width
		}
	}
	return cv.grid().String()
}

// paint draws the scene at a.t onto cv.
func (a Ambient) paint(cv canvas, w, h int) {
	switch a.kind {
	case rain:
		a.rain(cv, w, h)
	case snow:
		a.flakes(cv, w, h, false)
	case confetti:
		a.flakes(cv, w, h, true)
	case fireworks:
		a.fireworks(cv, w, h)
	case starfield:
		a.stars(cv, w, h)
	case static:
		a.static(cv, w, h)
	case plasma:
		a.plasma(cv, w, h)
	}
}

// color is the scene's i-th colour, round its list; nil when it has none.
func (a Ambient) color(i uint64) color.Color {
	if len(a.colors) == 0 {
		return nil
	}
	return a.colors[i%uint64(len(a.colors))]
}

// Active reports whether the scene is still playing.
func (a Ambient) Active() bool { return a.run.active() }

// Tick is the command for a's next frame, nil once it stopped.
func (a Ambient) Tick() tea.Cmd { return a.run.tick() }

// FrameAt is a's frame for at, for a test or a host that delivers frames itself.
func (a Ambient) FrameAt(at time.Time) FrameMsg { return a.run.frameAt(at) }

// Update plays on its own frames and asks for the next while it runs.
func (a Ambient) Update(msg tea.Msg) (Ambient, tea.Cmd) {
	f, ok := a.run.own(msg)
	if !ok {
		return a, nil
	}
	a, _ = a.Advance(f.At)
	var cmd tea.Cmd
	a.run, cmd = a.run.next()
	return a, cmd
}

// Advance is a at now; a scene never ends on its own, so done is false while
// it plays.
func (a Ambient) Advance(now time.Time) (Ambient, bool) {
	if !a.Active() {
		return a, true
	}
	a.run, _ = a.run.advance(now)
	a.t = paced(now.Sub(a.run.start)).Seconds()
	return a, false
}
