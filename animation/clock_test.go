package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAClockRunsUntilStopped(t *testing.T) {
	at := time.Unix(1000, 0)
	c, cmd := NewClock().Start()
	assert.NotNil(t, cmd)
	_, again := c.Start()
	assert.Nil(t, again, "a running clock starts no second chain")

	got, ok := c.Frame(c.FrameAt(at))
	assert.True(t, ok)
	assert.Equal(t, at, got)
	old := c.FrameAt(at)
	c, cmd = c.Next()
	assert.NotNil(t, cmd)
	_, ok = c.Frame(old)
	assert.False(t, ok, "a frame already taken is not taken twice")

	c = c.Stop()
	_, ok = c.Frame(c.FrameAt(at))
	assert.False(t, ok)
	_, cmd = c.Next()
	assert.Nil(t, cmd, "a stopped clock schedules nothing")
}

func TestAClockIgnoresOtherAnimationsFrames(t *testing.T) {
	c, _ := Clock{}.Start()
	_, ok := c.Frame(NewFloat(0, 1, time.Second).FrameAt(time.Unix(1000, 0)))
	assert.False(t, ok)
}

// Frames that come late slow the clock down, to a quarter of the rate at
// most; frames that keep up bring it back.
func TestTakePacesTheClock(t *testing.T) {
	start := time.Unix(1000, 0)
	c, _ := NewClock().Start()
	at := start
	take := func(gap time.Duration) {
		t.Helper()
		at = at.Add(gap)
		var ok bool
		c, _, ok = c.Take(c.FrameAt(at))
		assert.True(t, ok)
		c, _ = c.Next()
	}
	take(0)
	take(Frame)
	assert.Equal(t, Frame, c.gap(), "on time, nothing changes")
	take(100 * time.Millisecond)
	assert.Equal(t, 2*Frame, c.gap())
	take(200 * time.Millisecond)
	take(200 * time.Millisecond)
	assert.Equal(t, 4*Frame, c.gap(), "never slower than a quarter")
	take(4 * Frame)
	take(2 * Frame)
	assert.Equal(t, Frame, c.gap(), "back to full speed once frames keep up")
}

func TestTakeIgnoresOtherFrames(t *testing.T) {
	c, _ := NewClock().Start()
	_, _, ok := c.Take(FrameMsg{})
	assert.False(t, ok)
}
