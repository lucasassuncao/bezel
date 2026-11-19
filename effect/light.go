package effect

import (
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/hue"
)

// pressDark is the colour a pressed block sinks towards.
var pressDark color.Color = color.RGBA{0, 0, 0, 255}

func newLight(kind effectKind, c color.Color, strength float64, dur time.Duration) Effect {
	e := newEffect(kind, nil, dur)
	e.tint, e.amount = c, strength
	return e
}

// Tint washes the block towards c and back again, once: red on an error, green
// on a save.
func Tint(c color.Color, dur time.Duration) Effect { return newLight(tint, c, 0.6, dur) }

// Pulse washes the block towards c and back, over and over: the panel with
// the focus, the one with an error. WithRepeat gives it an end.
func Pulse(c color.Color, dur time.Duration) Effect {
	e := newLight(tint, c, 0.5, dur)
	e.run.Repeat = animation.Forever
	return e
}

// Press darkens the block and lets it come back, as a key pressed down.
func Press(dur time.Duration) Effect { return newLight(tint, pressDark, 0.45, dur) }

// Flash lights the block in c at once and lets the light fade out: a value
// that changed, a line that arrived.
func Flash(c color.Color, dur time.Duration) Effect { return newLight(flash, c, 0.8, dur) }

// Invert swaps the block's text and background for a moment, an alarm.
func Invert(dur time.Duration) Effect { return newEffect(invert, nil, dur) }

// Spotlight dims everything but the w×h part whose top left is x, y, and
// leaves it dimmed: onboarding, "look here". Reverse lifts it.
func Spotlight(x, y, w, h int, dim color.Color, dur time.Duration) Effect {
	e := newLight(spotlight, dim, 0.7, dur)
	e.spot = rect{x: x, y: y, w: max(w, 1), h: max(h, 1)}
	return e
}

// Vignette darkens the block towards dim the farther a cell is from its middle.
func Vignette(dim color.Color, dur time.Duration) Effect { return newLight(vignette, dim, 0.8, dur) }

// Contrast scales every colour's distance from mid grey by factor: above 1
// sharper, below 1 washed out, a pane leaving the focus. It stays.
func Contrast(factor float64, dur time.Duration) Effect { return newLight(contrast, nil, factor, dur) }

// Hover brings the block's background to c, from base where a cell has none,
// and holds it there: the item under the cursor. Reverse lets go.
func Hover(base, c color.Color, dur time.Duration) Effect {
	e := newLight(hover, c, 0, dur)
	e.base = base
	return e
}

// Melt lets the cells drip down out of the block one by one, tinted towards
// c as they fall when c is not nil: a line that leaves, a screen closing.
func Melt(c color.Color, dur time.Duration) Effect {
	e := newEffect(melt, Noise(1), dur)
	e.tint = c
	return e
}

// Ripple sends a ring of c out across the block from the cell at x, y, fading
// as it widens: where a click landed, a ClickMsg's X and Y.
func Ripple(x, y int, c color.Color, dur time.Duration) Effect {
	e := newLight(ripple, c, 0.8, dur)
	e.spot = rect{x: x, y: y, w: 1, h: 1}
	e.run.Ease = animation.Out(animation.Quad)
	return e
}

// Shine passes a slanted band of c across the block once, the colours either
// side of it kept: a button that just became ready, an item that is new.
func Shine(c color.Color, dur time.Duration) Effect { return newLight(shine, c, 0.75, dur) }

// WithStrength sets how far a light goes, in [0,1]: the share of c a tint,
// pulse, flash, spotlight, vignette, ripple, shine or flicker reaches at its height.
func (e Effect) WithStrength(k float64) Effect {
	e.amount = min(max(k, 0), 1)
	return e
}

// rippleBand is the width of a ripple's ring, as a share of the farthest a
// cell can be from where it starts.
const rippleBand = 0.18

// shineBand is how many columns either side of its middle a shine's band lights.
const shineBand = 4.0

