package animation

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

var red = "\x1b[38;2;255;0;0mx\x1b[m"

func TestHSLComesBackToTheSameColour(t *testing.T) {
	for _, c := range []color.RGBA{{255, 0, 0, 255}, {12, 200, 90, 255}, {128, 128, 128, 255}, {0, 0, 0, 255}} {
		h, s, l := toHSL(c)
		assert.Equal(t, c, fromHSL(h, s, l))
	}
}

func TestHueShiftTurnsTheColourWheel(t *testing.T) {
	e := HueShift(120, 100*ms).WithEase(Linear)
	assert.Equal(t, color.RGBA{255, 255, 0, 255}, toRGBA(parse(halfway(e).Apply(red))[0][0].fg), "60° on: yellow")
	assert.Equal(t, color.RGBA{0, 255, 0, 255}, toRGBA(parse(landed(t, e).Apply(red))[0][0].fg), "120° on: green")
}

func TestSaturateDrainsOrDeepensColour(t *testing.T) {
	grey := toRGBA(parse(landed(t, Saturate(-1, 100*ms)).Apply(red))[0][0].fg)
	assert.Equal(t, grey.R, grey.G)
	assert.Equal(t, grey.G, grey.B)
	pale := "\x1b[38;2;200;150;150mx\x1b[m"
	deep := toRGBA(parse(landed(t, Saturate(1, 100*ms)).Apply(pale))[0][0].fg)
	assert.Greater(t, int(deep.R)-int(deep.G), 50, "more red than before")
}

func TestColourEffectsLeaveTheTerminalsOwnAlone(t *testing.T) {
	assert.Equal(t, "plain", landed(t, HueShift(90, 100*ms)).Apply("plain"))
	assert.Equal(t, "plain", landed(t, Saturate(-1, 100*ms)).Apply("plain"))
}
