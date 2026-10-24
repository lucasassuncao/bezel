package effect

import (
	"image/color"
	"time"

	"github.com/lucasassuncao/bezel/internal/cells"
)

// stackDim is the colour, and stackDepth the share of it, the old block sinks
// into as the new one is laid over it.
var stackDim color.Color = color.RGBA{0, 0, 0, 255}

const stackDepth = 0.55

// Iris closes the old block down to a point in its middle and opens the new
// one out of it: the end of one scene and the start of the next.
func Iris(dur time.Duration) Transition { return newTransition(irisSwap, dur) }

// Blinds turns the screen over in n slats at once, each row of a slat
// flipping from the old block to the new, top first, like a venetian blind.
func Blinds(n int, dur time.Duration) Transition {
	t := newTransition(blinds, dur)
	t.slats = max(n, 1)
	return t
}

// Stack lays the new block over the old one travelling d, the old one
// sinking into the dark beneath it: a screen opened on top of another.
func Stack(d Direction, dur time.Duration) Transition {
	t := newTransition(stack, dur)
	t.dir = d
	return t
}

// irised puts a inside a circle closing over the first half, b inside one
// opening over the second.
func irised(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
	g, open := a, 1-2*p
	if p >= 0.5 {
		g, open = b, 2*p-1
	}
	radial := Radial()
	for y, row := range g {
		x := 0
		for _, c := range row {
			if radial(x, y, w, h) < open {
				cv.Put(x, y, c)
			}
			x += c.Width
		}
	}
}

// blinded puts each row of b whose turn in its slat has come, and a's elsewhere.
func (t Transition) blinded(cv cells.Canvas, a, b cells.Grid, h int, p float64) {
	slat := max((h+t.slats-1)/t.slats, 1)
	for y := range a {
		g := a
		if p > (float64(y%slat)+0.5)/float64(slat) {
			g = b
		}
		draw(cv, cells.Grid{g[y]}, 0, y)
	}
}

// stacked puts a sunk p of the way into the dark, and b sliding in over it.
func (t Transition) stacked(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
	sunk := make(cells.Grid, len(a))
	for y, row := range a {
		sunk[y] = make([]cells.Cell, len(row))
		for i, c := range row {
			sunk[y][i] = tinted(c, stackDim, 1-stackDepth*p)
		}
	}
	draw(cv, sunk, 0, 0)
	vx, vy := vector(t.dir, w, h)
	draw(cv, b, round(-vx*(1-p)), round(-vy*(1-p)))
}
