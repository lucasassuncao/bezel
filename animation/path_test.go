package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAPathTravelsAtAnEvenSpeed(t *testing.T) {
	moving(t)
	p := NewPath(100*ms, Point{0, 0}, Point{10, 0}, Point{10, 30}).WithEase(Linear)
	start := time.Unix(10, 0)
	p, _ = p.Advance(start)
	assert.Equal(t, Point{0, 0}, p.Cur)
	quarter, _ := p.Advance(start.Add(25 * ms))
	assert.InDelta(t, 10, quarter.Cur.X, 1e-9, "a quarter of 40 cells is the first leg")
	assert.InDelta(t, 0, quarter.Cur.Y, 1e-9)
	half, _ := p.Advance(start.Add(50 * ms))
	assert.InDelta(t, 10, half.Cur.Y, 1e-9)
	end, done := p.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, Point{10, 30}, end.Cur)
}

func TestAPathWithNowhereToGoStays(t *testing.T) {
	moving(t)
	assert.False(t, NewPath(100*ms, Point{3, 3}).Active())
	assert.False(t, NewPath(100*ms, Point{3, 3}, Point{3, 3}).Active())
	assert.False(t, NewPath(100*ms).Active())
	t.Setenv("BEZEL_NO_MOTION", "1")
	still := NewPath(100*ms, Point{0, 0}, Point{5, 5})
	assert.False(t, still.Active())
	assert.Equal(t, Point{5, 5}, still.Cur)
}

func TestBezierRunsFromItsFirstToItsLastPoint(t *testing.T) {
	pts := Bezier(Point{0, 0}, Point{0, 10}, Point{10, 10}, Point{10, 0}, 8)
	assert.Len(t, pts, 9)
	assert.Equal(t, Point{0, 0}, pts[0])
	assert.Equal(t, Point{10, 0}, pts[8])
	assert.InDelta(t, 7.5, pts[4].Y, 1e-9, "pulled up towards the handles")
}

func TestAPathTakesOnlyItsOwnFrames(t *testing.T) {
	moving(t)
	p := NewPath(100*ms, Point{0, 0}, Point{4, 0}).WithDelay(ms).WithYoyo().WithRepeat(1)
	_, cmd := p.Update(NewFloat(0, 1, ms).FrameAt(time.Unix(1, 0)))
	assert.Nil(t, cmd)
	_, cmd = p.Update(p.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, cmd)
}
