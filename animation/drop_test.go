package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A drop falls, overshoots nothing, hops back up off the floor and settles there.
func TestDropFallsBouncesAndSettles(t *testing.T) {
	start := time.Unix(1000, 0)
	d := NewDrop(0, 10)
	d, _ = d.Advance(start)
	d, _ = d.Advance(start.Add(200 * time.Millisecond))
	require.InDelta(t, 1.2, d.cur, 0.01, "half of g·t² after 0.2s at 60 cells/s²")

	hop, _ := d.Advance(start.Add(700 * time.Millisecond))
	require.Less(t, hop.cur, 10.0, "back up off the floor after the first hit")
	require.Greater(t, hop.cur, 0.0)

	d, done := d.Advance(start.Add(5 * time.Second))
	require.True(t, done)
	require.Equal(t, 10.0, d.cur)
	require.False(t, d.Active())
}

func TestDropFallsUpTooAndCanLandDead(t *testing.T) {
	start := time.Unix(1000, 0)
	d := NewDrop(10, 0).WithBounce(0).WithGravity(200)
	d, _ = d.Advance(start)
	d, done := d.Advance(start.Add(time.Second))
	require.True(t, done, "with no bounce it lands once")
	require.Equal(t, 0.0, d.cur)
}

func TestStillDropStartsLanded(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	d := NewDrop(0, 7)
	require.False(t, d.Active())
	require.Equal(t, 7.0, d.cur)
}

func TestDropTicksOnItsOwnFrames(t *testing.T) {
	d := NewDrop(0, 5)
	_, res := d.Update(d.FrameAt(time.Unix(1, 0)))
	require.NotNil(t, res.Cmd)
}
