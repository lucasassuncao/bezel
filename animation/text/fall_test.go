package text

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/animation"
)

func TestFallLandsTheWordOnItsLine(t *testing.T) {
	frames := animation.Record(NewFall("drop", 3, 400*ms), Fall.View, time.Second)
	assert.Equal(t, "    \n    \n    \n    ", frames[0], "nothing has let go")
	assert.Equal(t, "    \n    \n    \ndrop", frames[len(frames)-1])
	for _, f := range frames {
		lines := strings.Split(f, "\n")
		assert.Len(t, lines, 4)
		for _, l := range lines {
			assert.Equal(t, 4, ansi.StringWidth(l))
		}
	}
}

func TestFallLetsTheLettersGoAtDifferentTimes(t *testing.T) {
	f := NewFall("abcdefgh", 4, 400*ms)
	starts := map[int]bool{}
	for i := range 8 {
		starts[f.height(i)] = true
	}
	assert.Equal(t, map[int]bool{-1: true}, starts, "none has let go before the first frame")
	start := time.Unix(10, 0)
	f, _ = f.Advance(start)
	f, _ = f.Advance(start.Add(100 * ms))
	heights := map[int]bool{}
	for i := range 8 {
		heights[f.height(i)] = true
	}
	assert.Greater(t, len(heights), 1, "some are falling while others wait or landed")
}

func TestStillFallIsBornLanded(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, "  \nab", NewFall("ab", 1, time.Second).View())
}

func TestFallTakesOnlyItsOwnFrames(t *testing.T) {
	f := NewFall("ab", 2, 100*ms).WithDelay(ms)
	_, cmd := f.Update(NewFall("x", 1, ms).FrameAt(time.Unix(1, 0)))
	assert.Nil(t, cmd)
	_, cmd = f.Update(f.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, cmd)
}
