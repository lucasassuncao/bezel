package text

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/animation"
)

func TestHighlightMarksFromTheLeftAndStays(t *testing.T) {
	start := time.Unix(1000, 0)
	h := NewHighlight("found it", black, time.Second).WithEase(animation.Linear)
	assert.Equal(t, "found it", h.View(), "nothing marked before the first frame")
	h, _ = h.Advance(start)
	h, _ = h.Advance(start.Add(500 * time.Millisecond))
	v := h.View()
	assert.Equal(t, "found it", ansi.Strip(v))
	assert.NotEqual(t, "found it", v, "half the text sits on the marker")

	h, done := h.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Contains(t, h.View(), "found it")
}

func TestStillHighlightIsBornMarked(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	h := NewHighlight("x", black, time.Second)
	assert.False(t, h.Active())
	assert.NotEqual(t, "x", h.View())
}
