package animation

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwellMovesLettersOffTheMiddleRow(t *testing.T) {
	start := time.Unix(1000, 0)
	w := NewSwell("hello world", 1, time.Second)
	rows := strings.Split(w.View(), "\n")
	require.Len(t, rows, 3)
	assert.Equal(t, "hello world", rows[1], "before the first frame the text is flat")

	w, _ = w.Advance(start)
	w, _ = w.Advance(start.Add(250 * time.Millisecond))
	rows = strings.Split(w.View(), "\n")
	assert.NotEqual(t, "hello world", rows[1])
	for _, r := range rows {
		assert.Equal(t, ansi.StringWidth("hello world"), ansi.StringWidth(r), "every row is as wide as the text")
	}
}

func TestSwellWithRepeatEndsFlat(t *testing.T) {
	start := time.Unix(1000, 0)
	w := NewSwell("abc", 1, 100*time.Millisecond).WithRepeat(0)
	w, _ = w.Advance(start)
	w, done := w.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, "abc", strings.Split(w.View(), "\n")[1])
}

func TestSwellTicksOnItsOwnFrames(t *testing.T) {
	w := NewSwell("abc", 2, time.Second)
	w, cmd := w.Update(w.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, cmd)
	_, cmd = w.Update(FrameMsg{At: time.Unix(2, 0)})
	assert.Nil(t, cmd, "a frame that is not its own is ignored")
}
