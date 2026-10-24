package frames

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsTakesOnlyTheFrameForItsIDAndStep(t *testing.T) {
	at := time.Unix(10, 0)
	f, ok := Is(NewMsg(at, 3, 1), 3, 1)
	assert.True(t, ok)
	assert.Equal(t, at, f.At)
	_, ok = Is(NewMsg(at, 3, 0), 3, 1)
	assert.False(t, ok, "a twin from an earlier step")
	_, ok = Is(NewMsg(at, 4, 1), 3, 1)
	assert.False(t, ok, "another animation's frame")
	_, ok = Is("key", 3, 1)
	assert.False(t, ok)
}

func TestTickDeliversTheFrameItWasAskedFor(t *testing.T) {
	_, ok := Is(Tick(time.Millisecond, 5, 2)(), 5, 2)
	assert.True(t, ok)
}

func TestNextIDNeverRepeats(t *testing.T) {
	assert.NotEqual(t, NextID(), NextID())
}
