package frames

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStillAndSpeedAreReadFromTheEnvironment(t *testing.T) {
	moving(t)
	assert.False(t, Still())
	assert.InDelta(t, 1.0, Speed(), 0)
	t.Setenv("BEZEL_MOTION", "0.5")
	assert.InDelta(t, 0.5, Speed(), 0)
	t.Setenv("TERM", "dumb")
	assert.True(t, Still())
}

func TestPacedScalesWallTimeBySpeed(t *testing.T) {
	moving(t)
	t.Setenv("BEZEL_MOTION", "2")
	assert.Equal(t, 2*time.Second, Paced(time.Second))
	t.Setenv("BEZEL_MOTION", "0")
	assert.Equal(t, time.Second, Paced(time.Second), "speed 0 is Still, not frozen time")
}
