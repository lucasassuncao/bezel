package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTickerRollsAndFlashesItsDirection(t *testing.T) {
	start := time.Unix(1000, 0)
	tk := NewTicker(100, white, black, black, time.Second)
	assert.Equal(t, 100, tk.Value())
	assert.Equal(t, white, tk.Color())

	tk, cmd := tk.To(200)
	assert.NotNil(t, cmd)
	tk, _ = tk.Advance(start)
	assert.Equal(t, black, tk.Color(), "it starts in its direction's colour")
	tk, _ = tk.Advance(start.Add(500 * time.Millisecond))
	assert.Greater(t, tk.Value(), 100)
	assert.Less(t, tk.Value(), 200)

	tk, done := tk.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, 200, tk.Value())
	assert.Equal(t, white, tk.Color(), "settled back to base")
}

// A new value mid-roll starts from the number on screen, not the old target.
func TestTickerToMidRollStartsWhereItIs(t *testing.T) {
	start := time.Unix(1000, 0)
	tk := NewTicker(0, white, black, black, time.Second)
	tk, _ = tk.To(100)
	tk, _ = tk.Advance(start)
	tk, _ = tk.Advance(start.Add(500 * time.Millisecond))
	mid := tk.Value()
	tk, _ = tk.To(0)
	assert.Equal(t, mid, tk.Value())
}
