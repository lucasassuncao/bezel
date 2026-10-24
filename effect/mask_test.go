package effect

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/cells"
)

func TestCropAndPasteRoundTrip(t *testing.T) {
	g := cells.Parse("abcdef\nghijkl\nmnopqr")
	r := rect{x: 1, y: 1, w: 3, h: 1}
	sub := crop(g, r)
	assert.Equal(t, "hij", sub.String())
	sub[0][0].Text = "X"
	assert.Equal(t, "abcdef\ngXijkl\nmnopqr", paste(g, sub, r).String())
}

func TestRectRunsToTheEdge(t *testing.T) {
	r := rect{x: 2, y: 1}
	assert.True(t, r.holds(40, 9))
	assert.False(t, r.holds(1, 1))
	assert.Equal(t, "cdef", crop(cells.Parse("abcdef"), rect{x: 2}).String())
}

// OnRows changes the rows named and nothing else; Within a part of a row.
func TestEffectsKeepToTheirMasks(t *testing.T) {
	block := "abcd\nefgh\nijkl"
	out := ansi.Strip(landed(t, Dissolve(100*ms).OnRows(1)).Apply(block))
	assert.Equal(t, "abcd\n    \nijkl", out)
	out = ansi.Strip(landed(t, Dissolve(100*ms).Within(1, 0, 2, 0)).Apply(block))
	assert.Equal(t, "a  d\ne  h\ni  l", out)
}

// Reverse plays an effect backwards: Dissolve reversed ends with the block whole.
func TestReverseTurnsAnEffectRound(t *testing.T) {
	assert.Equal(t, "abcd", ansi.Strip(landed(t, Dissolve(100*ms).Reverse()).Apply("abcd")))
}
