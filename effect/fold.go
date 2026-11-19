package effect

import (
	"math"
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/internal/cells"
)

// turnShade is how far a block that turns away is darkened once edge on, and
// peelShade how far the back of a peeling page is.
const (
	turnShade = 0.6
	peelShade = 0.5
)

// Turn turns the screen over like a card: the old block narrows to its middle
// column, darkening as it goes edge on, and the new one widens out of it.
func Turn(dur time.Duration) Transition { return newTransition(turnSwap, dur) }

// Cube turns the screen like the faces of a cube, the new block coming round
// travelling d as the old one goes round the far side, each darker the more
// it faces away.
func Cube(d Direction, dur time.Duration) Transition {
	t := newTransition(cube, dur)
	t.dir = d
	return t
}

// Peel folds the old block back from its bottom right corner, the fold
// travelling to the top left; the new block lies beneath, and the back of the
// page shows shaded.
func Peel(dur time.Duration) Transition { return newTransition(peel, dur) }

// turned puts a narrowed about its middle column over the first half, then b
// widened out of it, each scaled across only and darker the narrower it is.
func turned(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
	g, s := a, 1-2*p
	if p >= 0.5 {
		g, s = b, 2*p-1
	}
	if s <= 0 {
		return
	}
	cx := float64(w-1) / 2
	for y := range h {
		src := columns(g[y])
		for x := range w {
			if math.Abs(float64(x)-cx) > s*float64(w)/2 {
				continue
			}
			if c, ok := src.at(round(cx + (float64(x)-cx)/s)); ok {
				cv.Put(x, y, tinted(c, stackDim, 1-turnShade*(1-s)))
			}
		}
	}
}

// cubed puts a and b side by side, squeezed into the shares of the box p
// gives them, the new one on the side it comes from.
func (t Transition) cubed(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
	old, fresh := turnShade*p, turnShade*(1-p)
	switch t.dir {
	case Right:
		bx := round(float64(w) * p)
		squeezed(cv, b, rect{w: bx, h: h}, w, h, fresh)
		squeezed(cv, a, rect{x: bx, w: w - bx, h: h}, w, h, old)
	case Down:
		by := round(float64(h) * p)
		squeezed(cv, b, rect{w: w, h: by}, w, h, fresh)
		squeezed(cv, a, rect{y: by, w: w, h: h - by}, w, h, old)
	case Up:
		by := round(float64(h) * (1 - p))
		squeezed(cv, a, rect{w: w, h: by}, w, h, old)
		squeezed(cv, b, rect{y: by, w: w, h: h - by}, w, h, fresh)
	default:
		bx := round(float64(w) * (1 - p))
		squeezed(cv, a, rect{w: bx, h: h}, w, h, old)
		squeezed(cv, b, rect{x: bx, w: w - bx, h: h}, w, h, fresh)
	}
}

// squeezed puts the w×h g on cv scaled into r, darkened shade of the way.
func squeezed(cv cells.Canvas, g cells.Grid, r rect, w, h int, shade float64) {
	if r.w <= 0 || r.h <= 0 {
		return
	}
	for y := r.y; y < r.y+r.h; y++ {
		src := columns(g[(y-r.y)*h/r.h])
		for x := r.x; x < r.x+r.w; x++ {
			if c, ok := src.at((x - r.x) * w / r.w); ok {
				cv.Put(x, y, tinted(c, stackDim, 1-shade))
			}
		}
	}
}

// peeled puts b where the fold has passed and a before it, but where a's
// peeled corner lies folded over a shows the back of the page instead: the
// cell mirrored across the fold, its text gone to a shade.
func peeled(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
	far := float64(w-1) + 2*float64(h-1)
	fold := far - p*(far+1) // on the line x + 2y = fold, rows counting double
	olds := make([]byColumn, h)
	for y := range h {
		olds[y] = columns(a[y])
	}
	for y := range h {
		news := columns(b[y])
		for x := range w {
			if float64(x+2*y) > fold {
				if c, ok := news.at(x); ok {
					cv.Put(x, y, c)
				}
				continue
			}
			mx, my := round(fold-2*float64(y)), round((fold-float64(x))/2)
			if mx >= 0 && mx < w && my >= 0 && my < h {
				if c, ok := olds[my].at(mx); ok {
					cv.Put(x, y, cells.Cell{Text: strings.Repeat("░", c.Width), Width: c.Width, Fg: tinted(c, stackDim, 1-peelShade).Fg})
				}
				continue
			}
			if c, ok := olds[y].at(x); ok {
				cv.Put(x, y, c)
			}
		}
	}
}
