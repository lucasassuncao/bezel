package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEasesRunFromZeroToOneWithoutTurningBack(t *testing.T) {
	for name, e := range map[string]Ease{"linear": Linear, "in": EaseIn, "out": EaseOut, "inout": EaseInOut} {
		assert.InDelta(t, 0, e(0), 1e-9, name)
		assert.InDelta(t, 1, e(1), 1e-9, name)
		prev := 0.0
		for i := 1; i <= 20; i++ {
			v := e(float64(i) / 20)
			assert.GreaterOrEqual(t, v, prev, "%s turned back at %d", name, i)
			prev = v
		}
	}
	assert.Less(t, EaseIn(0.25), 0.25, "in starts slow")
	assert.Greater(t, EaseOut(0.25), 0.25, "out starts fast")
}

func TestATweenFollowsItsEase(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := Tween{From: 0, To: 100, Start: start, Dur: 100 * time.Millisecond}
	v, _ := tw.At(start.Add(25 * time.Millisecond))
	assert.Equal(t, 13, v, "the default is EaseInOut")
	v, _ = tw.WithEase(Linear).At(start.Add(25 * time.Millisecond))
	assert.Equal(t, 25, v)
}

// curves are every in-curve, by name, for the tests that sweep them all.
var curves = map[string]Ease{
	"sine": Sine, "quad": EaseIn, "cubic": Cubic, "quart": Quart, "quint": Quint,
	"expo": Expo, "circ": Circ, "back": Back, "elastic": Elastic, "bounce": Bounce,
}

func TestEveryCurveStartsAtZeroAndEndsAtOne(t *testing.T) {
	for name, e := range curves {
		for _, v := range []struct {
			name string
			e    Ease
		}{{"in", e}, {"out", Out(e)}, {"in-out", InOut(e)}} {
			assert.InDelta(t, 0, v.e(0), 1e-9, "%s %s", name, v.name)
			assert.InDelta(t, 1, v.e(1), 1e-9, "%s %s", name, v.name)
		}
	}
}

func TestTheSmoothCurvesNeverTurnBack(t *testing.T) {
	for _, name := range []string{"sine", "quad", "cubic", "quart", "quint", "expo", "circ"} {
		for _, e := range []Ease{curves[name], Out(curves[name]), InOut(curves[name])} {
			prev := 0.0
			for i := 1; i <= 50; i++ {
				v := e(float64(i) / 50)
				assert.GreaterOrEqual(t, v, prev-1e-12, name)
				prev = v
			}
		}
	}
	assert.Less(t, Cubic(0.5), EaseIn(0.5), "a higher power starts slower")
	assert.Less(t, Quint(0.5), Quart(0.5))
}

// Back pulls away before it goes, and Out(Back) overshoots before it lands;
// Elastic rings past its end; Bounce never leaves [0,1].
func TestTheLivelyCurvesLeaveTheRange(t *testing.T) {
	low, high := 0.0, 0.0
	for i := 0; i <= 100; i++ {
		p := float64(i) / 100
		low = min(low, Back(p))
		high = max(high, Out(Back)(p), Out(Elastic)(p))
		assert.GreaterOrEqual(t, Bounce(p), -1e-9)
		assert.LessOrEqual(t, Out(Bounce)(p), 1+1e-9)
	}
	assert.Less(t, low, -0.05)
	assert.Greater(t, high, 1.05)
}

func TestStepsMovesInWholeSteps(t *testing.T) {
	s := Steps(4)
	assert.Equal(t, 0.0, s(0.2))
	assert.Equal(t, 0.25, s(0.3))
	assert.Equal(t, 0.75, s(0.99))
	assert.Equal(t, 1.0, s(1))
	assert.Equal(t, 1.0, Steps(0)(0.5), "no steps is a jump straight to the end")
}

// A curve that pulls below zero must not draw a non-digit on an odometer.
func TestAnOdometerOnAnOvershootingCurveStaysDigits(t *testing.T) {
	start := time.Unix(1000, 0)
	o := NewOdometer(0, 999, 300*time.Millisecond).WithEase(Back)
	o, _ = o.Update(o.FrameAt(start))
	for d := time.Duration(0); d < 300*time.Millisecond; d += Frame {
		o, _ = o.Update(o.FrameAt(start.Add(d)))
		for _, c := range o.View() {
			assert.True(t, c >= '0' && c <= '9', "%q", o.View())
		}
	}
}
