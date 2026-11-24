package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRubberGivesPastItsLimitsAndSpringsBack(t *testing.T) {
	r := NewRubber(0, 10).WithGive(4)
	require.InDelta(t, 5.0, r.Pull(5).Value(), 0, "inside the limits it goes where pulled")
	far := r.Pull(1000)
	require.Less(t, far.Value(), 14.0, "it never gets more than give past a limit")
	require.Greater(t, far.Value(), 13.0)
	require.InDelta(t, -2, r.Pull(-4).Value(), 1e-9, "4 past the floor resists to 2 past it")

	start := time.Unix(1000, 0)
	back, cmd := far.Release()
	require.NotNil(t, cmd)
	back, _ = back.Advance(start)
	back, done := back.Advance(start.Add(5 * time.Second))
	require.True(t, done)
	require.Equal(t, 10.0, back.Value())
}
