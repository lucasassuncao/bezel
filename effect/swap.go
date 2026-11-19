package effect

import (
	"image/color"
	"time"

	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/noise"
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

// Doors parts the old block down its middle, the halves sliding out to either
// side and uncovering the new one beneath: a lift's doors opening.
func Doors(dur time.Duration) Transition { return newTransition(doors, dur) }

// Shut slides the new block in from both sides over the old one, its halves
// meeting in the middle: Doors the other way round.
func Shut(dur time.Duration) Transition { return newTransition(shut, dur) }

// Comb slides the new block in over the old one row by row, every other row
// from the other side, the two sets meshing like the teeth of a comb.
func Comb(dur time.Duration) Transition { return newTransition(comb, dur) }

// Static turns the old block to the snow of a lost signal and settles the
// snow into the new one: a channel changed.
func Static(dur time.Duration) Transition { return newTransition(static, dur) }

// halves puts g on cv split down its middle, each half moved off columns
// outward; what leaves the box is gone.
func halves(cv cells.Canvas, g cells.Grid, w, off int) {
	mid := w / 2
	for y, row := range g {
		x := 0
		for _, c := range row {
			if x < mid {
				cv.Put(x-off, y, c)
			} else {
				cv.Put(x+off, y, c)
			}
			x += c.Width
		}
	}
}

// combed puts a, and over it each row of b slid in p of the way, the even
// rows from the right and the odd ones from the left.
func combed(cv cells.Canvas, a, b cells.Grid, w int, p float64) {
	draw(cv, a, 0, 0)
	off := round(float64(w) * (1 - p))
	for y, row := range b {
		dx := off
		if y%2 == 1 {
			dx = -off
		}
		draw(cv, cells.Grid{row}, dx, y)
	}
}

// snowGreys are the three greys a static's snow is drawn in, darkest first.
var snowGreys = []color.Color{color.RGBA{100, 100, 100, 255}, color.RGBA{150, 150, 150, 255}, color.RGBA{205, 205, 205, 255}}

// snow is what a static's snow is drawn with, its shades spread over the
// three greys so that neighbours often share one.
var snow = []cells.Cell{
	{Text: "░", Width: 1, Fg: snowGreys[0]}, {Text: "▒", Width: 1, Fg: snowGreys[0]}, {Text: "·", Width: 1, Fg: snowGreys[0]},
	{Text: "▓", Width: 1, Fg: snowGreys[1]}, {Text: "▀", Width: 1, Fg: snowGreys[1]}, {Text: "▄", Width: 1, Fg: snowGreys[1]},
	{Text: "█", Width: 1, Fg: snowGreys[2]}, {Text: " ", Width: 1, Fg: snowGreys[2]},
}

// snowSlots is how many times over its run a static's snow changes.
const snowSlots = 40

// snowed puts snow where a cell's moment has come, rising over a in the first
// half and falling away from b in the second; each cell keeps one moment.
func snowed(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
	g, snowing := a, func(n float64) bool { return n < 2*p }
	if p >= 0.5 {
		g, snowing = b, func(n float64) bool { return n >= 2*p-1 }
	}
	slot := uint64(p * snowSlots)
	for y := range h {
		src := columns(g[y])
		for x := range w {
			k := uint64(y*w + x) // #nosec G115 -- a cell's number, never negative
			if snowing(noise.Unit(noise.Hash(51, k))) {
				cv.Put(x, y, snow[noise.Hash(52, k*snowSlots+slot)%uint64(len(snow))])
				continue
			}
			if c, ok := src.at(x); ok {
				cv.Put(x, y, c)
			}
		}
	}
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
