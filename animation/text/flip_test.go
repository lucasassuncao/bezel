package text

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestFlipTurnsOnlyTheLettersThatChange(t *testing.T) {
	frames := animation.Record(NewFlip("12:59", "13:00", 200*ms), Flip.View, time.Second)
	require.Equal(t, "12:59", frames[0])
	require.Equal(t, "13:00", frames[len(frames)-1])
	turned := false
	for _, f := range frames {
		require.Equal(t, 5, ansi.StringWidth(f), "%q keeps its width", f)
		require.Equal(t, "1", f[:1], "the first digit never moves")
		require.Equal(t, ":", string([]rune(f)[2]), "the colon never moves")
		if f != "12:59" && f != "13:00" {
			turned = true
		}
	}
	require.True(t, turned)
}

func TestFlipStartsFromTheLeft(t *testing.T) {
	f := NewFlip("aaaa", "bbbb", 100*ms)
	start := time.Unix(10, 0)
	f, _ = f.Advance(start)
	f, _ = f.Advance(start.Add(30 * ms))
	got := []rune(f.View())
	require.NotEqual(t, 'a', got[0], "the first card is turning")
	require.Equal(t, 'a', got[3], "the last has not started")
}

func TestFlipPadsTheShorterText(t *testing.T) {
	end := animation.Record(NewFlip("long", "ok", 100*ms), Flip.View, time.Second)
	require.Equal(t, "ok  ", end[len(end)-1])
	f, cmd := NewFlip("a", "b", 100*ms).To("cc")
	require.NotNil(t, cmd)
	require.Equal(t, "b ", f.View())
}

func TestStillFlipShowsTheNewText(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	require.Equal(t, "new", NewFlip("old", "new", time.Second).View())
	require.False(t, NewFlip("same", "same", time.Second).Active())
}

func TestFlipTakesOnlyItsOwnFrames(t *testing.T) {
	f := NewFlip("ab", "cd", 100*ms).WithDelay(ms)
	_, res := f.Update(NewFlip("x", "y", ms).FrameAt(time.Unix(1, 0)))
	require.Nil(t, res.Cmd)
	_, res = f.Update(f.FrameAt(time.Unix(1, 0)))
	require.NotNil(t, res.Cmd)
}
