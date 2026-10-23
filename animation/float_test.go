package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFloatEasesBetweenItsEnds(t *testing.T) {
	start := time.Unix(1000, 0)
	f := NewFloat(0, 1, 100*time.Millisecond).WithEase(Linear)
	assert.True(t, f.Active())

	f, cmd := f.Update(f.FrameAt(start))
	assert.NotNil(t, cmd)
	f, _ = f.Update(f.FrameAt(start.Add(50 * time.Millisecond)))
	assert.InDelta(t, 0.5, f.Cur, 1e-9)
	f, cmd = f.Update(f.FrameAt(start.Add(time.Second)))
	assert.Nil(t, cmd)
	assert.Equal(t, 1.0, f.Cur)
	assert.False(t, f.Active())
}

func TestFloatTakesOnlyItsOwnFrames(t *testing.T) {
	start := time.Unix(1000, 0)
	f, other := NewFloat(0, 1, time.Second), NewFloat(0, 1, time.Second)
	next, cmd := f.Update(other.FrameAt(start))
	assert.Nil(t, cmd)
	assert.Equal(t, f, next)

	twin := f.FrameAt(start)
	f, _ = f.Update(twin)
	_, cmd = f.Update(twin)
	assert.Nil(t, cmd, "a twin frame starts no second clock")
}

func TestNewFloatDeclinesWhenPointless(t *testing.T) {
	f := NewFloat(3, 3, time.Second)
	assert.False(t, f.Active())
	assert.Nil(t, f.Tick())
	f = NewFloat(0, 2.6, 0)
	assert.False(t, f.Active())
	assert.Equal(t, 3, f.Int())
}
