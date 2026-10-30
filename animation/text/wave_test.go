package text

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestWaveMovesLettersOffTheMiddleRow(t *testing.T) {
	start := time.Unix(1000, 0)
	w := NewWave("hello world", 1, time.Second)
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

func TestWaveWithRepeatEndsFlat(t *testing.T) {
	start := time.Unix(1000, 0)
	w := NewWave("abc", 1, 100*time.Millisecond).WithRepeat(0)
	w, _ = w.Advance(start)
	w, done := w.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, "abc", strings.Split(w.View(), "\n")[1])
}

func TestWaveTicksOnItsOwnFrames(t *testing.T) {
	w := NewWave("abc", 2, time.Second)
	w, res := w.Update(w.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, res.Cmd)
	_, res = w.Update(animation.FrameMsg{At: time.Unix(2, 0)})
	assert.Nil(t, res.Cmd, "a frame that is not its own is ignored")
}
