// Package canvas is free drawing in braille: lines, boxes, circles, points
// and labels placed in the app's own units, each in a series colour.
package canvas

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/internal/braille"
	"github.com/lucasassuncao/bezel/theme"
)

// Styles are the colours a shape's Series picks from, in turn.
type Styles struct {
	Series []lipgloss.Style
}

// DefaultStyles are the theme's six series colours.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Series: th.Series[:]}
}

// Model is a drawing over the region xMin..xMax by yMin..yMax, y growing
// upwards. A value: every method returns the new canvas.
type Model struct {
	xMin, xMax, yMin, yMax float64
	width, height          int
	shapes                 []Shape
	edit                   func(Styles) Styles
}

// New is an empty canvas over x from xMin to xMax and y from yMin to yMax.
// It draws nothing until SetWidth and SetHeight size it.
func New(xMin, xMax, yMin, yMax float64) Model {
	return Model{xMin: xMin, xMax: xMax, yMin: yMin, yMax: yMax}
}

// WithStyles edits the styles the theme gives, again on every View.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

// SetWidth is the canvas's width in cells, two dots each.
func (m Model) SetWidth(w int) Model {
	m.width = max(w, 0)
	return m
}

// SetHeight is the canvas's height in cells, four dots each.
func (m Model) SetHeight(h int) Model {
	m.height = max(h, 0)
	return m
}

// SetBounds moves the region the canvas shows, as a plot that scrolls.
func (m Model) SetBounds(xMin, xMax, yMin, yMax float64) Model {
	m.xMin, m.xMax, m.yMin, m.yMax = xMin, xMax, yMin, yMax
	return m
}

// Draw adds shapes on top of those already drawn.
func (m Model) Draw(shapes ...Shape) Model {
	m.shapes = append(m.shapes[:len(m.shapes):len(m.shapes)], shapes...)
	return m
}

// Clear removes every shape.
func (m Model) Clear() Model {
	m.shapes = nil
	return m
}

// Width is the canvas's width in cells.
func (m Model) Width() int { return m.width }

// Height is the canvas's height in cells.
func (m Model) Height() int { return m.height }

// Bounds is the region the canvas shows.
func (m Model) Bounds() (xMin, xMax, yMin, yMax float64) {
	return m.xMin, m.xMax, m.yMin, m.yMax
}

// Shapes is what was drawn, oldest first.
func (m Model) Shapes() []Shape { return m.shapes[:len(m.shapes):len(m.shapes)] }

// View draws the shapes in order, a later one taking a shared cell's colour.
// An empty region (min not below max, or not finite) draws blank cells.
func (m Model) View(th theme.Resolved) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	g := braille.New(m.width, m.height)
	if p, ok := m.painter(g); ok {
		for _, s := range m.shapes {
			if s != nil {
				s.draw(p)
			}
		}
	}
	rows := make([]string, m.height)
	for r := range rows {
		rows[r] = g.Row(r, st.Series)
	}
	return strings.Join(rows, "\n")
}

func (m Model) painter(g braille.Grid) (painter, bool) {
	ok := finite(m.xMin, m.xMax, m.yMin, m.yMax) && m.xMax > m.xMin && m.yMax > m.yMin
	return painter{g: g, m: m}, ok
}

func finite(vs ...float64) bool {
	for _, v := range vs {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}

// painter turns the app's units into dots on a grid.
type painter struct {
	g braille.Grid
	m Model
}

// at is x, y in dots, not rounded and not kept within the grid.
func (p painter) at(x, y float64) (float64, float64) {
	dw, dh := p.g.Dots()
	fx := (x - p.m.xMin) / (p.m.xMax - p.m.xMin) * float64(dw-1)
	fy := (p.m.yMax - y) / (p.m.yMax - p.m.yMin) * float64(dh-1)
	return fx, fy
}

// dot is the dot nearest x, y, and false when it is off the grid.
func (p painter) dot(x, y float64) (int, int, bool) {
	fx, fy := p.at(x, y)
	dw, dh := p.g.Dots()
	ix, iy := math.Round(fx), math.Round(fy)
	if ix < 0 || iy < 0 || ix >= float64(dw) || iy >= float64(dh) {
		return 0, 0, false
	}
	return int(ix), int(iy), true
}

// scale is how many dots one unit is across and up.
func (p painter) scale() (sx, sy float64) {
	dw, dh := p.g.Dots()
	return float64(dw-1) / (p.m.xMax - p.m.xMin), float64(dh-1) / (p.m.yMax - p.m.yMin)
}

// line draws the part of x0, y0 to x1, y1 that crosses the grid, cut at its
// edges first so a line from far away costs no more than one inside.
func (p painter) line(x0, y0, x1, y1 float64, series int) {
	if !finite(x0, y0, x1, y1) {
		return
	}
	ax, ay := p.at(x0, y0)
	bx, by := p.at(x1, y1)
	dw, dh := p.g.Dots()
	t0, t1, ok := clip(ax, ay, bx-ax, by-ay, -0.5, -0.5, float64(dw)-0.5, float64(dh)-0.5)
	if !ok {
		return
	}
	dx, dy := bx-ax, by-ay
	p.g.Line(round(ax+t0*dx), round(ay+t0*dy), round(ax+t1*dx), round(ay+t1*dy), series)
}

func round(f float64) int { return int(math.Round(f)) }

// clip is the share of the segment from x, y by dx, dy that lies inside the
// box from lx, ly to hx, hy (Liang-Barsky), and false when none does.
func clip(x, y, dx, dy, lx, ly, hx, hy float64) (t0, t1 float64, ok bool) {
	t0, t1 = 0, 1
	for _, e := range [4][2]float64{{-dx, x - lx}, {dx, hx - x}, {-dy, y - ly}, {dy, hy - y}} {
		q, r := e[0], e[1]
		if q == 0 {
			if r < 0 {
				return 0, 0, false
			}
			continue
		}
		t := r / q
		if q < 0 {
			t0 = max(t0, t)
		} else {
			t1 = min(t1, t)
		}
	}
	return t0, t1, t0 <= t1
}
