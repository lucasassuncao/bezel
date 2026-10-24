package effect

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/cells"
)

func TestIrisInOpensFromTheMiddle(t *testing.T) {
	e := IrisIn(400 * ms)
	early := strings.Split(ansi.Strip(at(e, 0.3).Apply(screen)), "\n")
	assert.Equal(t, "        ", early[0], "the corners are still closed")
	assert.NotEqual(t, "        ", early[2], "the middle is open")
	assert.Equal(t, screen, landed(t, e).Apply(screen))
	sized(t, at(e, 0.3).Apply(screen), 8, 5)
}

func TestIrisOutClosesToNothing(t *testing.T) {
	assert.Zero(t, litRows(landed(t, IrisOut(300*ms)).Apply(screen)))
	assert.Positive(t, litRows(at(IrisOut(300*ms), 0.5).Apply(screen)))
}

func TestPixelateIsCoarsestHalfwayAndSharpAtTheEnd(t *testing.T) {
	coloured := "\x1b[38;2;255;0;0mab\x1b[38;2;0;0;255mcd\nefgh\x1b[m"
	mid := cells.Parse(at(Pixelate(400*ms), 0.5).Apply(coloured))
	assert.Equal(t, "█", mid[0][2].Text)
	assert.Equal(t, toRGBA(mid[0][0].Fg), toRGBA(mid[0][3].Fg), "one square, one colour")
	assert.Equal(t, toRGBA(mid[0][0].Fg), toRGBA(mid[1][1].Fg))
	assert.Equal(t, "abcd\nefgh", ansi.Strip(landed(t, Pixelate(400*ms)).Apply(coloured)))
	sized(t, at(Pixelate(400*ms), 0.5).Apply(coloured), 4, 2)
}

func TestStillShapesShowTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, screen, IrisIn(time.Second).Apply(screen))
	assert.Equal(t, screen, Pixelate(time.Second).Apply(screen))
}
