package effect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEveryPatternStaysInRange(t *testing.T) {
	for name, p := range map[string]Pattern{
		"sweep right": Sweep(Right), "sweep up": Sweep(Up), "radial": Radial(), "diamond": Diamond(),
		"diagonal": Diagonal(), "checker": Checker(), "spiral": Spiral(2), "wave": Wave(),
		"noise": Noise(7), "inverted": Inverted(Radial()),
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