// rippled is c lit by how near it sits to the ring, which fades as it widens.
func (e Effect) rippled(c cells.Cell, q float64, x, y, w, h int) cells.Cell {
	if q <= 0 || q >= 1 {
		return c
	}
	ring := q * (1 + rippleBand)
	lit := max(0, 1-math.Abs(distFrom(e.spot.x, e.spot.y, x, y, w, h)-ring)/rippleBand)
	return tinted(c, e.tint, 1-e.amount*lit*(1-0.6*q))
}

// shone is c lit by how near it sits to the band, slanted like / and
// travelling from the top left corner to the bottom right.
func (e Effect) shone(c cells.Cell, q float64, x, y, w, h int) cells.Cell {
	if q <= 0 || q >= 1 {
		return c
	}
	far := float64(w-1) + 2*float64(h-1)
	mid := q*(far+2*shineBand) - shineBand
	lit := max(0, 1-math.Abs(float64(x)+2*float64(y)-mid)/shineBand)
	return tinted(c, e.tint, 1-e.amount*lit)
}

// turnAt is c, at x, y of a w×h block, turned q of the way: the lights that
// depend on where a cell sits, then every other kind.
func (e Effect) turnAt(c cells.Cell, q float64, x, y, w, h int) cells.Cell {
	switch e.kind {
	case scanlines:
		return e.scanline(c, q, y)
	case skeleton:
		return e.skeleton(c, q, x, w)
	case trace:
		return e.traced(c, q, x, y, w, h)
	case marching:
		return e.marched(c, x, y, w, h)
	case ripple:
		return e.rippled(c, q, x, y, w, h)
	case shine:
		return e.shone(c, q, x, y, w, h)
	case decode:
		return e.decoded(c, q, x, y, w)
	case spotlight:
		if e.spot.holds(x, y) {
			return c
		}
		return tinted(c, e.tint, 1-e.amount*q)
	case vignette:
		dx, dy, mx, my := centre(x, y, w, h)
		d := 0.0
		if far := math.Hypot(mx, my); far > 0 {
			d = math.Hypot(dx, dy) / far
		}
		return tinted(c, e.tint, 1-e.amount*q*d*d)
	}
	return e.light(c, q)
}

// light is c turned q of the way by the kinds that do not care where it sits.
func (e Effect) light(c cells.Cell, q float64) cells.Cell {
	switch e.kind {
	case tint:
		return tinted(c, e.tint, 1-e.amount*math.Sin(math.Pi*q))
	case flash:
		return tinted(c, e.tint, 1-e.amount*(1-q))
	case flicker:
		return tinted(c, e.tint, 1-e.amount*e.stutter())
	case invert:
		if q > 0.15 && q < 0.85 {
			c.Attrs ^= cells.Reverse
		}
		return c
	case contrast:
		f := 1 + (e.amount-1)*q
		c.Fg, c.Bg = scaleContrast(c.Fg, f), scaleContrast(c.Bg, f)
		return c
	case hover:
		bg := c.Bg
		if bg == nil {
			bg = e.base
		}
		if bg != nil {
			c.Bg = animation.Blend(bg, e.tint, q)
		}
		return c
	}
	return e.turn(c, q)
}

// scaleContrast is c with its lightness k times as far from the middle.
func scaleContrast(c color.Color, k float64) color.Color {
	if c == nil {
		return nil
	}
	h, s, l := hue.ToHSL(c)
	return hue.FromHSL(h, s, min(max(0.5+(l-0.5)*k, 0), 1))
}

// melt is g with each cell dropped by how far it has turned, faster as it
// goes; a cell that falls out of the block is gone. Lower rows go down first,
// so a falling cell lands on top of what is below.
func (e Effect) melt(g cells.Grid, w, h int, p float64) cells.Grid {
	cv := cells.NewCanvas(w, h)
	for y := len(g) - 1; y >= 0; y-- {
		x := 0
		for _, c := range g[y] {
			q := e.reach(p, x, y, w, h)
			if q <= 0 || (q < 1 && strings.TrimSpace(c.Text) != "") {
				if e.tint != nil {
					c = tinted(c, e.tint, 1-q)
				}
				cv.Put(x, y+int(math.Round(q*q*float64(h+1))), c)
			}
			x += c.Width
		}
	}
	return cv.Grid()
}
