package animation

import (
	"image/color"
	"math"
	"strings"
	"time"
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

// unit is n as a share in [0,1).
func unit(n uint64) float64 { return float64(n>>11) / (1 << 53) }

// explode is g with each cell flown from its place by how far it has turned:
// away from the middle, a little off line, faster as it goes.
func (e Effect) explode(g grid, w, h int, p float64) grid {
	cv := newCanvas(w, h)
	cx, cy := float64(w-1)/2, float64(h-1)/2
	var k uint64 // the cell's number, for its own line and pace
	for y, row := range g {
		x := 0
		for _, c := range row {
			k++
			q := e.reach(p, x, y, w, h)
			switch {
			case q <= 0:
				cv.put(x, y, c)
			case q < 1 && strings.TrimSpace(c.text) != "":
				dx, dy := float64(x)-cx, 2*(float64(y)-cy)
				angle := math.Atan2(dy, dx) + (unit(noise(5, k))-0.5)*1.2
				if dx == 0 && dy == 0 {
					angle = 2 * math.Pi * unit(noise(5, k))
				}
				dist := q * q * (0.6 + 0.8*unit(noise(6, k))) * float64(max(w, h))
				cv.put(int(math.Round(float64(x)+math.Cos(angle)*dist)), int(math.Round(float64(y)+math.Sin(angle)*dist/2)), c)
			}
			x += c.width
		}
	}
	return cv.grid()
}

// expand is g with only the columns within p of the way out from the middle
// shown, and a half block at either front.
func (e Effect) expand(g grid, w int, p float64) grid {
	cx, open := float64(w)/2, p*float64(w)/2
	lo, hi := int(math.Round(cx-open)), int(math.Round(cx+open))
	out := make(grid, len(g))
	for y, row := range g {
		cv := newCanvas(w, 1)
		x := 0
		for _, c := range row {
			if x >= lo && x+c.width <= hi {
				cv.put(x, 0, c)
			}
			x += c.width
		}
		if p < 1 {
			cv.put(lo-1, 0, cell{text: "▐", width: 1, fg: e.tint})
			cv.put(hi, 0, cell{text: "▌", width: 1, fg: e.tint})
		}
		out[y] = cv.grid()[0]
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
func (e Effect) stretch(g grid, w, h int, p float64) grid {
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
	cv := newCanvas(w, h)
	for y, row := range g {
		x := 0
		for _, c := range row {
			if (horizontal && shown(x, c.width)) || (!horizontal && shown(y, 1)) {
				cv.put(x, y, c)
			}
			x += c.width
		}
	}
	if glyph != "" {
		for i := range max(w, h) {
			if horizontal && i < h {
				cv.put(front, i, cell{text: glyph, width: 1, fg: e.tint})
			} else if !horizontal && i < w {
				cv.put(i, front, cell{text: glyph, width: 1, fg: e.tint})
			}
		}
	}
	return cv.grid()
}

// canvas is a w×h block of columns that cells are put into, whole, then read
// back as a grid with every empty column blank.
type canvas struct {
	w, h  int
	slots [][]slot
}

type slot struct {
	c        cell
	used     bool // a cell starts here
	covering int  // a wide cell starting this many columns to the left covers it
}

func newCanvas(w, h int) canvas {
	cv := canvas{w: w, h: h, slots: make([][]slot, h)}
	for y := range cv.slots {
		cv.slots[y] = make([]slot, w)
	}
	return cv
}

// put places c at x, y, clearing any cell it lands on; one that would not fit
// whole is left out.
func (cv canvas) put(x, y int, c cell) {
	if y < 0 || y >= cv.h || x < 0 || x+c.width > cv.w {
		return
	}
	row := cv.slots[y]
	for i := x; i < x+c.width; i++ {
		if back := row[i].covering; back > 0 {
			cv.clear(row, i-back)
		}
		if row[i].used {
			cv.clear(row, i)
		}
	}
	row[x] = slot{c: c, used: true}
	for i := 1; i < c.width; i++ {
		row[x+i] = slot{covering: i}
	}
}

// clear empties the cell starting at x and the columns it covers.
func (cv canvas) clear(row []slot, x int) {
	width := max(row[x].c.width, 1)
	for i := x; i < x+width && i < len(row); i++ {
		row[i] = slot{}
	}
}

// grid reads the canvas back, a blank for every column no cell holds.
func (cv canvas) grid() grid {
	g := make(grid, cv.h)
	for y, row := range cv.slots {
		cells := []cell{}
		for x := 0; x < cv.w; {
			if row[x].used {
				cells = append(cells, row[x].c)
				x += row[x].c.width
				continue
			}
			cells = append(cells, cell{text: " ", width: 1})
			x++
		}
		g[y] = cells
	}
	return g
}
