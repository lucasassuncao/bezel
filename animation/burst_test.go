package animation

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestABurstThrowsSparksOutOfItsPoint(t *testing.T) {
	moving(t)
	b := NewBurst(Point{20, 10}, 12, 600*ms).WithGravity(0)
	start := time.Unix(10, 0)
	b, _ = b.Advance(start)
	first := b.Sparks()
	require.Len(t, first, 12)
	for _, s := range first {
		assert.Equal(t, Point{20, 10}, s.At, "every spark starts at the point")
	}
	later, _ := b.Advance(start.Add(300 * ms))
	far := 0.0
	for _, s := range later.Sparks() {
		far = max(far, math.Hypot(s.At.X-20, s.At.Y-10))
		assert.InDelta(t, 0.5, s.Life, 1e-9)
	}
	assert.Greater(t, far, 2.0)
}

func TestGravityPullsTheSparksDown(t *testing.T) {
	moving(t)
	start := time.Unix(10, 0)
	mean := func(g float64) float64 {
		b := NewBurst(Point{0, 0}, 40, time.Second).WithGravity(g).WithSpeed(10)
		b, _ = b.Advance(start)
		b, _ = b.Advance(start.Add(500 * ms))
		sum := 0.0
		for _, s := range b.Sparks() {
			sum += s.At.Y
		}
		return sum / 40
	}
	assert.Greater(t, mean(40), mean(0)+1)
}

func TestASpentBurstHasNoSparks(t *testing.T) {
	moving(t)
	b := NewBurst(Point{0, 0}, 5, 100*ms).WithSeed(3).WithDelay(0)
	start := time.Unix(10, 0)
	b, _ = b.Advance(start)
	b, done := b.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Empty(t, b.Sparks())

	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Empty(t, NewBurst(Point{0, 0}, 5, time.Second).Sparks())
}

func TestABurstTakesOnlyItsOwnFrames(t *testing.T) {
	moving(t)
	b := NewBurst(Point{0, 0}, 3, 100*ms)
	_, res := b.Update(NewFloat(0, 1, ms).FrameAt(time.Unix(1, 0)))
	assert.Nil(t, res.Cmd)
	_, res = b.Update(b.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, res.Cmd)
}
