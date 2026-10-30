package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// A drop falls, overshoots nothing, hops back up off the floor and settles there.
func TestDropFallsBouncesAndSettles(t *testing.T) {
	start := time.Unix(1000, 0)
	d := NewDrop(0, 10)
	d, _ = d.Advance(start)
	d, _ = d.Advance(start.Add(200 * time.Millisecond))
	assert.InDelta(t, 1.2, d.Cur, 0.01, "half of g·t² after 0.2s at 60 cells/s²")

	hop, _ := d.Advance(start.Add(700 * time.Millisecond))
	assert.Less(t, hop.Cur, 10.0, "back up off the floor after the first hit")
	assert.Greater(t, hop.Cur, 0.0)

	d, done := d.Advance(start.Add(5 * time.Second))
	assert.True(t, done)
	assert.Equal(t, 10.0, d.Cur)
	assert.False(t, d.Active())
}

func TestDropFallsUpTooAndCanLandDead(t *testing.T) {
	start := time.Unix(1000, 0)
	d := NewDrop(10, 0).WithBounce(0).WithGravity(200)
	d, _ = d.Advance(start)
	d, done := d.Advance(start.Add(time.Second))
	assert.True(t, done, "with no bounce it lands once")
	assert.Equal(t, 0.0, d.Cur)
}

func TestStillDropStartsLanded(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	d := NewDrop(0, 7)
	assert.False(t, d.Active())
	assert.Equal(t, 7.0, d.Cur)
}

func TestDropTicksOnItsOwnFrames(t *testing.T) {
	d := NewDrop(0, 5)
	_, res := d.Update(d.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, res.Cmd)
}
