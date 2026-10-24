package animation

import (
	"image/color"
	"math"
)

// rainGlyphs is what a falling drop of NewRain is made of.
var rainGlyphs = []rune("ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄ0123456789")

// confettiGlyphs turn over as a piece of confetti falls.
var confettiGlyphs = []string{"▪", "•", "▴", "■", "◆"}

// NewRain is a rain of characters down the columns, each drop's head in head
// and its trail in trail, fading out behind it.
func NewRain(head, trail color.Color) Ambient { return newAmbient(rain, head, trail) }

// NewSnow is flakes in c drifting down and swaying.
func NewSnow(c color.Color) Ambient { return newAmbient(snow, c) }

// NewConfetti is pieces in colors tumbling down, faster than snow and wider.
func NewConfetti(colors ...color.Color) Ambient { return newAmbient(confetti, colors...) }

// NewFireworks is rockets that climb and burst in colors, one after another.
func NewFireworks(colors ...color.Color) Ambient { return newAmbient(fireworks, colors...) }

// rain draws a drop in about two columns of three, each at its own speed and length.
func (a Ambient) rain(cv canvas, w, h int) {
	head, trail := a.color(0), a.color(1)
	for x := range w {
		k := uint64(x + 1)
		if unit(noise(25, k)) < 0.35 {
			continue
		}
		speed := 6 + 10*unit(noise(21, k))
		length := 4 + int(8*unit(noise(22, k)))
		span := float64(h + length)
		top := int(math.Mod(a.t*speed+unit(noise(23, k))*span, span))
		for d := range length {
			y := top - d
			if y < 0 || y >= h {
				continue
			}
			g := rainGlyphs[noise(24, k*4096+uint64(y)+uint64(a.t*8))%uint64(len(rainGlyphs))] // #nosec G115 -- any bits will do
			c := cell{text: string(g), width: 1, fg: trail}
			switch {
			case d == 0:
				c.fg, c.attrs = head, bold
			case d > length/2:
				c.attrs = faint
			}
			cv.put(x, y, c)
		}
	}
}

// flakes draws snow, or confetti when lively: each piece falling at its own
// speed from its own start, swaying about its column.
func (a Ambient) flakes(cv canvas, w, h int, lively bool) {
	n := max(w*h/14, 1)
	fall, sway := 1.5, 1.5
	if lively {
		n, fall, sway = max(w*h/10, 1), 4.0, 2.5
	}
	for i := range n {
		k := uint64(i + 1)
		y := math.Mod(a.t*fall*(1+1.5*unit(noise(31, k)))+unit(noise(32, k))*float64(h), float64(h))
		x := unit(noise(33, k))*float64(w) + sway*math.Sin(a.t*1.3+unit(noise(34, k))*2*math.Pi)
		c := cell{text: "·", width: 1, fg: a.color(0)}
		switch {
		case lively:
			turn := uint64(a.t*6) + k // #nosec G115 -- a frame count
			c.text, c.fg = confettiGlyphs[turn%uint64(len(confettiGlyphs))], a.color(k)
		case unit(noise(35, k)) < 0.3:
			c.text = "*"
		}
		cv.put(wrap(round(x), w), int(y), c)
	}
}

// wrap is x brought round into [0, w).
func wrap(x, w int) int { return ((x % w) + w) % w }

// fireworks draws the rockets still alive at a.t: a climb from the bottom,
// then a ring of sparks that spreads, falls and fades.
func (a Ambient) fireworks(cv canvas, w, h int) {
	const gap, life, climb = 0.9, 2.4, 0.6 // seconds: between launches, a rocket's whole life, its climb
	latest := int(a.t / gap)
	for i := max(latest-int(math.Ceil(life/gap))-1, 0); i <= latest; i++ {
		k := uint64(i + 1)
		age := a.t - float64(i)*gap
		if age < 0 || age > life {
			continue
		}
		cx := (0.15 + 0.7*unit(noise(41, k))) * float64(w)
		cy := (0.2 + 0.3*unit(noise(42, k))) * float64(h)
		col := a.color(k)
		if age < climb {
			y := float64(h-1) - (float64(h-1)-cy)*(age/climb)
			cv.put(round(cx), round(y), cell{text: "|", width: 1, fg: col})
			continue
		}
		b := age - climb // seconds since the burst
		reach := float64(min(w, 2*h)) / 3 * (1 - math.Exp(-3*b))
		glyph, attrs := "*", bold
		switch {
		case b > 1.2:
			glyph, attrs = "·", faint
		case b > 0.5:
			glyph, attrs = "+", 0
		}
		const sparks = 18
		for j := range sparks {
			ang := 2 * math.Pi * (float64(j) + unit(noise(43, k))) / sparks
			px := cx + math.Cos(ang)*reach
			py := cy + math.Sin(ang)*reach/2 + 1.2*b*b
			cv.put(round(px), round(py), cell{text: glyph, width: 1, fg: col, attrs: attrs})
		}
	}
}
