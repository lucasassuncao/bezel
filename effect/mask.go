package effect

import (
	"github.com/lucasassuncao/bezel/internal/cells"
)

// rect is a part of a block, in cells; a w or h of 0 or less runs to the
// block's edge.
type rect struct{ x, y, w, h int }

// holds reports whether the cell at x, y lies inside r.
func (r rect) holds(x, y int) bool {
	return x >= r.x && y >= r.y && (r.w <= 0 || x < r.x+r.w) && (r.h <= 0 || y < r.y+r.h)
}

// covers reports whether a cell from column x, wide columns, lies inside r's columns.
func (r rect) covers(x, wide int) bool { return x >= r.x && (r.w <= 0 || x+wide <= r.x+r.w) }

// crop is the part of g inside r: whole cells only, its rows as long as theirs.
func crop(g cells.Grid, r rect) cells.Grid {
	var out cells.Grid
	for y := max(r.y, 0); y < len(g) && (r.h <= 0 || y < r.y+r.h); y++ {
		row := []cells.Cell{}
		x := 0
		for _, c := range g[y] {
			if r.covers(x, c.Width) {
				row = append(row, c)
			}
			x += c.Width
		}
		out = append(out, row)
	}
	return out
}

// paste is g with sub laid back where crop took it from, the cells around
// it kept.
func paste(g, sub cells.Grid, r rect) cells.Grid {
	left := max(r.x, 0)
	for i, srow := range sub {
		y := max(r.y, 0) + i
		if y >= len(g) {
			break
		}
		cv := cells.NewCanvas(max(rowWidth(g[y]), left+rowWidth(srow)), 1)
		x := 0
		for _, c := range g[y] {
			if !r.covers(x, c.Width) {
				cv.Put(x, 0, c)
			}
			x += c.Width
		}
		x = left
		for _, c := range srow {
			cv.Put(x, 0, c)
			x += c.Width
		}
		g[y] = cv.Grid()[0]
	}
	return g
}

// rowWidth is how many columns row's cells take.
func rowWidth(row []cells.Cell) int {
	n := 0
	for _, c := range row {
		n += c.Width
	}
	return n
}
