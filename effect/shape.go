package effect

import (
	"math"
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/internal/cells"
)

// pixelMax is the largest block, in cells, Pixelate coarsens the block to.
const pixelMax = 4

// IrisIn opens the block from a point in its middle, a widening circle
// showing it: a scene that begins. Reverse closes it to the point.
func IrisIn(dur time.Duration) Effect { return newEffect(iris, nil, dur) }

// IrisOut closes the block down to a point in its middle.
func IrisOut(dur time.Duration) Effect { return IrisIn(dur).Reverse() }

// Pixelate coarsens the block into squares of colour, coarsest halfway, and
// sharpens it again: a screen being changed, a thing being redacted.
func Pixelate(dur time.Duration) Effect { return newEffect(pixelate, nil, dur) }

// iris is g with only the cells inside a circle p of the way out shown.
func (e Effect) iris(g cells.Grid, w, h int, p float64) cells.Grid {
	radial := Radial()
	for y, row := range g {
		x := 0
		for i, c := range row {
			if p < 1 && radial(x, y, w, h) >= p {
				row[i] = blank(c)
			}
			x += c.Width
		}
	}
	return g
}

// pixelate is g drawn in squares of a side that grows to pixelMax halfway:
// each square takes the colour of the cell at its top left.
func (e Effect) pixelate(g cells.Grid, w, h int, p float64) cells.Grid {
	side := 1 + int(math.Round(float64(pixelMax-1)*math.Sin(math.Pi*p)))
	if side <= 1 {
		return g
	}
	cv := cells.NewCanvas(w, h)
	for y, row := range g {
		corner := columns(g[y/side*side])
		x := 0
		for _, c := range row {
			src, ok := corner.at(x / side * side)
			if !ok {
				src = c
			}
			col := src.Fg
			if strings.TrimSpace(src.Text) == "" || src.Bg != nil {
				col = src.Bg
			}
			px := cells.Cell{Text: strings.Repeat(" ", c.Width), Width: c.Width}
			if col != nil {
				px = cells.Cell{Text: strings.Repeat("█", c.Width), Width: c.Width, Fg: col}
			}
			cv.Put(x, y, px)
			x += c.Width
		}
	}
	return cv.Grid()
}
