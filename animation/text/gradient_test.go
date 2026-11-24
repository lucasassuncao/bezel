package text

import (
	"image/color"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

var grey = color.RGBA{128, 128, 128, 255}

func TestGradientSpreadsItsColoursRoundTheText(t *testing.T) {
	g := NewGradient("abcd", []color.Color{black, white}, 100*ms)
	require.Equal(t, black, g.colorAt(0))
	require.Equal(t, grey, g.colorAt(1), "half way between the stops")
	require.Equal(t, white, g.colorAt(2))
	require.Equal(t, grey, g.colorAt(3), "and round back towards the first")
	require.Equal(t, "abcd", ansi.Strip(g.View()))
}

func TestGradientFlowsAlong(t *testing.T) {
	start := time.Unix(1000, 0)
	g := NewGradient("abcd", []color.Color{black, white}, 100*ms)
	g, _ = g.Update(g.FrameAt(start))
	g, _ = g.Update(g.FrameAt(start.Add(25 * ms)))
	require.Equal(t, grey, g.colorAt(0), "a quarter turn moves each letter a quarter round")
	g, _ = g.Update(g.FrameAt(start.Add(time.Hour)))
	require.True(t, g.Active(), "a gradient flows until it is replaced")
}

func TestAGradientWithoutColoursIsPlain(t *testing.T) {
	require.Equal(t, "abc", NewGradient("abc", nil, time.Second).View())
}
