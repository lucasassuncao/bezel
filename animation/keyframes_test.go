package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKeyframesPassThroughEveryStop(t *testing.T) {
	moving(t)
	k := NewKeyframes(100*ms,
		Key{At: 1, Value: 0, Ease: Linear},
		Key{At: 0, Value: 0},
		Key{At: 0.5, Value: 10, Ease: Linear},
	)
	start := time.Unix(10, 0)
	k, _ = k.Advance(start)
	require.Equal(t, 0.0, k.cur)
	mid, _ := k.Advance(start.Add(25 * ms))
	require.InDelta(t, 5, mid.cur, 1e-9, "halfway up")
	top, _ := k.Advance(start.Add(50 * ms))
	require.InDelta(t, 10, top.cur, 1e-9)
	down, _ := k.Advance(start.Add(75 * ms))
	require.InDelta(t, 5, down.cur, 1e-9, "on the way back")
	end, done := k.Advance(start.Add(time.Second))
	require.True(t, done)
	require.Equal(t, 0.0, end.cur)
}

func TestKeyframesWithTooFewKeysStayPut(t *testing.T) {
	moving(t)
	one := NewKeyframes(100*ms, Key{Value: 7})
	require.False(t, one.Active())
	require.Equal(t, 7.0, one.cur)
	require.False(t, NewKeyframes(100*ms).Active())
}

func TestStillKeyframesAreBornAtTheirLastKey(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	k := NewKeyframes(100*ms, Key{At: 0, Value: 1}, Key{At: 1, Value: 9})
	require.False(t, k.Active())
	require.Equal(t, 9.0, k.cur)
}

func TestKeyframesTakeOnlyTheirOwnFrames(t *testing.T) {
	moving(t)
	k := NewKeyframes(100*ms, Key{At: 0, Value: 0}, Key{At: 1, Value: 1}).WithDelay(10 * ms).WithYoyo(true).WithRepeat(1)
	_, res := k.Update(NewFloat(0, 1, ms).FrameAt(time.Unix(1, 0)))
	require.Nil(t, res.Cmd)
	_, res = k.Update(k.FrameAt(time.Unix(1, 0)))
	require.NotNil(t, res.Cmd)
}
