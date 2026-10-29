package scene

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/cells"
)

func TestBubblesRiseAndPopAtTheTop(t *testing.T) {
	popped := false
	for s := 1.0; s < 4; s += 0.1 {
		if strings.Contains(strings.Split(ansi.Strip(playedFor(Bubbles(white), s).View(40, 10)), "\n")[0], "*") {
			popped = true
		}
	}
	assert.True(t, popped)
}

func TestTheSeaFillsTheBottomUnderFoam(t *testing.T) {
	g := cells.Parse(playedFor(Sea(white, grey), 1).View(30, 10))
	assert.NotNil(t, g[9][0].Bg, "the bottom row is water")
	assert.Nil(t, g[0][0].Bg, "the top is sky")
	assert.Contains(t, ansi.Strip(playedFor(Sea(white, grey), 1).View(30, 10)), "~")
}
