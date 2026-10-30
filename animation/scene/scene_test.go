package scene

import (
	"image/color"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
)

// playedFor is a advanced to secs seconds after its first frame.
func playedFor(a Scene, secs float64) Scene {
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
func TestSceneScenesFillTheirSizeAndMove(t *testing.T) {
	for name, a := range map[string]Scene{
		"rain": Rain(white, grey), "snow": Snow(white), "confetti": Confetti(white, black),
		"fireworks": Fireworks(white), "stars": Starfield(white), "static": Static(white), "plasma": Plasma(black, white),
		"fire": Fire(white, grey), "aurora": Aurora(white, grey), "bubbles": Bubbles(white), "sea": Sea(white, grey),
		"life": Life(white),
	} {
		first := playedFor(a, 1).View(40, 12)
		w, h := size(t, first)
		assert.Equal(t, 40, w, name)
		assert.Equal(t, 12, h, name)
		assert.NotEqual(t, first, playedFor(a, 2.5).View(40, 12), name)
		assert.True(t, a.Active(), name)
	}
}

func TestSceneIsAFunctionOfTime(t *testing.T) {
	a := Rain(white, grey)
	assert.Equal(t, playedFor(a, 3).View(30, 8), playedFor(a, 3).View(30, 8))
}

func TestSceneUnderKeepsTheBlockOnTop(t *testing.T) {
	out := playedFor(Static(white), 1).Under("hello\n     ")
	lines := strings.Split(ansi.Strip(out), "\n")
	assert.Equal(t, "hello", lines[0])
	assert.Len(t, lines, 2)
}

func TestScenePlaysOnItsOwnFrames(t *testing.T) {
	a := Snow(white)
	a, res := a.Update(a.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, res.Cmd)
	a, done := a.Advance(time.Unix(100, 0))
	assert.False(t, done, "a scene never ends on its own")
	assert.Equal(t, "", a.View(0, 3))
}

func TestStillSceneHoldsItsFirstFrame(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	a := Starfield(white)
	assert.False(t, a.Active())
	assert.Equal(t, a.View(20, 5), playedFor(a, 5).View(20, 5))
}

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	grey  = color.RGBA{128, 128, 128, 255}
)

// TestMain clears the motion switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("BEZEL_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

func TestOwnFrameIsHandledAndAsksForTheNext(t *testing.T) {
	x := Snow(white)
	_, r := x.Update(x.FrameAt(time.Unix(1000, 0)))
	require.True(t, r.Handled)
	require.NotNil(t, r.Cmd)
}

func TestForeignFrameIsNotHandled(t *testing.T) {
	a, b := Snow(white), Snow(white)
	_, r := a.Update(b.FrameAt(time.Unix(1000, 0)))
	require.Equal(t, bezel.Result{}, r)
}
