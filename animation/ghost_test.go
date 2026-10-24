package animation

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestGhostFadesTheBlockTowardsTheBackground(t *testing.T) {
	assert.Equal(t, grey, toRGBA(parse(Ghost(whiteText, black, 0.5))[0][0].fg))
	assert.Equal(t, "abcd", ansi.Strip(Ghost(whiteText, black, 0.5)))
	assert.Equal(t, white, toRGBA(parse(Ghost(whiteText, black, 7))[0][0].fg), "alpha past 1 is the block itself")
}
