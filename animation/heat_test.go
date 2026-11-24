package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHeatMovesEachCellToItsNewValue(t *testing.T) {
	start := time.Unix(1000, 0)
	h := NewHeat([][]float64{{0, 10}}, 100*ms)
	require.Equal(t, [][]float64{{0, 10}}, h.Values())
	h, cmd := h.Set([][]float64{{10, 0}, {4}})
	require.NotNil(t, cmd)
	h.run.Ease = Linear
	h, _ = h.Advance(start)
	h, _ = h.Advance(start.Add(50 * ms))
	require.Equal(t, [][]float64{{5, 5}, {2}}, h.Values(), "a new cell rises from zero")
	h, done := h.Advance(start.Add(time.Second))
	require.True(t, done)
	require.Equal(t, [][]float64{{10, 0}, {4}}, h.Values())
}
