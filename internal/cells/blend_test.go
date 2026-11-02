package cells

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	grey  = color.RGBA{128, 128, 128, 255}
)

func TestMixGoesPartWay(t *testing.T) {
	require.Equal(t, grey, Mix(black, white, 0.5))
	require.Equal(t, black, Mix(black, white, -1), "f is kept in [0,1]")
	require.Equal(t, white, Mix(black, white, 2))
	require.Equal(t, white, Mix(nil, white, 0.3), "a missing colour gives way")
	require.Equal(t, black, Mix(black, nil, 0.3))
}

func TestTowardBlendsEveryColourButTheTerminalsBackground(t *testing.T) {
	g := Grid{{
		{Text: "a", Width: 1, Fg: white, Bg: white, Attrs: Bold},
		{Text: "b", Width: 1},
	}}
	out := g.Toward(black, white, 0.5)
	require.Equal(t, grey, out[0][0].Fg)
	require.Equal(t, grey, out[0][0].Bg)
	require.Equal(t, Bold, out[0][0].Attrs, "attributes stay")
	require.Equal(t, grey, out[0][1].Fg, "the terminal's foreground blends from ink")
	require.Nil(t, out[0][1].Bg, "the terminal's background is already there")
	require.Equal(t, white, g[0][0].Fg, "g is untouched")
}
