package animation

import (
	"math"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Ease maps progress in [0,1] onto how far along the motion is: 0 at the start
// and 1 at the end, and between them a lively one, Back or Elastic, may leave [0,1].
type Ease = frames.Ease

// Linear moves at one speed throughout: a counter, a typewriter.
func Linear(p float64) float64 { return p }

// EaseIn starts slow and ends fast, for something leaving.
func EaseIn(p float64) float64 { return p * p }

// EaseOut starts fast and settles, for something arriving.
func EaseOut(p float64) float64 { return 1 - (1-p)*(1-p) }

// EaseInOut is quadratic, gentler than a cubic on purpose: on whole cells a
// cubic freezes, then skips rows, while this never repeats a height more
// than once over 180ms. Every tween follows it unless told otherwise.
func EaseInOut(p float64) float64 { return frames.InOut(p) }

// Forever is the repeat count of a motion that runs until it is replaced.
const Forever = frames.Forever

// The curves below ease in, slow first; Out and InOut turn any one round.

// Sine eases in along a quarter of a sine wave, the gentlest of them.
func Sine(p float64) float64 { return 1 - math.Cos(p*math.Pi/2) }

// Cubic is p³, slower to start than EaseIn's p².
func Cubic(p float64) float64 { return p * p * p }

// Quart is p⁴.
func Quart(p float64) float64 { return p * p * p * p }

// Quint is p⁵.
func Quint(p float64) float64 { return p * p * p * p * p }

// Expo doubles its pace every tenth of the way: almost still, then gone.
func Expo(p float64) float64 {
	if p <= 0 {
		return 0
	}
	return math.Pow(2, 10*p-10)
}

// Circ eases along a quarter circle.
func Circ(p float64) float64 { return 1 - math.Sqrt(1-p*p) }

// backPull is how far Back pulls away, the common 10% or so.
const backPull = 1.70158

// Back pulls back a little before it goes; Out(Back) overshoots and settles.
func Back(p float64) float64 { return (backPull+1)*p*p*p - backPull*p*p }

// Elastic winds up like a spring before it lets go; Out(Elastic) rings past
// its end and settles.
func Elastic(p float64) float64 {
	if p <= 0 || p >= 1 {
		return math.Round(p)
	}
	return -math.Pow(2, 10*p-10) * math.Sin((10*p-10.75)*2*math.Pi/3)
}

// Bounce hops off the start a few times before it goes; Out(Bounce) drops
// onto its end like a ball, and never leaves [0,1].
func Bounce(p float64) float64 { return 1 - bounceOut(1-p) }

// bounceOut is the ball falling onto its end: four arcs, each lower.
func bounceOut(p float64) float64 {
	const n, d = 7.5625, 2.75
	switch {
	case p < 1/d:
		return n * p * p
	case p < 2/d:
		p -= 1.5 / d
		return n*p*p + 0.75
	case p < 2.5/d:
		p -= 2.25 / d
		return n*p*p + 0.9375
	default:
		p -= 2.625 / d
		return n*p*p + 0.984375
	}
}

// Out is e played backwards from the end: fast first, then settling.
func Out(e Ease) Ease { return func(p float64) float64 { return 1 - e(1-p) } }

// InOut is e easing in over the first half and out over the second.
func InOut(e Ease) Ease {
	return func(p float64) float64 {
		if p < 0.5 {
			return e(2*p) / 2
		}
		return 1 - e(2-2*p)/2
	}
}

// Steps moves in n equal jumps, like a clock's hand; none is one jump to the end.
func Steps(n int) Ease {
	return func(p float64) float64 {
		if n <= 0 || p >= 1 {
			return 1
		}
		return math.Floor(p*float64(n)) / float64(n)
	}
}
