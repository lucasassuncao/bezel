package text

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestShimmerSweepsAGlowAcrossTheText(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewShimmer("abcdefghij", black, white, 100*ms).WithBand(2)
	require.True(t, s.Active(), "a shimmer runs until it is replaced")
	require.Equal(t, black, s.colorAt(0), "before the first frame the glow is off the text")

	s, _ = s.Update(s.FrameAt(start))
	s, _ = s.Update(s.FrameAt(start.Add(50 * ms))) // the band's centre on the sixth letter
	require.Equal(t, white, s.colorAt(5))
	require.Equal(t, black, s.colorAt(0))
	require.Equal(t, "abcdefghij", ansi.Strip(s.View()))

	s, _ = s.Update(s.FrameAt(start.Add(time.Hour)))
	require.True(t, s.Active())
}

func TestAShimmerThatEndsGoesDark(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewShimmer("abc", black, white, 100*ms).WithRepeat(0)
	s, _ = s.Update(s.FrameAt(start))
	s, res := s.Update(s.FrameAt(start.Add(time.Second)))
	require.Nil(t, res.Cmd)
	for i := range 3 {
		require.Equal(t, black, s.colorAt(i))
	}
}
