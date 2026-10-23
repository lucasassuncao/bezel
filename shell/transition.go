package shell

import (
	"time"

	"github.com/lucasassuncao/bezel/animation"
)

// SlideTabs is a Transition that slides the new tab's panes in from the side
// its tab sits on: from the right for a later tab, from the left for an earlier.
func SlideTabs(dur time.Duration) func(from, to int) animation.Effect {
	return func(from, to int) animation.Effect {
		if to > from {
			return animation.SlideIn(animation.Left, dur).WithEase(animation.Out(animation.Cubic))
		}
		return animation.SlideIn(animation.Right, dur).WithEase(animation.Out(animation.Cubic))
	}
}

// Transition is the effect playing over the panes now, for a test or a host
// that delivers its frames itself; it is inactive when none plays.
func (s Shell) Transition() animation.Effect { return s.trans }
