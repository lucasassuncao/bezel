package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The leader gets ahead and each follower trails the one before it.
func TestFlockTrailsBehindItsLeader(t *testing.T) {
	start := time.Unix(1000, 0)
	f := NewFlock(3, Point{})
	f, cmd := f.To(Point{X: 20})
	require.NotNil(t, cmd)
	f, _ = f.Advance(start)
	f, _ = f.Advance(start.Add(100 * time.Millisecond))
	pts := f.Points()
	require.Greater(t, pts[0].X, pts[1].X)
	require.Greater(t, pts[1].X, pts[2].X)

	f, done := f.Advance(start.Add(10 * time.Second))
	require.True(t, done)
	for _, p := range f.Points() {
		require.Equal(t, Point{X: 20}, p)
	}
}

func TestStillFlockJumps(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	f, cmd := NewFlock(2, Point{}).To(Point{X: 3, Y: 4})
	require.Nil(t, cmd)
	require.Equal(t, []Point{{3, 4}, {3, 4}}, f.Points())
}
