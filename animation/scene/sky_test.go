package scene

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/cells"
)

func TestStarfieldDrifts(t *testing.T) {
	a := Starfield(white)
	assert.NotEqual(t, ansi.Strip(playedFor(a, 1).View(60, 10)), ansi.Strip(playedFor(a, 4).View(60, 10)))
}

func TestStaticChangesEveryFrame(t *testing.T) {
	a := Static(white)
	assert.NotEqual(t, playedFor(a, 1).View(10, 3), playedFor(a, 1.05).View(10, 3))
}

func TestPlasmaPaintsTheBackground(t *testing.T) {
	g := cells.Parse(playedFor(Plasma(black, white), 1).View(8, 2))
	assert.NotNil(t, g[0][0].Bg)
	assert.Equal(t, " ", g[0][0].Text)
}
