package frames

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStillAndSpeedAreReadFromTheEnvironment(t *testing.T) {
	moving(t)
	require.False(t, Still())
	require.InDelta(t, 1.0, Speed(), 0)
	t.Setenv("BEZEL_MOTION", "0.5")
	require.InDelta(t, 0.5, Speed(), 0)
	t.Setenv("TERM", "dumb")
	require.True(t, Still())
}

func TestPacedScalesWallTimeBySpeed(t *testing.T) {
	moving(t)
	t.Setenv("BEZEL_MOTION", "2")
	require.Equal(t, 2*time.Second, Paced(time.Second))
	t.Setenv("BEZEL_MOTION", "0")
	require.Equal(t, time.Second, Paced(time.Second), "speed 0 is Still, not frozen time")
}
