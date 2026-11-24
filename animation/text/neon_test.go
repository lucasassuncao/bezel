package text

import (
	"image/color"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

var pink = color.RGBA{255, 60, 160, 255}

func TestNeonIsMostlyLitAndSometimesFlickers(t *testing.T) {
	n := NewNeon("OPEN", pink, black)
	start := time.Unix(10, 0)
	n, _ = n.Advance(start)
	steady := n.View()
	flickers := 0
	for d := time.Duration(0); d < 10*time.Second; d += neonSlot {
		next, done := n.Advance(start.Add(d))
		require.False(t, done, "a sign burns until it is replaced")
		require.Equal(t, "OPEN", ansi.Strip(next.View()))
		if next.View() != steady {
			flickers++
		}
	}
	require.Greater(t, flickers, 0)
	require.Less(t, flickers, 100, "flickering is the exception")
}

func TestNeonFlickersTheSameWayForTheSameSeed(t *testing.T) {
	start := time.Unix(10, 0)
	a, _ := NewNeon("BAR", pink, black).WithSeed(3).Advance(start)
	b, _ := NewNeon("BAR", pink, black).WithSeed(3).Advance(start)
	for d := time.Duration(0); d < time.Second; d += neonSlot {
		a, _ = a.Advance(start.Add(d))
		b, _ = b.Advance(start.Add(d))
		require.Equal(t, a.View(), b.View())
	}
}

func TestStillNeonNeverFlickers(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	n := NewNeon("ON", pink, black)
	require.False(t, n.Active())
	require.Equal(t, 1.0, n.lit(0))
	require.Equal(t, 1.0, n.lit(1))
}

func TestNeonTakesOnlyItsOwnFrames(t *testing.T) {
	n := NewNeon("ON", pink, black)
	_, res := n.Update(NewNeon("X", pink, black).FrameAt(time.Unix(1, 0)))
	require.Nil(t, res.Cmd)
	_, res = n.Update(n.FrameAt(time.Unix(1, 0)))
	require.NotNil(t, res.Cmd)
}
