package text

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestRainbowTurnsItsColours(t *testing.T) {
	start := time.Unix(1000, 0)
	r := NewRainbow("abcdef", time.Second)
	first := r.colorAt(0)
	assert.NotEqual(t, first, r.colorAt(1), "neighbours sit at different hues")

	r, _ = r.Advance(start)
	r, _ = r.Advance(start.Add(250 * time.Millisecond))
	assert.NotEqual(t, first, r.colorAt(0), "the wheel has turned")
	assert.Equal(t, "abcdef", ansi.Strip(r.View()))
}

func TestRainbowWithToneKeepsTheThemeLightness(t *testing.T) {
	r := NewRainbow("ab", time.Second).WithTone(white)
	assert.Equal(t, white, r.colorAt(0), "white has no saturation, so every hue is white")
}
