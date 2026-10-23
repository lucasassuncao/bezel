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
