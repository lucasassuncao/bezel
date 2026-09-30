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
	is.False(New(0, 10, 0).Active(), "duration 0 means no animation")
	is.False(New(7, 7, time.Second).Active(), "no distance means no animation")
	is.True(New(0, 10, time.Second).Active(), "a real move animates")
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
