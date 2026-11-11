package canvas

import (
	"math"

	"github.com/charmbracelet/x/ansi"
)

// Shape is what Draw takes: Line, Rect, Circle, Points or Label. Series picks
// the shape's colour from Styles.Series, wrapping round it.
type Shape interface {
	draw(p painter)
}

// Line runs from X0, Y0 to X1, Y1.
type Line struct {
	X0, Y0, X1, Y1 float64
	Series         int
}

func (l Line) draw(p painter) { p.line(l.X0, l.Y0, l.X1, l.Y1, l.Series) }

// Rect is the outline of the box from X, Y (its bottom left corner), W wide
// and H tall.
type Rect struct {
	X, Y, W, H float64
	Series     int
}

func (r Rect) draw(p painter) {
	x1, y1 := r.X+r.W, r.Y+r.H
	p.line(r.X, r.Y, x1, r.Y, r.Series)
	p.line(x1, r.Y, x1, y1, r.Series)
	p.line(x1, y1, r.X, y1, r.Series)
	p.line(r.X, y1, r.X, r.Y, r.Series)
}

// Circle is the outline round X, Y at radius R, in the canvas's units, so
// a region wider than it is tall draws it as an ellipse.
type Circle struct {
	X, Y, R float64
	Series  int
}

func (c Circle) draw(p painter) {
	if !finite(c.X, c.Y, c.R) || c.R <= 0 {
		return
	}
	sx, sy := p.scale()
	// A step a dot of the longer way round, each joined by a line, keeps it whole.
	n := int(min(max(2*math.Pi*c.R*max(sx, sy), 8), 4096))
	px, py := c.X+c.R, c.Y
	for i := 1; i <= n; i++ {
		a := 2 * math.Pi * float64(i) / float64(n)
		x, y := c.X+c.R*math.Cos(a), c.Y+c.R*math.Sin(a)
		p.line(px, py, x, y, c.Series)
		px, py = x, y
	}
}

// Point is a place on the canvas, in its units.
type Point struct{ X, Y float64 }

// Points is one dot at each of Coords, as a scatter plot.
type Points struct {
	Coords []Point
	Series int
}

func (ps Points) draw(p painter) {
	for _, c := range ps.Coords {
		if finite(c.X, c.Y) {
			if x, y, ok := p.dot(c.X, c.Y); ok {
				p.g.Set(x, y, ps.Series)
			}
		}
	}
}

// Label is Text written from the cell that holds X, Y, over any dots there;
// what passes the right edge is cut, and escape codes are dropped.
type Label struct {
	X, Y   float64
	Text   string
	Series int
}

func (l Label) draw(p painter) {
	if !finite(l.X, l.Y) {
		return
	}
	x, y, ok := p.dot(l.X, l.Y)
	if !ok {
		return
	}
	p.g.Text(x/2, y/4, ansi.Strip(l.Text), l.Series)
}
