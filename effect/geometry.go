package effect

import (
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/noise"
)

// Explode throws the block's cells outward from its middle, each on its own
// line and pace, until none is left.
func Explode(dur time.Duration) Effect { return newEffect(explode, nil, dur) }

// Expand opens the block from its middle column outward, its two fronts drawn
// as half blocks in edge, like a curtain parting.
func Expand(edge color.Color, dur time.Duration) Effect {
	e := newEffect(expand, nil, dur)
	e.tint = edge
	return e
}

// Stretch fills the block in from an edge travelling d, its front cell drawn
// in eighths of a block in edge, so the fill moves smoother than a cell.
func Stretch(d Direction, edge color.Color, dur time.Duration) Effect {
	e := newEffect(stretch, nil, dur)
	e.dir, e.tint = d, edge
	return e
}

// explode is g with each cell flown from its place by how far it has turned:
// away from the middle, a little off line, faster as it goes.
func (e Effect) explode(g cells.Grid, w, h int, p float64) cells.Grid {
	cv := cells.NewCanvas(w, h)
	cx, cy := float64(w-1)/2, float64(h-1)/2
	var k uint64 // the cell's number, for its own line and pace
	for y, row := range g {
		x := 0
		for _, c := range row {
			k++
			q := e.reach(p, x, y, w, h)
			switch {
			case q <= 0:
				cv.Put(x, y, c)
			case q < 1 && strings.TrimSpace(c.Text) != "":
				dx, dy := float64(x)-cx, 2*(float64(y)-cy)
				angle := math.Atan2(dy, dx) + (noise.Unit(noise.Hash(5, k))-0.5)*1.2
				if dx == 0 && dy == 0 {
					angle = 2 * math.Pi * noise.Unit(noise.Hash(5, k))
				}
				dist := q * q * (0.6 + 0.8*noise.Unit(noise.Hash(6, k))) * float64(max(w, h))
				cv.Put(int(math.Round(float64(x)+math.Cos(angle)*dist)), int(math.Round(float64(y)+math.Sin(angle)*dist/2)), c)
			}
			x += c.Width
		}
	}
	return cv.Grid()
}

// expand is g with only the columns within p of the way out from the middle
// shown, and a half block at either front.
func (e Effect) expand(g cells.Grid, w int, p float64) cells.Grid {
	cx, open := float64(w)/2, p*float64(w)/2
	lo, hi := int(math.Round(cx-open)), int(math.Round(cx+open))
	out := make(cells.Grid, len(g))
	for y, row := range g {
		cv := cells.NewCanvas(w, 1)
		x := 0
		for _, c := range row {
			if x >= lo && x+c.Width <= hi {
				cv.Put(x, 0, c)
			}
			x += c.Width
		}
		if p < 1 {
			cv.Put(lo-1, 0, cells.Cell{Text: "▐", Width: 1, Fg: e.tint})
			cv.Put(hi, 0, cells.Cell{Text: "▌", Width: 1, Fg: e.tint})
		}
		out[y] = cv.Grid()[0]
	}
	return out
}

// Eighths of a block, a front edge each way it can travel: from the left,
// from the right, from the top and from the bottom.
var (
	fromLeft   = []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉"}
	fromRight  = []string{"▕", "▕", "▕", "▐", "▐", "▐", "▐"}
	fromTop    = []string{"▔", "▔", "▔", "▀", "▀", "▀", "▀"}
	fromBottom = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇"}
)

// stretch is g filled p of the way from the edge it starts at, the cells past
// the front blank and the front cell an eighth block.
func (e Effect) stretch(g cells.Grid, w, h int, p float64) cells.Grid {
	horizontal := e.dir == Right || e.dir == Left
	size := h
	if horizontal {
		size = w
	}
	f := p * float64(size)
	full := int(f)
	n := int((f - float64(full)) * 8)  // eighths of the front cell filled
	shown := func(at, span int) bool { // the cell from at, span long, is filled
		if e.dir == Right || e.dir == Down {
			return at+span <= full
		}
		return at >= size-full
	}
	front := full // where the front cell sits
	if e.dir == Left || e.dir == Up {
		front = size - full - 1
	}
	glyph := ""
	if n > 0 && p < 1 {
		glyph = map[Direction][]string{Right: fromLeft, Left: fromRight, Down: fromTop, Up: fromBottom}[e.dir][n-1]
	}
	cv := cells.NewCanvas(w, h)
	for y, row := range g {
		x := 0
		for _, c := range row {
			if (horizontal && shown(x, c.Width)) || (!horizontal && shown(y, 1)) {
				cv.Put(x, y, c)
			}
			x += c.Width
		}
	}
	if glyph != "" {
		for i := range max(w, h) {
			if horizontal && i < h {
				cv.Put(front, i, cells.Cell{Text: glyph, Width: 1, Fg: e.tint})
			} else if !horizontal && i < w {
				cv.Put(i, front, cells.Cell{Text: glyph, Width: 1, Fg: e.tint})
			}
		}
	}
	return cv.Grid()
}
