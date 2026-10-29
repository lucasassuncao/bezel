package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTweenAtEndpointsAndMidpoint(t *testing.T) {
	is := assert.New(t)

	start := time.Now()
	tw := Tween{From: 0, To: 12, Start: start, Dur: 100 * time.Millisecond}

	v, done := tw.At(start)
	is.Equal(0, v, "at t=0 the tween sits on its start value")
	is.False(done, "a tween at t=0 is not finished")

	// The curve is symmetric about 0.5, so the midpoint is the midpoint.
	v, done = tw.At(start.Add(50 * time.Millisecond))
	is.Equal(6, v, "at t=0.5 the eased value is half the distance")
	is.False(done, "a tween at t=0.5 is not finished")

	v, done = tw.At(start.Add(100 * time.Millisecond))
	is.Equal(12, v, "a finished tween lands exactly on its end value")
	is.True(done, "at t=1 the tween is finished")

	v, done = tw.At(start.Add(time.Second))
	is.Equal(12, v, "overshooting time does not overshoot the value")
	is.True(done, "past t=1 the tween is finished")
}

// TestTweenEasesRatherThanInterpolatesLinearly is what separates this from a
// plain lerp: the eased curve starts slow, so the first quarter of the time
// covers noticeably less than a quarter of the distance.
func TestTweenEasesRatherThanInterpolatesLinearly(t *testing.T) {
	is := assert.New(t)
	start := time.Now()
	tw := Tween{From: 0, To: 100, Start: start, Dur: 100 * time.Millisecond}

	v, _ := tw.At(start.Add(25 * time.Millisecond))
	is.Less(v, 25, "ease-in must lag a linear interpolation early on")

	v, _ = tw.At(start.Add(75 * time.Millisecond))
	is.Greater(v, 75, "ease-out must lead a linear interpolation late on")
}

func TestZeroTweenIsInactive(t *testing.T) {
	is := assert.New(t)
	var tw Tween
	is.False(tw.Active(), "the zero tween is inactive")
	_, done := tw.At(time.Now())
	is.True(done, "the zero tween reports as finished")
}

func TestNewDeclinesWhenPointless(t *testing.T) {
	is := assert.New(t)
	is.False(NewTween(0, 10, 0).Active(), "duration 0 means no animation")
	is.False(NewTween(7, 7, time.Second).Active(), "no distance means no animation")
	is.True(NewTween(0, 10, time.Second).Active(), "a real move animates")
	// An app that draws Cur after every Update must not see a declined tween at 0.
	is.Equal(10, NewTween(0, 10, 0).Cur, "a declined tween rests at its target")
	is.Equal(7, NewTween(7, 7, time.Second).Cur)
}

