package animation

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// playedFor is a advanced to secs seconds after its first frame.
func playedFor(a Ambient, secs float64) Ambient {
	start := time.Unix(1000, 0)
	a, _ = a.Advance(start)
	a, _ = a.Advance(start.Add(time.Duration(secs * float64(time.Second))))
	return a
}

// size is the width of every line of s and how many lines it has.
func size(t *testing.T, s string) (int, int) {
	t.Helper()
	lines := strings.Split(s, "\n")
	w := ansi.StringWidth(lines[0])
	for _, l := range lines {
		require.Equal(t, w, ansi.StringWidth(l), "every line is as wide")
	}
	return w, len(lines)
}

// Every scene fills the size asked and changes as time passes.
func TestAmbientScenesFillTheirSizeAndMove(t *testing.T) {
	for name, a := range map[string]Ambient{
		"rain": NewRain(white, grey), "snow": NewSnow(white), "confetti": NewConfetti(white, black),
		"fireworks": NewFireworks(white), "stars": NewStarfield(white), "static": NewStatic(white), "plasma": NewPlasma(black, white),
	} {
		first := playedFor(a, 1).View(40, 12)
		w, h := size(t, first)
		assert.Equal(t, 40, w, name)
		assert.Equal(t, 12, h, name)
		assert.NotEqual(t, first, playedFor(a, 2.5).View(40, 12), name)
		assert.True(t, a.Active(), name)
	}
}

func TestAmbientIsAFunctionOfTime(t *testing.T) {
	a := NewRain(white, grey)
	assert.Equal(t, playedFor(a, 3).View(30, 8), playedFor(a, 3).View(30, 8))
}

func TestAmbientUnderKeepsTheBlockOnTop(t *testing.T) {
	out := playedFor(NewStatic(white), 1).Under("hello\n     ")
	lines := strings.Split(ansi.Strip(out), "\n")
	assert.Equal(t, "hello", lines[0])
	assert.Len(t, lines, 2)
}

func TestAmbientPlaysOnItsOwnFrames(t *testing.T) {
	a := NewSnow(white)
	a, cmd := a.Update(a.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, cmd)
	a, done := a.Advance(time.Unix(100, 0))
	assert.False(t, done, "a scene never ends on its own")
	assert.Equal(t, "", a.View(0, 3))
}

func TestStillAmbientHoldsItsFirstFrame(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	a := NewStarfield(white)
	assert.False(t, a.Active())
	assert.Equal(t, a.View(20, 5), playedFor(a, 5).View(20, 5))
}
