// Package hue turns colours round the HSL wheel, for the animations that
// tint, cycle or drain a block's colours.
package hue

import (
	"image/color"
	"math"
)

// Shift is c turned degrees round the wheel; nil stays nil.
func Shift(c color.Color, degrees float64) color.Color {
	if c == nil || degrees == 0 {
		return c
	}
	h, s, l := ToHSL(c)
	return FromHSL(math.Mod(h+degrees+360*math.Ceil(math.Abs(degrees)/360), 360), s, l)
}

// Saturate is c with its saturation times k, kept within [0,1].
func Saturate(c color.Color, k float64) color.Color {
	if c == nil || k == 1 {
		return c
	}
	h, s, l := ToHSL(c)
	return FromHSL(h, min(max(s*k, 0), 1), l)
}

// ToHSL is c as hue in degrees, saturation and lightness in [0,1].
func ToHSL(c color.Color) (h, s, l float64) {
	r16, g16, b16, _ := c.RGBA()
	r, g, b := float64(r16)/0xffff, float64(g16)/0xffff, float64(b16)/0xffff
	hi, lo := max(r, g, b), min(r, g, b)
	l = (hi + lo) / 2
	d := hi - lo
	if d == 0 {
		return 0, 0, l
	}
	s = d / (1 - math.Abs(2*l-1))
	switch hi {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return h, s, l
}

// FromHSL is the colour of hue h in degrees, saturation s and lightness l.
func FromHSL(h, s, l float64) color.RGBA {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	byte8 := func(v float64) uint8 { return uint8(math.Round(min(max(v+m, 0), 1) * 255)) }
	return color.RGBA{byte8(r), byte8(g), byte8(b), 255}
}
