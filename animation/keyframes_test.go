package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, 0.0, k.Cur)
	mid, _ := k.Advance(start.Add(25 * ms))
	assert.InDelta(t, 5, mid.Cur, 1e-9, "halfway up")
	top, _ := k.Advance(start.Add(50 * ms))
	assert.InDelta(t, 10, top.Cur, 1e-9)
	down, _ := k.Advance(start.Add(75 * ms))
	assert.InDelta(t, 5, down.Cur, 1e-9, "on the way back")
	end, done := k.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, 0.0, end.Cur)
}

func TestKeyframesWithTooFewKeysStayPut(t *testing.T) {
	moving(t)
	one := NewKeyframes(100*ms, Key{Value: 7})
	assert.False(t, one.Active())
	assert.Equal(t, 7.0, one.Cur)
	assert.False(t, NewKeyframes(100*ms).Active())
}

func TestStillKeyframesAreBornAtTheirLastKey(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	k := NewKeyframes(100*ms, Key{At: 0, Value: 1}, Key{At: 1, Value: 9})
	assert.False(t, k.Active())
	assert.Equal(t, 9.0, k.Cur)
}

func TestKeyframesTakeOnlyTheirOwnFrames(t *testing.T) {
	moving(t)
	k := NewKeyframes(100*ms, Key{At: 0, Value: 0}, Key{At: 1, Value: 1}).WithDelay(10 * ms).WithYoyo().WithRepeat(1)
	_, cmd := k.Update(NewFloat(0, 1, ms).FrameAt(time.Unix(1, 0)))
	assert.Nil(t, cmd)
	_, cmd = k.Update(k.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, cmd)
}
