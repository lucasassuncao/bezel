package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOrbitGoesRoundItsCentre(t *testing.T) {
	start := time.Unix(1000, 0)
	o := NewOrbit(Point{10, 5}, 6, 3, time.Second)
	o, _ = o.Advance(start)
	require.InDelta(t, 16, o.Value().X, 1e-9)
	quarter, _ := o.Advance(start.Add(250 * time.Millisecond))
	require.InDelta(t, 10, quarter.Value().X, 1e-9)
	require.InDelta(t, 8, quarter.Value().Y, 1e-9)
	require.True(t, quarter.Near(), "below the centre is the near side")
	far, _ := o.Advance(start.Add(750 * time.Millisecond))
	require.False(t, far.Near())
}

func TestOrbitWithRepeatStops(t *testing.T) {
	start := time.Unix(1000, 0)
	o := NewOrbit(Point{}, 1, 1, 100*time.Millisecond).WithRepeat(0)
	o, _ = o.Advance(start)
	_, done := o.Advance(start.Add(time.Second))
	require.True(t, done)
}
