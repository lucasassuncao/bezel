package effect

import (
	"time"
)

// HueShift turns every coloured cell round the colour wheel by degrees: with
// WithYoyo, a block that warms and cools; Forever, one that cycles.
func HueShift(degrees float64, dur time.Duration) Effect {
	e := newEffect(hueShift, nil, dur)
	e.amount = degrees
	return e
}

// Saturate scales every coloured cell's saturation by 1+amount: -1 drains it
// to grey, 1 doubles it. The terminal's own colours have none and are left.
func Saturate(amount float64, dur time.Duration) Effect {
	e := newEffect(saturate, nil, dur)
	e.amount = amount
	return e
}
