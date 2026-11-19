package effect

import (
	"image/color"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/cells"
)

const screen = "abcdefgh\nijklmnop\nqrstuvwx\nyz012345\n6789ABCD"

// at is e played to p of its run.
func at(e Effect, p float64) Effect {
	start := time.Unix(10, 0)
	e, _ = e.Advance(start)
	e, _ = e.Advance(start.Add(time.Duration(p * float64(e.run.Dur))))
	return e
}

func TestGlitchTearsRowsAndComesBackWhole(t *testing.T) {
	e := Glitch(500 * ms)
	torn := false
	for p := 0.05; p < 1; p += 0.05 {
		out := at(e, p).Apply(screen)
		sized(t, out, 8, 5)
		if ansi.Strip(out) != screen {
			torn = true
		}
	}
	assert.True(t, torn, "some frame tears a row")
	assert.Equal(t, screen, landed(t, e).Apply(screen))
}

func TestPowerOffSquashesToALineThenADotThenNothing(t *testing.T) {
	e := PowerOff(white, 1000*ms)
	squashed := litRows(at(e, 0.3).Apply(screen))
	assert.Positive(t, squashed)
	assert.Less(t, squashed, 5, "fewer rows, squashed towards the middle")
	line := ansi.Strip(at(e, 0.7).Apply(screen))
	assert.Contains(t, line, "─")
	dot := ansi.Strip(at(e, 0.95).Apply(screen))
	assert.Equal(t, 1, strings.Count(dot, "•"))
	assert.Zero(t, litRows(landed(t, e).Apply(screen)))
	for _, p := range []float64{0.3, 0.7, 0.95} {
		sized(t, at(e, p).Apply(screen), 8, 5)
	}
}

// litRows is how many rows of block draw anything.
func litRows(block string) int {
	n := 0
	for _, l := range strings.Split(ansi.Strip(block), "\n") {
		if strings.TrimSpace(l) != "" {
			n++
		}
	}
	return n
}

func TestPowerOnEndsWithTheBlock(t *testing.T) {
	e := PowerOn(white, 300*ms)
	assert.Equal(t, screen, landed(t, e).Apply(screen))
	assert.Contains(t, ansi.Strip(at(e, 0.08).Apply(screen)), "•", "it starts from a dot")
}

func TestScanlinesDarkenEveryOtherRow(t *testing.T) {
	lit := "\x1b[38;2;200;200;200mab\nab\nab\x1b[m"
	g := cells.Parse(landed(t, Scanlines(black, 100*ms)).Apply(lit))
	top, gap := toRGBA(g[0][0].Fg), toRGBA(g[1][0].Fg)
	assert.Equal(t, uint8(200), top.R)
	assert.Less(t, gap.R, top.R)
	assert.Equal(t, toRGBA(g[2][0].Fg), top)
}

func TestStillRetroEffectsShowTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, screen, Glitch(time.Second).Apply(screen))
	assert.Equal(t, screen, PowerOn(white, time.Second).Apply(screen))
	assert.Empty(t, strings.TrimSpace(ansi.Strip(PowerOff(white, time.Second).Apply(screen))))
}

// A row past the block crops to an empty grid; the band must not read row -1.
func TestPowerOffOnARowPastTheBlockLeavesItAlone(t *testing.T) {
	assert.NotPanics(t, func() { at(PowerOff(white, time.Second).OnRows(9), 0.3).Apply(screen) })
}

func TestDecodeFallsAndSettles(t *testing.T) {
	e := Decode(color.RGBA{0, 255, 0, 255}, 400*ms)
	assert.Equal(t, strings.Repeat(" ", 8), strings.Split(ansi.Strip(e.Apply(screen)), "\n")[0], "nothing has fallen yet")
	mid := at(e, 0.5).Apply(screen)
	sized(t, mid, 8, 5)
	assert.NotEqual(t, screen, ansi.Strip(mid))
	head := false
	for _, row := range cells.Parse(mid) {
		for _, c := range row {
			head = head || c.Attrs&cells.Bold != 0
		}
	}
	assert.True(t, head, "a falling column has a bold head")
	assert.Equal(t, screen, ansi.Strip(landed(t, e).Apply(screen)))
}

func TestFlickerStuttersNowAndThen(t *testing.T) {
	e := Flicker(black, 2400*ms)
	lit, dimmed := 0, 0
	for i := range 48 {
		if fgAt(at(e, float64(i)/48), whiteText) == white {
			lit++
		} else {
			dimmed++
		}
	}
	assert.Greater(t, lit, dimmed, "lit most of the time")
	assert.Positive(t, dimmed, "but not all of it")
	assert.True(t, at(e, 0.5).Active(), "it runs until it is no longer applied")
}

func TestWarpSwaysTheRowsAndLaysThemStraight(t *testing.T) {
	e := Warp(400 * ms)
	moved := false
	for p := 0.05; p < 1; p += 0.05 {
		out := ansi.Strip(at(e, p).Apply(screen))
		sized(t, out, 8, 5)
		moved = moved || out != screen
	}
	assert.True(t, moved)
	assert.Equal(t, screen, ansi.Strip(landed(t, e).Apply(screen)))
}

func TestStillFlickerAndDecodeShowTheBlock(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, white, fgAt(Flicker(black, time.Second), whiteText))
	assert.Equal(t, screen, ansi.Strip(Decode(white, time.Second).Apply(screen)))
}
