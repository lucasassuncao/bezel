package animation

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// moving clears both switches, so a test reads motion the way a normal terminal has it.
func moving(t *testing.T) {
	t.Helper()
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("BEZEL_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
}

func TestStillIsReadFromTheEnvironment(t *testing.T) {
	moving(t)
	assert.False(t, Still())
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.True(t, Still())
	moving(t)
	t.Setenv("TERM", "dumb")
	assert.True(t, Still(), "a dumb terminal redraws too poorly to animate")
}

// Still, every animation is born where it would have ended, asking for no frame.
func TestStillAnimationsAreBornAtTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")

	tw := New(0, 10, time.Second)
	assert.False(t, tw.Active())
	assert.Equal(t, 10, tw.Cur)
	assert.Nil(t, tw.Tick())

	f := NewFloat(0, 1, time.Second)
	assert.False(t, f.Active())
	assert.Equal(t, 1.0, f.Cur)

	assert.Equal(t, white, NewFade(black, white, time.Second).Color())
	s, cmd := NewSpring(0).To(10)
	assert.Nil(t, cmd)
	assert.Equal(t, 10.0, s.Pos, "a still spring jumps")
}

// TestMain clears both switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("BEZEL_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

func TestSpeedIsReadFromTheEnvironment(t *testing.T) {
	moving(t)
	assert.InDelta(t, 1.0, Speed(), 0)
	t.Setenv("BEZEL_MOTION", "2")
	assert.InDelta(t, 2.0, Speed(), 0)
	t.Setenv("BEZEL_MOTION", "fast")
	assert.InDelta(t, 1.0, Speed(), 0, "a value that is not a number plays at normal speed")
	t.Setenv("BEZEL_MOTION", "0")
	assert.True(t, Still(), "speed 0 is motion off")
}

// At speed 2 a tween is over in half its time; a spring lives its time faster too.
func TestSpeedScalesEveryClock(t *testing.T) {
	moving(t)
	t.Setenv("BEZEL_MOTION", "2")
	start := time.Unix(100, 0)
	tw := New(0, 10, time.Second)
	tw, _ = tw.Advance(start)
	tw, done := tw.Advance(start.Add(500 * time.Millisecond))
	assert.True(t, done)
	assert.Equal(t, 10, tw.Cur)

	assert.Equal(t, time.Second, frames.Paced(500*time.Millisecond))
}
