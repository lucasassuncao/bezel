package text

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestBurnFlaresThenLeavesAshThenNothing(t *testing.T) {
	frames := animation.Record(NewBurn("secret", 300*ms), Burn.View, time.Second)
	require.Equal(t, "secret", frames[0])
	require.Equal(t, "      ", frames[len(frames)-1])
	ashes := false
	for _, f := range frames {
		require.Equal(t, 6, ansi.StringWidth(f), "%q keeps its width", f)
		if strings.Contains(ansi.Strip(f), "·") {
			ashes = true
		}
	}
	require.True(t, ashes)
}

func TestBurnCatchesFromTheLeft(t *testing.T) {
	b := NewBurn("abcdefghij", 300*ms)
	require.Less(t, b.catches(0), b.catches(9))
	start := time.Unix(10, 0)
	b, _ = b.Advance(start)
	b, _ = b.Advance(start.Add(60 * ms))
	got := ansi.Strip(b.View())
	require.Equal(t, "j", got[len(got)-1:], "the last letter is untouched yet")
}

func TestStillBurnIsBornGone(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	require.Equal(t, "   ", NewBurn("abc", time.Second).View())
}

func TestBurnTakesOnlyItsOwnFrames(t *testing.T) {
	b := NewBurn("ab", 100*ms).WithDelay(ms)
	_, res := b.Update(NewBurn("x", ms).FrameAt(time.Unix(1, 0)))
	require.Nil(t, res.Cmd)
	_, res = b.Update(b.FrameAt(time.Unix(1, 0)))
	require.NotNil(t, res.Cmd)
}
