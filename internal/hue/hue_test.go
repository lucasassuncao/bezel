package hue

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHSLComesBackToTheSameColour(t *testing.T) {
	for _, c := range []color.RGBA{{255, 0, 0, 255}, {12, 200, 90, 255}, {128, 128, 128, 255}, {0, 0, 0, 255}} {
		h, s, l := ToHSL(c)
		require.Equal(t, c, FromHSL(h, s, l))
	}
}

func TestShiftTurnsTheWheelEitherWay(t *testing.T) {
	red := color.RGBA{255, 0, 0, 255}
	require.Equal(t, color.RGBA{0, 255, 0, 255}, Shift(red, 120))
	require.Equal(t, color.RGBA{0, 0, 255, 255}, Shift(red, -120))
	require.Nil(t, Shift(nil, 90), "the terminal's own colour has no hue")
}

func TestSaturateStaysWithinRange(t *testing.T) {
	grey := Saturate(color.RGBA{200, 100, 100, 255}, -5).(color.RGBA)
	require.Equal(t, grey.R, grey.G)
	require.Nil(t, Saturate(nil, 2))
}
