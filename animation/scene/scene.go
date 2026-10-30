// Package scene draws whole scenes from the time alone: rain, snow, confetti,
// fireworks, stars, static and plasma, on their own or behind a block.
// Non-themable: colours come as arguments, never from the theme.
package scene

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/frames"
)

type sceneKind int

const (
	rain sceneKind = iota
	snow
	confetti
	fireworks
	starfield
	static
	plasma
	fire
	aurora
	bubbles
	sea
	life
)

// Scene plays behind content, or in place of it: rain,
// snow, stars, fireworks. It is drawn from the time alone, at any size, for as
// long as it lives; Still, it holds its first frame.
type Scene struct {
	kind   sceneKind
	colors []color.Color
	t      float64 // seconds since the first frame, at Speed
	run    frames.Run
}

var _ bezel.Model[Scene] = Scene{}

func newScene(kind sceneKind, colors ...color.Color) Scene {
	a := Scene{kind: kind, colors: colors, run: frames.NewRun(time.Second)}
	a.run.Repeat = animation.Forever
	return a
}

// View is the scene w columns wide and h rows tall.
func (a Scene) View(w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	cv := cells.NewCanvas(w, h)
	a.paint(cv, w, h)
	return cv.Grid().String()
}

// Under is the scene with block on top of it: every cell of block that draws
// something, a grapheme or a background, hides the scene there.
func (a Scene) Under(block string) string {
	g := cells.Parse(block)
	w, h := g.Width(), len(g)
	cv := cells.NewCanvas(w, h)
	a.paint(cv, w, h)
	for y, row := range g {
		x := 0
		for _, c := range row {
			if strings.TrimSpace(c.Text) != "" || c.Bg != nil {
				cv.Put(x, y, c)
			}
			x += c.Width
		}
	}
	return cv.Grid().String()
}

// paint draws the scene at a.t onto cv.
func (a Scene) paint(cv cells.Canvas, w, h int) {
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
	case fire:
		a.fire(cv, w, h)
	case aurora:
		a.aurora(cv, w, h)
	case bubbles:
		a.bubbles(cv, w, h)
	case sea:
		a.sea(cv, w, h)
	case life:
		a.life(cv, w, h)
	}
}

// color is the scene's i-th colour, round its list; nil when it has none.
func (a Scene) color(i uint64) color.Color {
	if len(a.colors) == 0 {
		return nil
	}
	return a.colors[i%uint64(len(a.colors))]
}

// Active reports whether the scene is still playing.
func (a Scene) Active() bool { return a.run.Active() }

// Tick is the command for a's next frame, nil once it stopped.
func (a Scene) Tick() tea.Cmd { return a.run.Tick() }

// FrameAt is a's frame for at, for a test or a host that delivers frames itself.
func (a Scene) FrameAt(at time.Time) animation.FrameMsg { return a.run.FrameAt(at) }

// Update plays on its own frames and asks for the next while it runs.
func (a Scene) Update(msg tea.Msg) (Scene, bezel.Result) {
	f, ok := a.run.Own(msg)
	if !ok {
		return a, bezel.Result{Handled: ok}
	}
	a, _ = a.Advance(f.At)
	var cmd tea.Cmd
	a.run, cmd = a.run.Next()
	return a, bezel.Result{Handled: true, Cmd: cmd}
}

// Advance is a at now; a scene never ends on its own, so done is false while
// it plays.
func (a Scene) Advance(now time.Time) (Scene, bool) {
	if !a.Active() {
		return a, true
	}
	a.run, _ = a.run.Advance(now)
	a.t = frames.Paced(now.Sub(a.run.Start)).Seconds()
	return a, false
}
