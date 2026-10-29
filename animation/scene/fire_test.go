package scene

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestFireBurnsHottestAtTheBottom(t *testing.T) {
	rows := strings.Split(ansi.Strip(playedFor(Fire(white, grey), 1).View(30, 10)), "\n")
	assert.Greater(t, strings.Count(rows[9], "█")+strings.Count(rows[9], "▓"), strings.Count(rows[0], "█")+strings.Count(rows[0], "▓"))
	assert.Equal(t, strings.Repeat(" ", 30), rows[0], "the top is above the flames")
}

func TestSmoothNoiseChangesGently(t *testing.T) {
	assert.InDelta(t, smooth(1, 2, 3), smooth(1, 2.01, 3), 0.05)
	assert.Equal(t, smooth(4, 1.5, 2.5), smooth(4, 1.5, 2.5))
	v := smooth(7, 0.3, 0.8)
	assert.True(t, v >= 0 && v < 1)
}

func TestAuroraHangsInTheSkyAndNeedsAColour(t *testing.T) {
	rows := strings.Split(ansi.Strip(playedFor(Aurora(white, grey), 2).View(40, 12)), "\n")
	assert.Equal(t, strings.Repeat(" ", 40), rows[11], "the ground is dark")
	assert.NotEqual(t, strings.Repeat(" ", 40), rows[4])
	assert.Equal(t, strings.Repeat(" ", 10), ansi.Strip(playedFor(Aurora(), 1).View(10, 1)))
}
