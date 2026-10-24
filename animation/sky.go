package animation

import (
	"image/color"
	"math"
)

// staticGlyphs are the grains of NewStatic, lightest first.
var staticGlyphs = []string{" ", "░", "▒", "▓"}

// starLayers are the starfield's depths: how fast each drifts, in columns a
// second, and what its stars look like.
var starLayers = []struct {
	speed float64
	glyph string
}{{0.5, "·"}, {1.5, "+"}, {3, "*"}}

// NewStarfield is stars in c that twinkle and drift left, the nearer the faster.
func NewStarfield(c color.Color) Ambient { return newAmbient(starfield, c) }

// NewStatic is the snow of a television with no signal, in c.
func NewStatic(c color.Color) Ambient { return newAmbient(static, c) }

// NewPlasma is a field of colour between a and b that ripples without end,
// drawn as background: a splash screen, an about page.
func NewPlasma(a, b color.Color) Ambient { return newAmbient(plasma, a, b) }

// dark is the colour stars and static fade towards.
var dark color.Color = color.RGBA{0, 0, 0, 255}

// stars draws each layer: a star at a column of the drifting sky when that
// column's hash says so, its brightness rising and falling on its own beat.
func (a Ambient) stars(cv canvas, w, h int) {
	c := a.color(0)
	for li, layer := range starLayers {
		drift := int(a.t * layer.speed)
		for y := range h {
			for x := range w {
				k := uint64(y)<<32 | uint64(x+drift)<<4 | uint64(li) // #nosec G115 -- a hash key
				if unit(noise(51, k)) > 0.012 {
					continue
				}
				glow := 0.5 + 0.5*math.Sin(a.t*(1+2*unit(noise(52, k)))+unit(noise(53, k))*2*math.Pi)
				fg := c
				if c != nil {
					fg = Blend(dark, c, 0.3+0.7*glow)
				}
				cv.put(x, y, cell{text: layer.glyph, width: 1, fg: fg})
			}
		}
	}
}

// static draws a fresh grain in every cell, two dozen times a second.
func (a Ambient) static(cv canvas, w, h int) {
	frame := uint64(a.t*24) << 32 // #nosec G115 -- a frame count
	c := a.color(0)
	for y := range h {
		for x := range w {
			k := frame | uint64(y*w+x) // #nosec G115 -- a hash key
			fg := c
			if c != nil {
				fg = Blend(dark, c, unit(noise(62, k)))
			}
			cv.put(x, y, cell{text: staticGlyphs[noise(61, k)%uint64(len(staticGlyphs))], width: 1, fg: fg})
		}
	}
}

// plasma colours every cell's background by the classic sum of sines, rows
// counted double so the ripples are round.
func (a Ambient) plasma(cv canvas, w, h int) {
	from, to := a.color(0), a.color(1)
	for y := range h {
		for x := range w {
			fx, fy := float64(x), 2*float64(y)
			v := math.Sin(fx/8+a.t) + math.Sin(fy/6+a.t*1.3) +
				math.Sin((fx+fy)/12+a.t*0.7) + math.Sin(math.Hypot(fx-float64(w)/2, fy-float64(h))/8-a.t)
			bg := from
			if from != nil && to != nil {
				bg = Blend(from, to, (v+4)/8)
			}
			cv.put(x, y, cell{text: " ", width: 1, bg: bg})
		}
	}
}
