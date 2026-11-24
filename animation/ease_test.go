package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEasesRunFromZeroToOneWithoutTurningBack(t *testing.T) {
	for name, e := range map[string]Ease{"linear": Linear, "in": Quad, "out": Out(Quad), "inout": InOut(Quad)} {
		require.InDelta(t, 0, e(0), 1e-9, name)
		require.InDelta(t, 1, e(1), 1e-9, name)
		prev := 0.0
		for i := 1; i <= 20; i++ {
			v := e(float64(i) / 20)
			require.GreaterOrEqual(t, v, prev, "%s turned back at %d", name, i)
			prev = v
		}
	}
	require.Less(t, Quad(0.25), 0.25, "in starts slow")
	require.Greater(t, Out(Quad)(0.25), 0.25, "out starts fast")
}

func TestATweenFollowsItsEase(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := Tween{from: 0, to: 100, start: start, dur: 100 * time.Millisecond}
	v, _ := tw.At(start.Add(25 * time.Millisecond))
	require.Equal(t, 13, v, "the default is InOut(Quad)")
	v, _ = tw.WithEase(Linear).At(start.Add(25 * time.Millisecond))
	require.Equal(t, 25, v)
}

// curves are every in-curve, by name, for the tests that sweep them all.
var curves = map[string]Ease{
	"sine": Sine, "quad": Quad, "cubic": Cubic, "quart": Quart, "quint": Quint,
	"expo": Expo, "circ": Circ, "back": Back, "elastic": Elastic, "bounce": Bounce,
}

func TestEveryCurveStartsAtZeroAndEndsAtOne(t *testing.T) {
	for name, e := range curves {
		for _, v := range []struct {
			name string
			e    Ease
		}{{"in", e}, {"out", Out(e)}, {"in-out", InOut(e)}} {
			require.InDelta(t, 0, v.e(0), 1e-9, "%s %s", name, v.name)
			require.InDelta(t, 1, v.e(1), 1e-9, "%s %s", name, v.name)
		}
	}
}

func TestTheSmoothCurvesNeverTurnBack(t *testing.T) {
	for _, name := range []string{"sine", "quad", "cubic", "quart", "quint", "expo", "circ"} {
		for _, e := range []Ease{curves[name], Out(curves[name]), InOut(curves[name])} {
			prev := 0.0
			for i := 1; i <= 50; i++ {
				v := e(float64(i) / 50)
				require.GreaterOrEqual(t, v, prev-1e-12, name)
				prev = v
			}
		}
	}
	require.Less(t, Cubic(0.5), Quad(0.5), "a higher power starts slower")
	require.Less(t, Quint(0.5), Quart(0.5))
}

// Back pulls away before it goes, and Out(Back) overshoots before it lands;
// Elastic rings past its end; Bounce never leaves [0,1].
func TestTheLivelyCurvesLeaveTheRange(t *testing.T) {
	low, high := 0.0, 0.0
	for i := 0; i <= 100; i++ {
		p := float64(i) / 100
		low = min(low, Back(p))
		high = max(high, Out(Back)(p), Out(Elastic)(p))
		require.GreaterOrEqual(t, Bounce(p), -1e-9)
		require.LessOrEqual(t, Out(Bounce)(p), 1+1e-9)
	}
	require.Less(t, low, -0.05)
	require.Greater(t, high, 1.05)
}

func TestStepsMovesInWholeSteps(t *testing.T) {
	s := Steps(4)
	require.Equal(t, 0.0, s(0.2))
	require.Equal(t, 0.25, s(0.3))
	require.Equal(t, 0.75, s(0.99))
	require.Equal(t, 1.0, s(1))
	require.Equal(t, 1.0, Steps(0)(0.5), "no steps is a jump straight to the end")
}
