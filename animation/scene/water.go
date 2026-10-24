package scene

import (
	"image/color"
	"math"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/noise"
)

// bubbleGlyphs grow as a bubble rises, and the last is it popping.
var bubbleGlyphs = []string{".", "o", "O", "*"}

// NewBubbles is bubbles in c rising from the bottom, swaying and growing,
// and popping at the top.
func NewBubbles(c color.Color) Scene { return newScene(bubbles, c) }

// NewSea is waves rolling along the bottom of the screen, foam in foam on
// their crests and the water in water below.
func NewSea(foam, water color.Color) Scene { return newScene(sea, foam, water) }

// bubbles draws about one bubble for every forty cells, each on its own column,
// speed and sway.
func (a Scene) bubbles(cv cells.Canvas, w, h int) {
	c := a.color(0)
	for k := range uint64(max(w*h/40, 1)) {
		speed := 2 + 4*noise.Unit(noise.Hash(91, k))
		span := float64(h + 1)
		rise := math.Mod(a.t*speed+noise.Unit(noise.Hash(92, k))*span, span) // rows risen
		y := h - 1 - int(rise)
		x := int(noise.Unit(noise.Hash(93, k))*float64(w)) + int(math.Round(math.Sin(a.t*2+float64(k))*0.8))
		g := bubbleGlyphs[min(int(rise/span*float64(len(bubbleGlyphs)-1)), len(bubbleGlyphs)-2)]
		if y <= 0 {
			y, g = 0, bubbleGlyphs[len(bubbleGlyphs)-1]
		}
		cv.Put(x, y, cells.Cell{Text: g, Width: 1, Fg: c})
	}
}

// sea draws the water up to a surface made of two waves rolling at their own
// speeds, the cell at the surface foam and the water deeper the lower it is.
func (a Scene) sea(cv cells.Canvas, w, h int) {
	foam, water := a.color(0), a.color(1)
	base := float64(h) * 0.55
	for x := range w {
		fx := float64(x)
		surface := base + math.Sin(fx*0.25-a.t*2)*1.2 + math.Sin(fx*0.09+a.t*1.1)*1.5
		top := int(math.Round(surface))
		for y := max(top, 0); y < h; y++ {
			c := cells.Cell{Text: " ", Width: 1, Bg: water}
			if water != nil {
				depth := float64(y-top) / float64(max(h-top, 1))
				c.Bg = animation.Blend(water, dark, depth*0.6)
			}
			if y == top {
				c.Text, c.Fg = "~", foam
			}
			cv.Put(x, y, c)
		}
	}
}
