package effect

import (
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
)

// skeletonBand is how many cells either side of its centre a skeleton's glow reaches.
const skeletonBand = 6

// traceTail is how many cells of the border a trace's light trails behind it.
const traceTail = 8

// Skeleton draws the block as grey bars where its text would be, with a glow
// sweeping across them once every dur: the shape of what is still loading.
// It sweeps until it is no longer applied; WithRepeat gives it an end.
func Skeleton(base, glow color.Color, dur time.Duration) Effect {
	e := newEffect(skeleton, nil, dur)
	e.tint, e.base = glow, base
	e.run.Ease, e.run.Repeat = animation.Linear, animation.Forever
	if !e.Active() {
		e.p = 0 // Still: the bars, without the glow moving
	}
	return e
}

// Trace runs a light round the block's outer edge once every dur, a short
// tail fading behind it: a panel at work, quieter than a spinner. It runs
// until it is no longer applied; WithRepeat gives it an end.
func Trace(c color.Color, dur time.Duration) Effect {
	e := newEffect(trace, nil, dur)
	e.tint = c
	e.run.Ease, e.run.Repeat = animation.Linear, animation.Forever
	return e
}

// Marching runs dashes of c round the block's outer edge, a dash and a gap
// further on every dur: a selection still held. It runs until it is no longer
// applied; WithRepeat gives it an end. Still, the dashes stand where they start.
func Marching(c color.Color, dur time.Duration) Effect {
	e := newEffect(marching, nil, dur)
	e.tint = c
	e.run.Ease, e.run.Repeat = animation.Linear, animation.Forever
	return e
}

// marchDash is how many cells long a marching edge's dashes are, and its gaps.
const marchDash = 2

// marched is c in the dashes' colour when it sits on the edge under a dash.
func (e Effect) marched(c cells.Cell, x, y, w, h int) cells.Cell {
	idx, ok := edgeIndex(x, y, w, h)
	if !ok {
		return c
	}
	period := 2 * marchDash
	step := int(e.p * float64(period))
	if ((idx-step)%period+period)%period < marchDash {
		c.Fg = e.tint
	}
	return c
}

// skeleton is c as a bar, lit by how near the glow it sits at x.
func (e Effect) skeleton(c cells.Cell, q float64, x, w int) cells.Cell {
	if strings.TrimSpace(c.Text) == "" {
		return blank(c)
	}
	centre := q*float64(w+2*skeletonBand) - skeletonBand
	glow := max(0, 1-math.Abs(float64(x)-centre)/skeletonBand)
	if !e.Active() {
		glow = 0
	}
	return cells.Cell{Text: strings.Repeat("▒", c.Width), Width: c.Width, Fg: animation.Blend(e.base, e.tint, glow)}
}

// edgeIndex is how far round the edge of a w×h block x, y sits, clockwise
// from the top left, and false inside the block.
func edgeIndex(x, y, w, h int) (int, bool) {
	switch {
	case y == 0:
		return x, true
	case x == w-1:
		return w - 1 + y, true
	case y == h-1:
		return (w - 1) + (h - 1) + (w - 1 - x), true
	case x == 0:
		return 2*(w-1) + (h - 1) + (h - 1 - y), true
	}
	return 0, false
}

// traced is c lit by the light at q of the way round, when it sits on the edge.
func (e Effect) traced(c cells.Cell, q float64, x, y, w, h int) cells.Cell {
	idx, ok := edgeIndex(x, y, w, h)
	round := 2*(w-1) + 2*(h-1)
	if !ok || round <= 0 || !e.Active() {
		return c
	}
	head := int(q * float64(round))
	behind := ((head-idx)%round + round) % round
	if behind >= traceTail {
		return c
	}
	c.Fg = animation.Blend(c.Fg, e.tint, 1-float64(behind)/traceTail)
	return c
}
