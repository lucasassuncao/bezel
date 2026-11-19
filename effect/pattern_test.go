package effect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEveryPatternStaysInRange(t *testing.T) {
	for name, p := range map[string]Pattern{
		"sweep right": Sweep(Right), "sweep up": Sweep(Up), "radial": Radial(), "diamond": Diamond(),
		"diagonal": Diagonal(), "checker": Checker(), "spiral": Spiral(2), "wave": Wave(),
		"noise": Noise(7), "inverted": Inverted(Radial()), "radial from": RadialFrom(2, 1),
		"radial from outside": RadialFrom(-5, 9), "bayer": Bayer(), "interlace": Interlace(), "columns": Columns(3),
	} {
		for _, size := range [][2]int{{1, 1}, {7, 5}, {40, 3}} {
			for y := range size[1] {
				for x := range size[0] {
					v := p(x, y, size[0], size[1])
					assert.True(t, v >= 0 && v <= 1, "%s at %d,%d of %v: %v", name, x, y, size, v)
				}
			}
		}
	}
}

func TestSweepsGoTheirWay(t *testing.T) {
	assert.Less(t, Sweep(Right)(0, 0, 10, 3), Sweep(Right)(9, 0, 10, 3))
	assert.Greater(t, Sweep(Left)(0, 0, 10, 3), Sweep(Left)(9, 0, 10, 3))
	assert.Less(t, Sweep(Down)(0, 0, 10, 3), Sweep(Down)(0, 2, 10, 3))
	assert.Greater(t, Sweep(Up)(0, 0, 10, 3), Sweep(Up)(0, 2, 10, 3))
}

func TestRadialStartsInTheMiddle(t *testing.T) {
	r := Radial()
	assert.Equal(t, 0.0, r(3, 2, 7, 5))
	assert.InDelta(t, 1, r(0, 0, 7, 5), 1e-9)
	assert.Equal(t, 1.0, Inverted(r)(3, 2, 7, 5))
}

func TestCheckerAlternates(t *testing.T) {
	c := Checker()
	assert.NotEqual(t, c(0, 0, 4, 4), c(1, 0, 4, 4))
	assert.Equal(t, c(0, 0, 4, 4), c(1, 1, 4, 4))
}

func TestNoiseIsTheSameEveryTimeForASeed(t *testing.T) {
	assert.Equal(t, Noise(3)(4, 1, 9, 9), Noise(3)(4, 1, 9, 9))
	assert.NotEqual(t, Noise(3)(4, 1, 9, 9), Noise(4)(4, 1, 9, 9))
}

func TestRadialFromStartsAtItsCell(t *testing.T) {
	r := RadialFrom(1, 1)
	assert.Equal(t, 0.0, r(1, 1, 7, 5))
	assert.Less(t, r(2, 1, 7, 5), r(5, 1, 7, 5))
	assert.InDelta(t, 1, r(6, 4, 7, 5), 1e-9, "the farthest corner is the last")
}

func TestBayerLightsEveryCellOfItsTileOnce(t *testing.T) {
	seen := map[float64]bool{}
	for y := range 4 {
		for x := range 4 {
			seen[Bayer()(x, y, 4, 4)] = true
		}
	}
	assert.Len(t, seen, 16)
	assert.Equal(t, Bayer()(0, 0, 9, 9), Bayer()(4, 4, 9, 9), "the tile repeats")
}

func TestInterlaceTakesTheEvenRowsFirst(t *testing.T) {
	i := Interlace()
	for _, even := range []int{0, 2, 4} {
		for _, odd := range []int{1, 3} {
			assert.Less(t, i(0, even, 3, 5), i(0, odd, 3, 5), "row %d before row %d", even, odd)
		}
	}
	assert.Less(t, i(0, 0, 3, 5), i(0, 2, 3, 5), "top down within the evens")
}

func TestColumnsRunDownEachColumnFromItsOwnStart(t *testing.T) {
	c := Columns(5)
	assert.Less(t, c(2, 0, 9, 6), c(2, 5, 9, 6))
	starts := map[float64]bool{}
	for x := range 9 {
		starts[c(x, 0, 9, 6)] = true
	}
	assert.Greater(t, len(starts), 1, "the columns do not start together")
}
