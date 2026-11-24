package effect

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/internal/cells"
)

func TestIrisInOpensFromTheMiddle(t *testing.T) {
	e := IrisIn(400 * ms)
	early := strings.Split(ansi.Strip(at(e, 0.3).Apply(screen)), "\n")
	require.Equal(t, "        ", early[0], "the corners are still closed")
	require.NotEqual(t, "        ", early[2], "the middle is open")
	require.Equal(t, screen, landed(t, e).Apply(screen))
	sized(t, at(e, 0.3).Apply(screen), 8, 5)
}

func TestIrisOutClosesToNothing(t *testing.T) {
	require.Zero(t, litRows(landed(t, IrisOut(300*ms)).Apply(screen)))
	require.Positive(t, litRows(at(IrisOut(300*ms), 0.5).Apply(screen)))
}

func TestPixelateIsCoarsestHalfwayAndSharpAtTheEnd(t *testing.T) {
	coloured := "\x1b[38;2;255;0;0mab\x1b[38;2;0;0;255mcd\nefgh\x1b[m"
	mid := cells.Parse(at(Pixelate(400*ms), 0.5).Apply(coloured))
	require.Equal(t, "█", mid[0][2].Text)
	require.Equal(t, toRGBA(mid[0][0].Fg), toRGBA(mid[0][3].Fg), "one square, one colour")
	require.Equal(t, toRGBA(mid[0][0].Fg), toRGBA(mid[1][1].Fg))
	require.Equal(t, "abcd\nefgh", ansi.Strip(landed(t, Pixelate(400*ms)).Apply(coloured)))
	sized(t, at(Pixelate(400*ms), 0.5).Apply(coloured), 4, 2)
}

func TestStillShapesShowTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	require.Equal(t, screen, IrisIn(time.Second).Apply(screen))
	require.Equal(t, screen, Pixelate(time.Second).Apply(screen))
}
