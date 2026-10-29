package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTickerRollsAndFlashesItsDirection(t *testing.T) {
	start := time.Unix(1000, 0)
	tk := NewTicker(100, time.Second).WithBase(white).WithUp(black).WithDown(black)
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
	tk := NewTicker(0, time.Second).WithBase(white).WithUp(black).WithDown(black)
	tk, _ = tk.To(100)
	tk, _ = tk.Advance(start)
	tk, _ = tk.Advance(start.Add(500 * time.Millisecond))
	mid := tk.Value()
	tk, _ = tk.To(0)
	assert.Equal(t, mid, tk.Value())
}

// Without colours the number rolls in the terminal's own.
func TestTickerWithoutColoursRollsUncoloured(t *testing.T) {
	tk, _ := NewTicker(1, time.Second).To(9)
	tk, _ = tk.Advance(time.Unix(1000, 0))
	assert.Nil(t, tk.Color())
	assert.Equal(t, "1", tk.View())
}
