package animation

import "github.com/lucasassuncao/bezel/internal/frames"

// Still reports whether motion is switched off: BEZEL_NO_MOTION set to
// anything, TERM=dumb, or BEZEL_MOTION at 0. Then every animation is born where
// it would end, for a user who finds motion hard to follow, a slow link or a CI log.
func Still() bool { return frames.Still() }

// Speed is how fast motion plays, from BEZEL_MOTION: 2 twice as fast, 0.5 at
// half speed, 1 when it is unset or not a number.
func Speed() float64 { return frames.Speed() }