// TestCurveDoesNotStallOrLurch is the smoothness contract: heights are whole
// rows, so stutter is a height frozen for frames, then a multi-row lurch.
// Spans cover the hint panel from a 24-row terminal (floored at 5) to 60 rows.
func TestCurveDoesNotStallOrLurch(t *testing.T) {
	is := assert.New(t)

	const dur = 180 * time.Millisecond
	frames := int(dur / Frame)

	for _, span := range []struct {
		name     string
		from, to int
		maxJump  int
	}{
		{"short panel, opening", 0, 5, 2},
		{"typical panel, closing", 10, 0, 3},
		{"tall panel, opening", 0, 17, 4},
	} {
		start := time.Now()
		tw := Tween{From: span.from, To: span.to, Start: start, Dur: dur}

		// Fewer rows than frames must repeat heights: quantisation, not stutter.
		// The bound is an even distribution - no worse than the arithmetic forces.
		maxStall := max(frames/abs(span.to-span.from), 1)

		prev, _ := tw.At(start)
		stall, worstStall, worstJump := 0, 0, 0
		for i := 1; i <= frames; i++ {
			v, _ := tw.At(start.Add(time.Duration(i) * Frame))
			if v == prev {
				stall++
				worstStall = max(worstStall, stall)
			} else {
				stall = 0
				if jump := abs(v - prev); jump > worstJump {
					worstJump = jump
				}
			}
			prev = v
		}

		is.LessOrEqual(worstStall, maxStall, "%s: the height must not freeze beyond what quantisation forces", span.name)
		is.LessOrEqual(worstJump, span.maxJump, "%s: the height must not lurch", span.name)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Tween is a value: Advance hands back the sampled tween and leaves its own
// receiver alone, like every other method on it.
func TestAdvanceReturnsTheSampledTween(t *testing.T) {
	start := time.Now()
	tw := Tween{From: 0, To: 10, Cur: 0, Start: start, Dur: 100 * time.Millisecond}

	next, done := tw.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, 10, next.Cur)
	assert.Equal(t, 0, tw.Cur, "the receiver changed")
}

// New reads no clock: time reaches a tween only through the frames it gets.
func TestNewLeavesTheClockToTheFirstFrame(t *testing.T) {
	tw := NewTween(0, 10, 100*time.Millisecond)
	assert.True(t, tw.Start.IsZero())

	at := time.Unix(1000, 0)
	tw, done := tw.Advance(at)
	assert.False(t, done)
	assert.Equal(t, 0, tw.Cur, "the first frame shows the start value")
	assert.Equal(t, at, tw.Start)
}

func TestUpdateFollowsItsOwnFramesUntilItLands(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := NewTween(0, 10, 100*time.Millisecond)

	tw, cmd := tw.Update(tw.FrameAt(start))
	assert.NotNil(t, cmd, "an unfinished tween asks for the next frame")
	tw, cmd = tw.Update(tw.FrameAt(start.Add(50 * time.Millisecond)))
	assert.Equal(t, 5, tw.Cur)
	assert.NotNil(t, cmd)
	tw, cmd = tw.Update(tw.FrameAt(start.Add(100 * time.Millisecond)))
	assert.Equal(t, 10, tw.Cur)
	assert.Nil(t, cmd, "a tween that landed stops its clock")
}

// A tween replaced mid-flight still has a frame on its way: the new one must not take it.
func TestUpdateIgnoresAnotherTweensFrames(t *testing.T) {
	start := time.Unix(1000, 0)
	old, tw := NewTween(0, 10, time.Second), NewTween(0, 10, time.Second)

	next, cmd := tw.Update(old.FrameAt(start))
	assert.Nil(t, cmd)
	assert.Equal(t, tw, next)

	var zero Tween
	assert.Nil(t, zero.Tick(), "nothing to animate, no clock")
	next, cmd = zero.Update(old.FrameAt(start))
	assert.Nil(t, cmd)
	assert.Equal(t, zero, next)
}

func TestTickDeliversAFrameForThisTween(t *testing.T) {
	tw := NewTween(0, 10, time.Second)
	next, cmd := tw.Update(tw.Tick()())
	assert.False(t, next.Start.IsZero(), "the tick's frame started the clock")
	assert.NotNil(t, cmd)
}

// Landing ends the tween: Active turns false and no further frame is asked.
func TestALandedTweenIsInactiveAtItsEnd(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := NewTween(0, 10, 100*time.Millisecond)
	tw, _ = tw.Update(tw.FrameAt(start))
	tw, _ = tw.Update(tw.FrameAt(start.Add(time.Second)))
	assert.False(t, tw.Active())
	assert.Nil(t, tw.Tick())
	assert.Equal(t, 10, tw.Cur)

	next, _ := NewTween(0, 10, 100*time.Millisecond).Advance(start)
	next, done := next.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.False(t, next.Active(), "Advance lands it too")
}

// Tick called twice sends two frames for one step; the tween takes one.
func TestASecondClockIsDropped(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := NewTween(0, 10, time.Second)
	first, second := tw.FrameAt(start), tw.FrameAt(start)
	tw, cmd := tw.Update(first)
	assert.NotNil(t, cmd)
	next, cmd := tw.Update(second)
	assert.Nil(t, cmd, "the twin frame starts no second clock")
	assert.Equal(t, tw, next)
}
