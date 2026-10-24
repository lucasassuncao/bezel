package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHeatMovesEachCellToItsNewValue(t *testing.T) {
	start := time.Unix(1000, 0)
	h := NewHeat([][]float64{{0, 10}}, 100*ms)
	assert.Equal(t, [][]float64{{0, 10}}, h.Values())
	h, cmd := h.Set([][]float64{{10, 0}, {4}})
	assert.NotNil(t, cmd)
	h.run.Ease = Linear
	h, _ = h.Advance(start)
	h, _ = h.Advance(start.Add(50 * ms))
	assert.Equal(t, [][]float64{{5, 5}, {2}}, h.Values(), "a new cell rises from zero")
	h, done := h.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, [][]float64{{10, 0}, {4}}, h.Values())
}
