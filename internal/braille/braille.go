// Package braille is a grid of braille cells, each two dots wide and four
// tall, coloured by whatever lit it last; chart.Lines and canvas draw on it.
package braille

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// dotBits are the braille bits of a cell's dots, by column then by row from
// the top.
var dotBits = [2][4]rune{{0x01, 0x02, 0x04, 0x40}, {0x08, 0x10, 0x20, 0x80}}

// wide marks the cell a wide rune of text spills into.
const wide = -1

// Grid is w by h cells. Its slices are shared, so a copy draws on the same
// cells.
type Grid struct {
	w, h  int
	bits  []rune
	owner []int
	text  []rune // a rune of text over the dots, 0 for none, wide for spill
}

// New is an empty grid of w by h cells.
func New(w, h int) Grid {
	w, h = max(w, 0), max(h, 0)
	return Grid{w: w, h: h, bits: make([]rune, w*h), owner: make([]int, w*h), text: make([]rune, w*h)}
}

// Dots is the grid's size in dots.
func (g Grid) Dots() (w, h int) { return g.w * 2, g.h * 4 }

// Set lights dot x, y (from the top left) for owner; one outside is dropped.
func (g Grid) Set(x, y, owner int) {
	if x < 0 || y < 0 || x >= g.w*2 || y >= g.h*4 {
		return
	}
	c := (y/4)*g.w + x/2
	g.bits[c] |= dotBits[x%2][y%4]
	g.owner[c] = owner
}

// Line lights the dots from x0, y0 to x1, y1, both ends included.
func (g Grid) Line(x0, y0, x1, y1, owner int) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	e := dx + dy
	for {
		g.Set(x0, y0, owner)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * e
		if e2 >= dy {
			e += dy
			x0 += sx
		}
		if e2 <= dx {
			e += dx
			y0 += sy
		}
	}
}

// Text writes s over the dots from cell col, row; what falls outside is cut.
func (g Grid) Text(col, row int, s string, owner int) {
	if row < 0 || row >= g.h {
		return
	}
	for _, r := range s {
		w := ansi.StringWidth(string(r))
		if w == 0 {
			continue
		}
		if col >= 0 && col+w <= g.w {
			c := row*g.w + col
			g.text[c], g.owner[c] = r, owner
			if w == 2 {
				g.text[c+1] = wide
			}
		}
		col += w
	}
}

// Row is cell row r, each run of one owner's cells in styles[owner], wrapping
// round the slice; with no styles it is plain.
func (g Grid) Row(r int, styles []lipgloss.Style) string {
	if r < 0 || r >= g.h {
		return ""
	}
	var b strings.Builder
	lo := r * g.w
	for i := lo; i < lo+g.w; {
		if g.empty(i) {
			b.WriteByte(' ')
			i++
			continue
		}
		var run strings.Builder
		j := i
		for ; j < lo+g.w && !g.empty(j) && g.owner[j] == g.owner[i]; j++ {
			switch t := g.text[j]; {
			case t == wide:
			case t != 0:
				run.WriteRune(t)
			default:
				run.WriteRune(0x2800 + g.bits[j])
			}
		}
		b.WriteString(paint(run.String(), g.owner[i], styles))
		i = j
	}
	return b.String()
}

func (g Grid) empty(c int) bool { return g.bits[c] == 0 && g.text[c] == 0 }

func paint(s string, owner int, styles []lipgloss.Style) string {
	if len(styles) == 0 {
		return s
	}
	return styles[((owner%len(styles))+len(styles))%len(styles)].Render(s)
}

func abs(n int) int { return max(n, -n) }

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}
