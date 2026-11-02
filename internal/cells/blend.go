package cells

import "image/color"

// Toward is g with every colour blended f of the way to to, for a screen
// that steps back. A cell in the terminal's own foreground blends from ink;
// one on the terminal's own background keeps it, which is to already.
func (g Grid) Toward(to, ink color.Color, f float64) Grid {
	out := make(Grid, len(g))
	for y, row := range g {
		out[y] = make([]Cell, len(row))
		for x, c := range row {
			fg := c.Fg
			if fg == nil {
				fg = ink
			}
			c.Fg = Mix(fg, to, f)
			if c.Bg != nil {
				c.Bg = Mix(c.Bg, to, f)
			}
			out[y][x] = c
		}
	}
	return out
}

// Mix is a blended f of the way to b, f kept in [0,1]; a missing colour
// gives way to the other.
func Mix(a, b color.Color, f float64) color.Color {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	}
	f = min(max(f, 0), 1)
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	at := func(x, y uint32) uint8 {
		return uint8(float64(x>>8) + (float64(y>>8)-float64(x>>8))*f + 0.5) // #nosec G115 -- within a byte
	}
	return color.RGBA{at(ar, br), at(ag, bg), at(ab, bb), 255}
}
