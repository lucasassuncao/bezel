package animation

import (
	"image/color"
	"math"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// soft is the share of the way one cell takes to turn: a band of cells is
// always part way, so a front reads as a front and not a hard edge.
const soft = 0.25

// shades are what Evolve steps a cell through before its own grapheme.
var shades = []string{"░", "▒", "▓", "█"}

type effectKind int

const (
	fadeIn effectKind = iota
	fadeOut
	dissolve
	coalesce
	sweepIn
	slideIn
	slideOut
	evolve
	explode
	expand
	stretch
	hue
	saturate
	tint
	flash
	invert
	spotlight
	vignette
	contrast
	hover
	melt
)

// Effect changes a block that is already drawn, cell by cell: a pane that
// fades in, a dialog that dissolves, a screen that slides in. Apply it to
// whatever View drew; it reads the colours back from the escape sequences.
type Effect struct {
	kind     effectKind
	pattern  Pattern // nil reaches every cell at once
	tint     color.Color
	base     color.Color // a hover's background where a cell has none
	dir      Direction
	amount   float64 // degrees for a hue shift, the change for a saturation, a light's strength
	spot     rect    // the lit part of a spotlight
	masks    []rect  // the parts of the block it changes; none is all of it
	reversed bool
	p        float64 // raw progress: 0 before the first frame, 1 at the end
	run      run
}

func newEffect(kind effectKind, pattern Pattern, dur time.Duration) Effect {
	e := Effect{kind: kind, pattern: pattern, run: newRun(dur)}
	if !e.Active() {
		e.p = 1 // Still, or no time: born where it would end
	}
	return e
}

// FadeIn raises every cell's colours from from to its own: from the
// background colour, a block that surfaces.
func FadeIn(from color.Color, dur time.Duration) Effect {
	e := newEffect(fadeIn, nil, dur)
	e.tint = from
	return e
}

// FadeOut sinks every cell's colours into to.
func FadeOut(to color.Color, dur time.Duration) Effect {
	e := newEffect(fadeOut, nil, dur)
	e.tint = to
	return e
}

// Dissolve blanks the cells one by one, in an order that looks random.
func Dissolve(dur time.Duration) Effect { return newEffect(dissolve, Noise(1), dur) }

// Coalesce gathers the block out of nothing, one cell at a time.
func Coalesce(dur time.Duration) Effect { return newEffect(coalesce, Noise(1), dur) }

// SweepIn reveals the block behind a front tinted glow, left to right.
func SweepIn(glow color.Color, dur time.Duration) Effect {
	e := newEffect(sweepIn, Sweep(Right), dur)
	e.tint = glow
	return e
}

// SlideIn brings the block in travelling d: Right comes in from the left edge.
func SlideIn(d Direction, dur time.Duration) Effect {
	e := newEffect(slideIn, nil, dur)
	e.dir = d
	return e
}

// SlideOut takes the block away travelling d.
func SlideOut(d Direction, dur time.Duration) Effect {
	e := newEffect(slideOut, nil, dur)
	e.dir = d
	return e
}

// Evolve grows each cell out of the shades ░▒▓█ into its own grapheme.
func Evolve(dur time.Duration) Effect { return newEffect(evolve, Sweep(Right), dur) }

// WithPattern reaches the cells in p's order; nil reaches them all at once.
// A slide moves the block whole and takes no pattern.
func (e Effect) WithPattern(p Pattern) Effect {
	e.pattern = p
	return e
}

// WithEase is e following ease instead of EaseInOut.
func (e Effect) WithEase(ease Ease) Effect {
	e.run.ease = ease
	return e
}

// WithDelay holds e at its start for d after its first frame.
func (e Effect) WithDelay(d time.Duration) Effect {
	e.run.delay = max(d, 0)
	return e
}

// WithRepeat runs e n more times after the first, or without end for Forever.
func (e Effect) WithRepeat(n int) Effect {
	e.run.repeat = max(n, Forever)
	return e
}

// WithYoyo sends every other run of e back: in, then out again.
func (e Effect) WithYoyo() Effect {
	e.run.yoyo = true
	return e
}

// Apply is block as this frame shows it, its size kept. Within and OnRows
// keep the change to their parts of it.
func (e Effect) Apply(block string) string {
	g := parse(block)
	if len(e.masks) == 0 {
		return e.applyGrid(g).String()
	}
	for _, r := range e.masks {
		g = paste(g, e.applyGrid(crop(g, r)), r)
	}
	return g.String()
}

// applyGrid is g as this frame shows it.
func (e Effect) applyGrid(g grid) grid {
	w, h := g.width(), len(g)
	raw := e.p
	if e.reversed {
		raw = 1 - raw
	}
	p := e.run.ease.or()(raw)
	switch e.kind {
	case slideIn, slideOut:
		return e.slide(g, w, h, p)
	case explode:
		return e.explode(g, w, h, p)
	case expand:
		return e.expand(g, w, p)
	case stretch:
		return e.stretch(g, w, h, p)
	case melt:
		return e.melt(g, w, h, p)
	}
	for y, row := range g {
		x := 0
		for i, c := range row {
			row[i] = e.turnAt(c, e.reach(p, x, y, w, h), x, y, w, h)
			x += c.width
		}
	}
	return g
}

// Within keeps e to the w×h part of the block whose top left is x, y. A w or h
// of 0 or less runs to the block's right or bottom edge. It may be given again.
func (e Effect) Within(x, y, w, h int) Effect {
	e.masks = append(slices.Clone(e.masks), rect{x: x, y: y, w: w, h: h})
	return e
}

// OnRows keeps e to whole rows of the block: the lines a diff added, the
// ones undo restored. Each row turns on its own, as a one-line block.
func (e Effect) OnRows(rows ...int) Effect {
	for _, r := range rows {
		e = e.Within(0, r, 0, 1)
	}
	return e
}

// Reverse plays e backwards: an effect that brings a block in takes it away,
// a hover lets go.
func (e Effect) Reverse() Effect {
	e.reversed = !e.reversed
	return e
}

// reach is how far the cell at x, y has turned, in [0,1], with e at p.
func (e Effect) reach(p float64, x, y, w, h int) float64 {
	if e.pattern == nil {
		return min(max(p, 0), 1)
	}
	t := e.pattern(x, y, w, h)
	return min(max((p-t*(1-soft))/soft, 0), 1)
}

// turn is c turned q of the way.
func (e Effect) turn(c cell, q float64) cell {
	switch e.kind {
	case fadeIn:
		return tinted(c, e.tint, q)
	case fadeOut:
		return tinted(c, e.tint, 1-q)
	case dissolve:
		if q >= 0.5 {
			return blank(c)
		}
	case coalesce:
		if q < 0.5 {
			return blank(c)
		}
	case sweepIn:
		if q <= 0 {
			return blank(c)
		}
		return tinted(c, e.tint, q)
	case evolve:
		switch {
		case q <= 0:
			return blank(c)
		case q < 1 && strings.TrimSpace(c.text) != "":
			c.text = shades[min(int(q*float64(len(shades))), len(shades)-1)] + strings.Repeat(" ", c.width-1)
		}
	case hue:
		c.fg, c.bg = shiftHue(c.fg, e.amount*q), shiftHue(c.bg, e.amount*q)
	case saturate:
		c.fg, c.bg = scaleSaturation(c.fg, 1+e.amount*q), scaleSaturation(c.bg, 1+e.amount*q)
	}
	return c
}

// tinted is c with its colours q of the way from tint to its own.
func tinted(c cell, tint color.Color, q float64) cell {
	if q >= 1 {
		return c
	}
	c.fg = Blend(tint, c.fg, q)
	if c.bg != nil {
		c.bg = Blend(tint, c.bg, q)
	}
	return c
}

// blank is a cell as wide as c with nothing in it, not even a colour.
func blank(c cell) cell { return cell{text: strings.Repeat(" ", c.width), width: c.width} }

// slide is g moved whole by the share of its size p says, the cells that
// leave its w×h box gone and the room they leave blank.
func (e Effect) slide(g grid, w, h int, p float64) grid {
	off := 1 - p // how far from home, in shares of the box
	if e.kind == slideOut {
		off = p
	}
	sign := 1.0 // an entering block starts on the side it travels from
	if e.kind == slideIn {
		sign = -1
	}
	if e.dir == Left || e.dir == Up {
		sign = -sign
	}
	dx, dy := 0, 0
	if e.dir == Left || e.dir == Right {
		dx = int(math.Round(sign * off * float64(w)))
	} else {
		dy = int(math.Round(sign * off * float64(h)))
	}
	out := make(grid, h)
	for y := range out {
		src := y - dy
		if src < 0 || src >= h {
			out[y] = blankRow(w)
			continue
		}
		out[y] = shift(g[src], dx, w)
	}
	return out
}

// shift is row moved dx columns within w, whole cells only, padded with blanks.
func shift(row []cell, dx, w int) []cell {
	out := []cell{}
	at := 0 // the column out has reached
	x := dx
	for _, c := range row {
		if x >= 0 && x+c.width <= w {
			for ; at < x; at++ {
				out = append(out, cell{text: " ", width: 1})
			}
			out = append(out, c)
			at = x + c.width
		}
		x += c.width
	}
	for ; at < w; at++ {
		out = append(out, cell{text: " ", width: 1})
	}
	return out
}

func blankRow(w int) []cell { return shift(nil, 0, w) }

// Active reports whether e is still changing the block.
func (e Effect) Active() bool { return e.run.active() }

// Tick is the command for e's next frame, nil once it ended.
func (e Effect) Tick() tea.Cmd { return e.run.tick() }

// FrameAt is e's frame for at, for a test or a host that delivers frames itself.
func (e Effect) FrameAt(at time.Time) FrameMsg { return e.run.frameAt(at) }

// Update advances on its own frames and asks for the next until it ends.
func (e Effect) Update(msg tea.Msg) (Effect, tea.Cmd) {
	f, ok := e.run.own(msg)
	if !ok {
		return e, nil
	}
	if e, _ = e.Advance(f.At); !e.Active() {
		return e, nil
	}
	var cmd tea.Cmd
	e.run, cmd = e.run.next()
	return e, cmd
}

// Advance is e at now, and whether it ended.
func (e Effect) Advance(now time.Time) (Effect, bool) {
	if !e.Active() {
		return e, true
	}
	e.run, e.p = e.run.advance(now)
	return e, !e.Active()
}

// Owns reports whether msg is e's current frame, for a host that routes frames
// to whoever owns them before handing the rest on.
func (e Effect) Owns(msg tea.Msg) bool {
	_, ok := e.run.own(msg)
	return ok
}
