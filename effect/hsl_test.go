package effect

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
)

var red = "\x1b[38;2;255;0;0mx\x1b[m"

func TestHueShiftTurnsTheColourWheel(t *testing.T) {
	e := HueShift(120, 100*ms).WithEase(animation.Linear)
	assert.Equal(t, color.RGBA{255, 255, 0, 255}, toRGBA(cells.Parse(halfway(e).Apply(red))[0][0].Fg), "60° on: yellow")
	assert.Equal(t, color.RGBA{0, 255, 0, 255}, toRGBA(cells.Parse(landed(t, e).Apply(red))[0][0].Fg), "120° on: green")
}

func TestSaturateDrainsOrDeepensColour(t *testing.T) {
	grey := toRGBA(cells.Parse(landed(t, Saturate(-1, 100*ms)).Apply(red))[0][0].Fg)
	assert.Equal(t, grey.R, grey.G)
	assert.Equal(t, grey.G, grey.B)
	pale := "\x1b[38;2;200;150;150mx\x1b[m"
	deep := toRGBA(cells.Parse(landed(t, Saturate(1, 100*ms)).Apply(pale))[0][0].Fg)
	assert.Greater(t, int(deep.R)-int(deep.G), 50, "more red than before")
}

func TestColourEffectsLeaveTheTerminalsOwnAlone(t *testing.T) {
	assert.Equal(t, "plain", landed(t, HueShift(90, 100*ms)).Apply("plain"))
	assert.Equal(t, "plain", landed(t, Saturate(-1, 100*ms)).Apply("plain"))
}
