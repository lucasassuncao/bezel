package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestADropStretchesAsItFallsAndSquashesAsItLands(t *testing.T) {
	moving(t)
	d := NewDrop(0, 10)
	start := time.Unix(10, 0)
	d, _ = d.Advance(start)
	w, h := d.Shape()
	assert.InDelta(t, 1, w*h, 1e-9)

	falling, _ := d.Advance(start.Add(500 * ms))
	_, h = falling.Shape()
	assert.Greater(t, h, 1.2, "fast near the bottom: drawn tall")

	var squashed bool
	for at := 400 * ms; at < 700*ms; at += ms {
		hit, _ := d.Advance(start.Add(at))
		if w, h := hit.Shape(); w > 1.2 && h < 1 {
			squashed = true
			assert.InDelta(t, 1, w*h, 1e-9)
		}
	}
	assert.True(t, squashed, "some frame at the floor is wide and flat")
}

func TestASettledDropIsRound(t *testing.T) {
	moving(t)
	d := NewDrop(0, 4)
	start := time.Unix(10, 0)
	d, _ = d.Advance(start)
	d, _ = d.Advance(start.Add(10 * time.Second))
	w, h := d.Shape()
	assert.Equal(t, 1.0, w)
	assert.Equal(t, 1.0, h)
}
