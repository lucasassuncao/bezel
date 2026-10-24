package scene

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestRainFallsDownItsColumns(t *testing.T) {
	a := NewRain(white, grey)
	assert.NotEqual(t, strings.Repeat(" ", 20), strings.Split(ansi.Strip(playedFor(a, 2).View(20, 10)), "\n")[5], "drops cross the middle")
}

func TestConfettiTurnsAndColours(t *testing.T) {
	v := playedFor(NewConfetti(white, black), 1).View(30, 10)
	assert.Contains(t, v, "\x1b[")
}

// A firework climbs from the bottom row before it bursts.
func TestFireworksClimbThenBurst(t *testing.T) {
	a := NewFireworks(white)
	early := strings.Split(ansi.Strip(playedFor(a, 0.05).View(30, 12)), "\n")
	assert.Contains(t, early[11]+early[10], "|")
	late := ansi.Strip(playedFor(a, 1).View(30, 12))
	assert.Contains(t, late, "*")
}

func TestWrap(t *testing.T) {
	assert.Equal(t, 9, wrap(-1, 10))
	assert.Equal(t, 0, wrap(10, 10))
}
