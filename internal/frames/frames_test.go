package frames

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsTakesOnlyTheFrameForItsIDAndStep(t *testing.T) {
	at := time.Unix(10, 0)
	f, ok := Is(NewMsg(at, 3, 1), 3, 1)
	require.True(t, ok)
	require.Equal(t, at, f.At)
	_, ok = Is(NewMsg(at, 3, 0), 3, 1)
	require.False(t, ok, "a twin from an earlier step")
	_, ok = Is(NewMsg(at, 4, 1), 3, 1)
	require.False(t, ok, "another animation's frame")
	_, ok = Is("key", 3, 1)
	require.False(t, ok)
}

func TestTickDeliversTheFrameItWasAskedFor(t *testing.T) {
	_, ok := Is(Tick(time.Millisecond, 5, 2)(), 5, 2)
	require.True(t, ok)
}

func TestNextIDNeverRepeats(t *testing.T) {
	require.NotEqual(t, NextID(), NextID())
}
