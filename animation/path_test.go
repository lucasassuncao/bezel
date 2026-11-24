package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAPathTravelsAtAnEvenSpeed(t *testing.T) {
	moving(t)
	p := NewPath(100*ms, Point{0, 0}, Point{10, 0}, Point{10, 30}).WithEase(Linear)
	start := time.Unix(10, 0)
	p, _ = p.Advance(start)
	require.Equal(t, Point{0, 0}, p.cur)
	quarter, _ := p.Advance(start.Add(25 * ms))
	require.InDelta(t, 10, quarter.cur.X, 1e-9, "a quarter of 40 cells is the first leg")
	require.InDelta(t, 0, quarter.cur.Y, 1e-9)
	half, _ := p.Advance(start.Add(50 * ms))
	require.InDelta(t, 10, half.cur.Y, 1e-9)
	end, done := p.Advance(start.Add(time.Second))
	require.True(t, done)
	require.Equal(t, Point{10, 30}, end.cur)
}

func TestAPathWithNowhereToGoStays(t *testing.T) {
	moving(t)
	require.False(t, NewPath(100*ms, Point{3, 3}).Active())
	require.False(t, NewPath(100*ms, Point{3, 3}, Point{3, 3}).Active())
	require.False(t, NewPath(100*ms).Active())
	t.Setenv("BEZEL_NO_MOTION", "1")
	still := NewPath(100*ms, Point{0, 0}, Point{5, 5})
	require.False(t, still.Active())
	require.Equal(t, Point{5, 5}, still.cur)
}

func TestBezierRunsFromItsFirstToItsLastPoint(t *testing.T) {
	pts := Bezier(Point{0, 0}, Point{0, 10}, Point{10, 10}, Point{10, 0}, 8)
	require.Len(t, pts, 9)
	require.Equal(t, Point{0, 0}, pts[0])
	require.Equal(t, Point{10, 0}, pts[8])
	require.InDelta(t, 7.5, pts[4].Y, 1e-9, "pulled up towards the handles")
}

func TestAPathTakesOnlyItsOwnFrames(t *testing.T) {
	moving(t)
	p := NewPath(100*ms, Point{0, 0}, Point{4, 0}).WithDelay(ms).WithYoyo(true).WithRepeat(1)
	_, res := p.Update(NewFloat(0, 1, ms).FrameAt(time.Unix(1, 0)))
	require.Nil(t, res.Cmd)
	_, res = p.Update(p.FrameAt(time.Unix(1, 0)))
	require.NotNil(t, res.Cmd)
}
