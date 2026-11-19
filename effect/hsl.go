package effect

import (
	"image/color"
	"time"

	"github.com/lucasassuncao/bezel/animation"
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

// Duotone recolours every coloured cell between dark and light by how bright
// it is, and leaves it so: a pane put aside, an old photograph. Reverse lifts it.
func Duotone(dark, light color.Color, dur time.Duration) Effect {
	e := newEffect(duotone, nil, dur)
	e.base, e.tint = dark, light
	return e
}

// toned is c q of the way to its place between a duotone's dark and light.
func (e Effect) toned(c color.Color, q float64) color.Color {
	if c == nil {
		return nil
	}
	return animation.Blend(c, animation.Blend(e.base, e.tint, luma(c)), q)
}

// luma is how bright c looks, in [0,1], green counting most and blue least.
func luma(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) / 0xffff
}
