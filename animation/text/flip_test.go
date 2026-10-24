package text

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/animation"
)

func TestFlipTurnsOnlyTheLettersThatChange(t *testing.T) {
	frames := animation.Record(NewFlip("12:59", "13:00", 200*ms), Flip.View, time.Second)
	assert.Equal(t, "12:59", frames[0])
	assert.Equal(t, "13:00", frames[len(frames)-1])
	turned := false
	for _, f := range frames {
		assert.Equal(t, 5, ansi.StringWidth(f), "%q keeps its width", f)
		assert.Equal(t, "1", f[:1], "the first digit never moves")
		assert.Equal(t, ":", string([]rune(f)[2]), "the colon never moves")
		if f != "12:59" && f != "13:00" {
			turned = true
		}
	}
	assert.True(t, turned)
}

func TestFlipStartsFromTheLeft(t *testing.T) {
	f := NewFlip("aaaa", "bbbb", 100*ms)
	start := time.Unix(10, 0)
	f, _ = f.Advance(start)
	f, _ = f.Advance(start.Add(30 * ms))
	got := []rune(f.View())
	assert.NotEqual(t, 'a', got[0], "the first card is turning")
	assert.Equal(t, 'a', got[3], "the last has not started")
}

func TestFlipPadsTheShorterText(t *testing.T) {
	end := animation.Record(NewFlip("long", "ok", 100*ms), Flip.View, time.Second)
	assert.Equal(t, "ok  ", end[len(end)-1])
	f, cmd := NewFlip("a", "b", 100*ms).To("cc")
	assert.NotNil(t, cmd)
	assert.Equal(t, "b ", f.View())
}

func TestStillFlipShowsTheNewText(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, "new", NewFlip("old", "new", time.Second).View())
	assert.False(t, NewFlip("same", "same", time.Second).Active())
}

func TestFlipTakesOnlyItsOwnFrames(t *testing.T) {
	f := NewFlip("ab", "cd", 100*ms).WithDelay(ms)
	_, cmd := f.Update(NewFlip("x", "y", ms).FrameAt(time.Unix(1, 0)))
	assert.Nil(t, cmd)
	_, cmd = f.Update(f.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, cmd)
}
