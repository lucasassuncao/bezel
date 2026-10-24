package effect

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/cells"
)

func TestGhostFadesTheBlockTowardsTheBackground(t *testing.T) {
	assert.Equal(t, grey, toRGBA(cells.Parse(Ghost(whiteText, black, 0.5))[0][0].Fg))
	assert.Equal(t, "abcd", ansi.Strip(Ghost(whiteText, black, 0.5)))
	assert.Equal(t, white, toRGBA(cells.Parse(Ghost(whiteText, black, 7))[0][0].Fg), "alpha past 1 is the block itself")
}
