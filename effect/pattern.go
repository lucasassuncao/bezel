package effect

import (
	"math"

	"github.com/lucasassuncao/bezel/internal/noise"
)

// Pattern is the order an effect reaches a block's cells in: for the cell at
// column x of row y in a w×h block, the share of the way, in [0,1], at which
// it turns. Rows count double, since a cell is about twice as tall as wide.
type Pattern func(x, y, w, h int) float64

// Direction is the way a sweep or a slide travels.
type Direction int

const (
	Right Direction = iota
	Left
	Down
	Up
)

// share is i of n-1, 0 for a single row or column.
func share(i, n int) float64 {
	if n <= 1 {
		return 0
	}
	return float64(i) / float64(n-1)
}

// Sweep reaches the cells in a straight front travelling d.
func Sweep(d Direction) Pattern {
	return func(x, y, w, h int) float64 {
		switch d {
		case Left:
			return 1 - share(x, w)
		case Down:
			return share(y, h)
		case Up:
			return 1 - share(y, h)
		}
		return share(x, w)
	}
}

// centre is how far x, y sits from the block's middle on each axis, rows
// doubled, and the farthest a cell can be.
func centre(x, y, w, h int) (dx, dy, mx, my float64) {
	cx, cy := float64(w-1)/2, float64(h-1)/2
	return math.Abs(float64(x) - cx), 2 * math.Abs(float64(y)-cy), cx, 2 * cy
}

// Radial reaches the cells in a widening circle from the middle.
func Radial() Pattern {
	return func(x, y, w, h int) float64 {
		dx, dy, mx, my := centre(x, y, w, h)
		if far := math.Hypot(mx, my); far > 0 {
			return math.Hypot(dx, dy) / far
		}
		return 0
	}
}

// Diamond reaches the cells in a widening diamond from the middle.
func Diamond() Pattern {
	return func(x, y, w, h int) float64 {
		dx, dy, mx, my := centre(x, y, w, h)
		if far := mx + my; far > 0 {
			return (dx + dy) / far
		}
		return 0
	}
}

// Diagonal reaches the cells from the top left corner to the bottom right.
func Diagonal() Pattern {
	return func(x, y, w, h int) float64 {
		if far := float64(w-1) + 2*float64(h-1); far > 0 {
			return (float64(x) + 2*float64(y)) / far
		}
		return 0
	}
}

// Checker reaches every other cell first, then the rest.
func Checker() Pattern {
	return func(x, y, _, _ int) float64 { return float64((x + y) % 2) }
}

// Spiral reaches the cells along arms turning out from the middle.
func Spiral(arms int) Pattern {
	arms = max(arms, 1)
	return func(x, y, w, h int) float64 {
		cx, cy := float64(w-1)/2, float64(h-1)/2
		a := (math.Atan2(2*(float64(y)-cy), float64(x)-cx) + math.Pi) / (2 * math.Pi)
		r := Radial()(x, y, w, h)
		v := a/float64(arms) + r
		return v - math.Floor(v)
	}
}

// Wave reaches the cells in a front travelling right whose edge ripples.
func Wave() Pattern {
	return func(x, y, w, h int) float64 {
		ripple := 0.12 * math.Sin(2*math.Pi*share(y, h)*1.5)
		return 0.12 + 0.76*share(x, w) + ripple
	}
}

// Noise reaches the cells in an order that looks random, the same for the
// same seed, so a dissolve does not flicker from one frame to the next.
func Noise(seed uint64) Pattern {
	return func(x, y, w, _ int) float64 {
		k := uint64(max(y*w+x, 0)) // #nosec G115 -- clamped to zero or more
		return noise.Unit(noise.Hash(seed, k))
	}
}

// Inverted reaches the cells in p's order turned round: last first.
func Inverted(p Pattern) Pattern {
	return func(x, y, w, h int) float64 { return 1 - p(x, y, w, h) }
}
