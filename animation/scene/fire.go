package scene

import (
	"image/color"
	"math"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/noise"
)

// fireGlyphs are a flame's cells, coolest first.
var fireGlyphs = []string{" ", ".", ":", "*", "#", "▓", "█"}

// NewFire is flames licking up from the bottom, hot in hot and dying out
// in cool towards their tips: a campfire screen, a "this is fine".
func NewFire(hot, cool color.Color) Scene { return newScene(fire, hot, cool) }

// NewAurora is curtains of light in colors waving across the top of the sky.
func NewAurora(colors ...color.Color) Scene { return newScene(aurora, colors...) }

// smooth is noise that changes gently between whole x and y: the hashes at
// the four corners around a point, blended by how near each it is.
func smooth(seed uint64, x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	fx, fy := x-x0, y-y0
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	at := func(dx, dy float64) float64 {
		k := uint64(int64(x0+dx)+1<<20)<<32 | uint64(int64(y0+dy)+1<<20) // #nosec G115 -- a hash key
		return noise.Unit(noise.Hash(seed, k))
	}
	top := at(0, 0) + (at(1, 0)-at(0, 0))*fx
	bottom := at(0, 1) + (at(1, 1)-at(0, 1))*fx
	return top + (bottom-top)*fy
}

// fire draws each cell's heat: highest at the bottom, broken into tongues by
// noise that rises with time.
func (a Scene) fire(cv cells.Canvas, w, h int) {
	hot, cool := a.color(0), a.color(1)
	for y := range h {
		up := float64(h-1-y) / float64(max(h-1, 1)) // 0 at the bottom, 1 at the top
		for x := range w {
			n := smooth(71, float64(x)*0.35, (float64(y)+a.t*9)*0.45)
			heat := min(max(1.15-up*1.6+(n-0.5)*0.9, 0), 1)
			g := fireGlyphs[min(int(heat*float64(len(fireGlyphs))), len(fireGlyphs)-1)]
			fg := hot
			if hot != nil && cool != nil {
				fg = animation.Blend(cool, hot, heat)
			}
			cv.Put(x, y, cells.Cell{Text: g, Width: 1, Fg: fg})
		}
	}
}

// aurora draws a band of light whose middle waves along the columns, its
// colour drifting through colors from left to right.
func (a Scene) aurora(cv cells.Canvas, w, h int) {
	if len(a.colors) == 0 {
		return
	}
	for x := range w {
		fx := float64(x)
		mid := float64(h)*0.35 + math.Sin(fx*0.12+a.t*0.7)*float64(h)*0.15 + math.Sin(fx*0.05-a.t*0.4)*float64(h)*0.1
		thick := float64(h)*0.18 + smooth(81, fx*0.1, a.t*0.3)*float64(h)*0.15
		shade := (math.Sin(fx*0.04+a.t*0.25) + 1) / 2 * float64(len(a.colors)-1)
		i := int(shade)
		c := a.colors[i]
		if i+1 < len(a.colors) {
			c = animation.Blend(a.colors[i], a.colors[i+1], shade-float64(i))
		}
		for y := range h {
			glow := 1 - math.Abs(float64(y)-mid)/thick
			if glow <= 0 {
				continue
			}
			g := "░"
			switch {
			case glow > 0.7:
				g = "▓"
			case glow > 0.35:
				g = "▒"
			}
			cv.Put(x, y, cells.Cell{Text: g, Width: 1, Fg: animation.Blend(dark, c, 0.3+0.7*glow)})
		}
	}
}
