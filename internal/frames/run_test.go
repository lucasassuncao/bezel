package frames

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const ms = time.Millisecond

func moving(t *testing.T) {
	t.Helper()
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("BEZEL_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
}

func TestARunCountsFromItsFirstFrameToItsEnd(t *testing.T) {
	moving(t)
	start := time.Unix(10, 0)
	r := NewRun(100 * ms)
	r, p, ok, cmd := r.Update(r.FrameAt(start))
	require.True(t, ok)
	require.Equal(t, 0.0, p)
	require.NotNil(t, cmd)
	r, p, _, _ = r.Update(r.FrameAt(start.Add(50 * ms)))
	require.InDelta(t, 0.5, p, 1e-9)
	r, p, _, cmd = r.Update(r.FrameAt(start.Add(200 * ms)))
	require.Equal(t, 1.0, p)
	require.Nil(t, cmd)
	require.False(t, r.Active())
}

func TestARunIgnoresFramesThatAreNotItsOwn(t *testing.T) {
	moving(t)
	r := NewRun(100 * ms)
	_, _, ok, cmd := r.Update(NewMsg(time.Unix(10, 0), r.ID+1, r.Tag))
	require.False(t, ok)
	require.Nil(t, cmd)
}

func TestAStillRunIsBornSpent(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	require.False(t, NewRun(time.Second).Active())
	require.Nil(t, NewRun(time.Second).Tick())
}

func TestPhaseHoldsThroughTheDelayAndGoesBackOnAYoyo(t *testing.T) {
	l := Loop{Delay: 50 * ms, Repeat: 1, Yoyo: true}
	p, done := l.Phase(20*ms, 100*ms)
	require.Equal(t, 0.0, p)
	require.False(t, done)
	p, _ = l.Phase(175*ms, 100*ms)
	require.InDelta(t, 0.75, p, 1e-9, "the second run goes back")
	p, done = l.Phase(time.Second, 100*ms)
	require.Equal(t, 0.0, p, "a yoyo ends where it began")
	require.True(t, done)
}

func TestOrFallsBackToInOut(t *testing.T) {
	require.Equal(t, InOut(0.25), Or(nil)(0.25))
	require.Equal(t, 0.25, Or(func(p float64) float64 { return p })(0.25))
}
