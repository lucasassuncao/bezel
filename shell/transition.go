package shell

import (
	"time"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/effect"
)

// SlideTabs is a Transition that slides the new tab's panes in from the side
// its tab sits on: from the right for a later tab, from the left for an earlier.
// Experimental, as the effect package is.
func SlideTabs(dur time.Duration) func(from, to int) effect.Effect {
	return func(from, to int) effect.Effect {
		if to > from {
			return effect.SlideIn(effect.Left, dur).WithEase(animation.Out(animation.Cubic))
		}
		return effect.SlideIn(effect.Right, dur).WithEase(animation.Out(animation.Cubic))
	}
}

// Transition is the effect playing over the panes now, for a test or a host
// that delivers its frames itself; it is inactive when none plays.
// Experimental, as the effect package is.
func (s Shell) Transition() effect.Effect { return s.trans }
