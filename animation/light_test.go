package animation

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// fgAt is the foreground of the first cell of e's frame of block.
func fgAt(e Effect, block string) any { return toRGBA(parse(e.Apply(block))[0][0].fg) }

// A tint rises to its colour and comes back: untouched at both ends.
func TestTintGoesAndComesBack(t *testing.T) {
	e := Tint(black, 100*ms).WithEase(Linear).WithStrength(1)
	assert.Equal(t, white, fgAt(e, whiteText), "nothing before the first frame")
	assert.Equal(t, black, fgAt(halfway(e), whiteText), "at its height, all tint")
	assert.Equal(t, white, fgAt(landed(t, e), whiteText))
}

func TestPulseRepeatsUntilReplaced(t *testing.T) {
	e := Pulse(black, 100*ms)
	start := time.Unix(1000, 0)
	e, _ = e.Update(e.FrameAt(start))
	e, cmd := e.Update(e.FrameAt(start.Add(time.Hour)))
	assert.NotNil(t, cmd, "a pulse keeps going")
	assert.True(t, e.Active())
}

func TestFlashStartsLitAndFades(t *testing.T) {
	e := Flash(black, 100*ms).WithStrength(1)
	assert.Equal(t, black, fgAt(e, whiteText))
	assert.Equal(t, white, fgAt(landed(t, e), whiteText))
}

func TestInvertSwapsForAMoment(t *testing.T) {
	e := Invert(100 * ms).WithEase(Linear)
	assert.NotZero(t, parse(halfway(e).Apply("ab"))[0][0].attrs&reverse)
	assert.Zero(t, parse(landed(t, e).Apply("ab"))[0][0].attrs&reverse)
}

func TestSpotlightDimsAllButItsPart(t *testing.T) {
	block := "\x1b[38;2;255;255;255mabcd\x1b[m"
	g := parse(landed(t, Spotlight(1, 0, 2, 1, black, 100*ms).WithStrength(1)).Apply(block))[0]
	assert.Equal(t, black, toRGBA(g[0].fg))
	assert.Equal(t, white, toRGBA(g[1].fg), "the lit part keeps its colour")
	assert.Equal(t, white, toRGBA(g[2].fg))
	assert.Equal(t, black, toRGBA(g[3].fg))
}

func TestVignetteLeavesTheMiddleAlone(t *testing.T) {
	block := "\x1b[38;2;255;255;255mabcde\x1b[m"
	g := parse(landed(t, Vignette(black, 100*ms).WithStrength(1)).Apply(block))[0]
	assert.Equal(t, white, toRGBA(g[2].fg))
	assert.Equal(t, black, toRGBA(g[0].fg), "the edge goes all the way")
}

func TestContrastWashesOut(t *testing.T) {
	g := parse(landed(t, Contrast(0, 100*ms)).Apply(whiteText))[0]
	assert.Equal(t, grey, toRGBA(g[0].fg), "no contrast at all is mid grey")
}

func TestHoverHoldsAndReverseLetsGo(t *testing.T) {
	e := Hover(black, white, 100*ms)
	assert.Equal(t, white, toRGBA(parse(landed(t, e).Apply("ab"))[0][0].bg), "a cell with no background takes base, then c")
	back := landed(t, e.Reverse()).Apply("ab")
	assert.Equal(t, black, toRGBA(parse(back)[0][0].bg))
}

func TestPressDarkensAndReturns(t *testing.T) {
	e := Press(100 * ms).WithEase(Linear)
	assert.NotEqual(t, white, fgAt(halfway(e), whiteText))
	assert.Equal(t, white, fgAt(landed(t, e), whiteText))
}

// Melted cells fall out of the block; its size stays.
func TestMeltDripsTheBlockAway(t *testing.T) {
	block := "abcd\nefgh"
	out := landed(t, Melt(nil, 100*ms)).Apply(block)
	assert.Equal(t, "    \n    ", ansi.Strip(out))
	mid := halfway(Melt(black, 100*ms).WithPattern(Sweep(Right)).WithEase(Linear)).Apply(block)
	assert.Equal(t, 2, len(parse(mid)))
}

func TestStillLightsAreBornAtTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, white, fgAt(Tint(black, time.Second), whiteText))
	assert.Equal(t, white, fgAt(Flash(black, time.Second), whiteText))
	assert.False(t, Pulse(black, time.Second).Active())
}
