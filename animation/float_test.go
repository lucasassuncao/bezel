package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFloatEasesBetweenItsEnds(t *testing.T) {
	start := time.Unix(1000, 0)
	f := NewFloat(0, 1, 100*time.Millisecond).WithEase(Linear)
	require.True(t, f.Active())

	f, res := f.Update(f.FrameAt(start))
	require.NotNil(t, res.Cmd)
	f, _ = f.Update(f.FrameAt(start.Add(50 * time.Millisecond)))
	require.InDelta(t, 0.5, f.cur, 1e-9)
	f, res = f.Update(f.FrameAt(start.Add(time.Second)))
	require.Nil(t, res.Cmd)
	require.Equal(t, 1.0, f.cur)
	require.False(t, f.Active())
}

func TestFloatTakesOnlyItsOwnFrames(t *testing.T) {
	start := time.Unix(1000, 0)
	f, other := NewFloat(0, 1, time.Second), NewFloat(0, 1, time.Second)
	next, res := f.Update(other.FrameAt(start))
	require.Nil(t, res.Cmd)
	require.Equal(t, f, next)

	twin := f.FrameAt(start)
	f, _ = f.Update(twin)
	_, res = f.Update(twin)
	require.Nil(t, res.Cmd, "a twin frame starts no second clock")
}

func TestNewFloatDeclinesWhenPointless(t *testing.T) {
	f := NewFloat(3, 3, time.Second)
	require.False(t, f.Active())
	require.Nil(t, f.Tick())
	f = NewFloat(0, 2.6, 0)
	require.False(t, f.Active())
	require.Equal(t, 3, f.Int())
}

// at drives f by its own frames: the first at start, then one at each offset.
func at(f Float, start time.Time, offsets ...time.Duration) Float {
	f, _ = f.Update(f.FrameAt(start))
	for _, d := range offsets {
		f, _ = f.Update(f.FrameAt(start.Add(d)))
	}
	return f
}

const ms = time.Millisecond

func TestADelayHoldsTheStartValue(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	f := NewFloat(0, 1, 100*ms).WithEase(Linear).WithDelay(100 * ms)
	f = at(f, start, 50*ms)
	require.Equal(t, 0.0, f.cur)
	require.True(t, f.Active(), "waiting is still moving")
	f = at(f, start, 50*ms, 150*ms)
	require.InDelta(t, 0.5, f.cur, 1e-9)
}

func TestRepeatRunsAgainThenLands(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	f := NewFloat(0, 1, 100*ms).WithEase(Linear).WithRepeat(1)
	require.InDelta(t, 0.5, at(f, start, 150*ms).cur, 1e-9, "the second run is half way")
	f = at(f, start, 250*ms)
	require.False(t, f.Active())
	require.Equal(t, 1.0, f.cur)
}

// A yoyo runs back on every other run, so an even count of runs lands on From.
func TestYoyoComesBack(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	f := NewFloat(0, 1, 100*ms).WithEase(Linear).WithYoyo(true).WithRepeat(1)
	require.InDelta(t, 0.75, at(f, start, 25*ms, 125*ms).cur, 1e-9, "on the way back")
	f = at(f, start, 250*ms)
	require.False(t, f.Active())
	require.Equal(t, 0.0, f.cur)
	f, done := f.Advance(start.Add(time.Hour))
	require.True(t, done)
	require.Equal(t, 0.0, f.cur, "a landed animation stays where it landed")
}

func TestForeverNeverLands(t *testing.T) {
	moving(t)
	f := at(NewFloat(0, 1, 100*ms).WithYoyo(true).WithRepeat(Forever), time.Unix(1000, 0), time.Hour)
	require.True(t, f.Active())
	require.NotNil(t, f.Tick())
}
