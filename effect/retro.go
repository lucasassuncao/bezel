package effect

import (
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/noise"
)

// glitchSlots is how many times a glitch jumps to a new shape over its run.
const glitchSlots = 14

// The two colours a glitched row's text splits towards, as a screen's
// channels slip apart.
var (
	glitchRed  color.Color = color.RGBA{255, 40, 80, 255}
	glitchCyan color.Color = color.RGBA{40, 220, 255, 255}
)

// Glitch tears the block for a moment: rows jump sideways and their colours
// split, worst halfway, and the block comes back whole. An error, a hack.
func Glitch(dur time.Duration) Effect {
	e := newEffect(glitch, nil, dur)
	e.run.Ease = animation.Linear
	return e
}

// PowerOff turns the block off like an old tube: it squashes to a glowing
// line, the line to a dot, and the dot goes out. glow is the line's light.
func PowerOff(glow color.Color, dur time.Duration) Effect {
	e := newEffect(powerOff, nil, dur)
	e.tint = glow
	e.run.Ease = animation.Linear
	return e
}

// PowerOn is PowerOff played backwards: a dot, a line, then the block.
func PowerOn(glow color.Color, dur time.Duration) Effect { return PowerOff(glow, dur).Reverse() }

// Scanlines darkens every other row towards dim, as the lines of an old
// screen, and leaves them darkened. Reverse lifts them.
func Scanlines(dim color.Color, dur time.Duration) Effect { return newLight(scanlines, dim, 0.35, dur) }

// Decode drops columns of glyphs in c down the block, each column from a
// moment of its own, and leaves the block's own text settled behind them.
func Decode(c color.Color, dur time.Duration) Effect {
	e := newEffect(decode, Columns(1), dur)
	e.tint = c
	e.run.Ease = animation.Linear
	return e
}

// Flicker dims the block towards dim now and then, a stutter at a time, as a
// failing tube: a link that dropped, a service unwell. It runs until it is no
// longer applied; WithRepeat gives it an end.
func Flicker(dim color.Color, dur time.Duration) Effect {
	e := newLight(flicker, dim, 0.8, dur)
	e.run.Ease, e.run.Repeat = animation.Linear, animation.Forever
	return e
}

// Warp sways the rows sideways along a travelling wave, worst halfway, and
// lays them straight again: heat over a road, a memory.
func Warp(dur time.Duration) Effect {
	e := newEffect(warp, nil, dur)
	e.run.Ease = animation.Linear
	return e
}

// decodeGlyphs are what a decode's columns show as they fall, and decodeSlots
// how many times over its run a falling glyph changes.
const (
	decodeGlyphs = "0123456789ABCDEF<>/\\|=+*#%&@$?"
	decodeSlots  = 30
)

// decoded is c blank until its column reaches it, then a glyph in the rain's
// colour that settles into c; the newest glyph of a column, its head, is bold.
func (e Effect) decoded(c cells.Cell, q float64, x, y, w int) cells.Cell {
	switch {
	case q <= 0:
		return blank(c)
	case q >= 1:
		return c
	}
	k := uint64(max(y*w+x, 0))*decodeSlots + uint64(e.p*decodeSlots) // #nosec G115 -- clamped to zero or more
	glyph := string(decodeGlyphs[noise.Hash(33, k)%uint64(len(decodeGlyphs))])
	own := c.Fg
	if own == nil {
		own = e.tint
	}
	out := cells.Cell{Text: glyph + strings.Repeat(" ", c.Width-1), Width: c.Width, Fg: animation.Blend(e.tint, own, q*q), Bg: c.Bg}
	if q < 0.2 {
		out.Attrs = cells.Bold
	}
	return out
}

// flickerSlots is how many times over one run a flicker rolls for a stutter.
const flickerSlots = 24

// stutter is how far a flicker has dimmed the block now, 0 for fully lit.
func (e Effect) stutter() float64 {
	if !e.Active() {
		return 0
	}
	slot := uint64(e.p * flickerSlots)
	if noise.Unit(noise.Hash(41, slot)) > 0.18 {
		return 0
	}
	return 0.45 + 0.55*noise.Unit(noise.Hash(42, slot))
}

// warpReach is how many columns a warped row sways at most, warpRows how many
// rows one wave spans, and warpTurns how many waves pass over a run.
const (
	warpReach = 2.0
	warpRows  = 6.0
	warpTurns = 2.0
)

// warp is g with each row moved sideways along the wave at p, which swells
// to halfway and dies away, the block's size kept.
func (e Effect) warp(g cells.Grid, w int, p float64) cells.Grid {
	force := math.Sin(math.Pi * p)
	for y, row := range g {
		g[y] = shift(row, round(warpReach*force*math.Sin(2*math.Pi*(float64(y)/warpRows+p*warpTurns))), w)
	}
	return g
}

// glitch is g torn at raw progress raw.
func (e Effect) glitch(g cells.Grid, w int, raw float64) cells.Grid {
	if raw <= 0 || raw >= 1 {
		return g
	}
	force := math.Sin(math.Pi * raw)
	slot := uint64(raw * glitchSlots)
	out := make(cells.Grid, len(g))
	for y, row := range g {
		k := slot*1000 + uint64(y)
		if noise.Unit(noise.Hash(21, k)) > 0.45*force {
			out[y] = row
			continue
		}
		dx := 1 + int(noise.Hash(22, k)%3)
		split := glitchRed
		if noise.Hash(23, k)%2 == 0 {
			dx, split = -dx, glitchCyan
		}
		torn := shift(row, dx, w)
		for i, c := range torn {
			if strings.TrimSpace(c.Text) != "" {
				torn[i].Fg = animation.Blend(c.Fg, split, 0.6*force)
			}
		}
		out[y] = torn
	}
	return out
}

// powerOff is g on its way out at p: squashed rows first, then a line
// narrowing to a dot, then nothing.
func (e Effect) powerOff(g cells.Grid, w, h int, p float64) cells.Grid {
	const squash, narrow = 0.55, 0.9
	cv := cells.NewCanvas(w, h)
	mid := (h - 1) / 2
	line := func(width int, glyph string) {
		lo := (w - width) / 2
		for x := lo; x < lo+width; x++ {
			cv.Put(x, mid, cells.Cell{Text: glyph, Width: 1, Fg: e.tint})
		}
	}
	switch {
	case p <= 0:
		return g
	case p < squash:
		q := p / squash
		band := max(int(math.Round(float64(h)*(1-q))), 1)
		top := mid - (band-1)/2
		for y := range band {
			src := g[min(int(float64(y)/float64(band)*float64(h)), h-1)]
			x := 0
			for _, c := range src {
				cv.Put(x, top+y, tinted(c, e.tint, 1-q))
				x += c.Width
			}
		}
	case p < narrow:
		q := (p - squash) / (narrow - squash)
		line(max(int(math.Round(float64(w)*(1-q))), 1), "─")
	case p < 1:
		line(1, "•")
	}
	return cv.Grid()
}

// scanline is c on row y, darkened q of the way when the row is a gap line.
func (e Effect) scanline(c cells.Cell, q float64, y int) cells.Cell {
	if y%2 == 0 {
		return c
	}
	return tinted(c, e.tint, 1-e.amount*q)
}
