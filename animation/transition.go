package animation

import (
	"image/color"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
)

type transitionKind int

const (
	push transitionKind = iota
	cover
	reveal
	crossFade
	wipe
	zoom
	shuffle
)

// Transition carries a screen from one block to the next: the old one pushed
// out, covered, faded or wiped away, the new one in its place. Apply it to
// both blocks every frame; a smaller one is padded to the larger's size.
type Transition struct {
	kind    transitionKind
	dir     Direction
	pattern Pattern
	p       float64 // raw progress: 0 is all from, 1 all to
	run     run
}

func newTransition(kind transitionKind, dur time.Duration) Transition {
	t := Transition{kind: kind, run: newRun(dur)}
	if !t.Active() {
		t.p = 1
	}
	return t
}

// Push slides the new block in and the old one out ahead of it, both
// travelling d, the two side by side halfway.
func Push(d Direction, dur time.Duration) Transition {
	t := newTransition(push, dur)
	t.dir = d
	return t
}

// Cover slides the new block in travelling d over the old one, which stays put.
func Cover(d Direction, dur time.Duration) Transition {
	t := newTransition(cover, dur)
	t.dir = d
	return t
}

// Reveal slides the old block away travelling d, uncovering the new one beneath.
func Reveal(d Direction, dur time.Duration) Transition {
	t := newTransition(reveal, dur)
	t.dir = d
	return t
}

// CrossFade blends the old block's colours into the new one's, the text
// changing over halfway.
func CrossFade(dur time.Duration) Transition { return newTransition(crossFade, dur) }

// Wipe turns each cell from the old block to the new one in p's order: a
// front, a circle, a spiral.
func Wipe(p Pattern, dur time.Duration) Transition {
	t := newTransition(wipe, dur)
	t.pattern = p
	return t
}

// Zoom grows the new block out of the middle of the old one.
func Zoom(dur time.Duration) Transition { return newTransition(zoom, dur) }

// Shuffle scatters the old block's cells and gathers the new one's out of
// the scatter.
func Shuffle(dur time.Duration) Transition { return newTransition(shuffle, dur) }

// WithEase is t following e instead of EaseInOut.
func (t Transition) WithEase(e Ease) Transition {
	t.run.ease = e
	return t
}

// WithDelay holds t on the old block for d after its first frame.
func (t Transition) WithDelay(d time.Duration) Transition {
	t.run.delay = max(d, 0)
	return t
}

// Apply is the frame between from and to: all from before the first frame,
// all to once t ends.
func (t Transition) Apply(from, to string) string {
	a, b := parse(from), parse(to)
	w, h := max(a.width(), b.width()), max(len(a), len(b))
	a, b = padGrid(a, w, h), padGrid(b, w, h)
	p := t.run.ease.or()(t.p)
	switch {
	case p <= 0:
		return a.String()
	case p >= 1:
		return b.String()
	}
	cv := newCanvas(w, h)
	switch t.kind {
	case push:
		vx, vy := vector(t.dir, w, h)
		draw(cv, a, round(vx*p), round(vy*p))
		draw(cv, b, round(-vx*(1-p)), round(-vy*(1-p)))
	case cover:
		vx, vy := vector(t.dir, w, h)
		draw(cv, a, 0, 0)
		draw(cv, b, round(-vx*(1-p)), round(-vy*(1-p)))
	case reveal:
		vx, vy := vector(t.dir, w, h)
		draw(cv, b, 0, 0)
		draw(cv, a, round(vx*p), round(vy*p))
	case crossFade:
		return crossFaded(a, b, p).String()
	case wipe:
		t.wiped(cv, a, b, w, h, p)
	case zoom:
		zoomed(cv, a, b, w, h, p)
	case shuffle:
		shuffled(cv, a, b, w, h, p)
	}
	return cv.grid().String()
}

// vector is how far a block travels d to leave a w×h box.
func vector(d Direction, w, h int) (float64, float64) {
	switch d {
	case Left:
		return -float64(w), 0
	case Down:
		return 0, float64(h)
	case Up:
		return 0, -float64(h)
	}
	return float64(w), 0
}

func round(f float64) int { return int(math.Round(f)) }

// padGrid is g with h rows of w columns, blanks added where it is short.
func padGrid(g grid, w, h int) grid {
	out := make(grid, h)
	for y := range out {
		if y < len(g) {
			out[y] = shift(g[y], 0, w)
		} else {
			out[y] = blankRow(w)
		}
	}
	return out
}

// draw puts every cell of g on cv moved dx, dy.
func draw(cv canvas, g grid, dx, dy int) {
	for y, row := range g {
		x := 0
		for _, c := range row {
			cv.put(x+dx, y+dy, c)
			x += c.width
		}
	}
}

// columns maps the column each cell of row starts at to the cell.
func columns(row []cell) map[int]cell {
	m := make(map[int]cell, len(row))
	x := 0
	for _, c := range row {
		m[x] = c
		x += c.width
	}
	return m
}

// crossFaded is b's text past halfway, a's before, every cell's colours
// blended between the two cells that share its column.
func crossFaded(a, b grid, p float64) grid {
	out := make(grid, len(a))
	for y := range a {
		src, other := a[y], columns(b[y])
		if p >= 0.5 {
			src, other = b[y], columns(a[y])
		}
		row := make([]cell, len(src))
		x := 0
		for i, c := range src {
			if o, ok := other[x]; ok {
				from, to := o, c // the colours run from a's cell to b's
				if p < 0.5 {
					from, to = c, o
				}
				c.fg, c.bg = blendOr(from.fg, to.fg, p), blendOr(from.bg, to.bg, p)
			}
			row[i] = c
			x += c.width
		}
		out[y] = row
	}
	return out
}

// blendOr is Blend, or whichever side has a colour when the other is the
// terminal's own.
func blendOr(a, b color.Color, p float64) color.Color {
	switch {
	case a == nil && b == nil:
		return nil
	case a == nil:
		if p >= 0.5 {
			return b
		}
		return nil
	case b == nil:
		if p < 0.5 {
			return a
		}
		return nil
	}
	return Blend(a, b, p)
}

// wiped puts on cv each cell of b the pattern has reached, and a's elsewhere.
func (t Transition) wiped(cv canvas, a, b grid, w, h int, p float64) {
	pattern := t.pattern
	if pattern == nil {
		pattern = Sweep(Right)
	}
	for y := range a {
		x := 0
		for _, c := range a[y] {
			if pattern(x, y, w, h) > p {
				cv.put(x, y, c)
			}
			x += c.width
		}
		x = 0
		for _, c := range b[y] {
			if pattern(x, y, w, h) <= p {
				cv.put(x, y, c)
			}
			x += c.width
		}
	}
}

// zoomed puts a on cv, then b scaled by p about the middle over it.
func zoomed(cv canvas, a, b grid, w, h int, p float64) {
	draw(cv, a, 0, 0)
	cx, cy := float64(w-1)/2, float64(h-1)/2
	for y := range h {
		sy := round(cy + (float64(y)-cy)/p)
		if sy < 0 || sy >= h || math.Abs(float64(y)-cy) > p*float64(h)/2 {
			continue
		}
		src := columns(b[sy])
		for x := range w {
			if math.Abs(float64(x)-cx) > p*float64(w)/2 {
				continue
			}
			if c, ok := src[round(cx+(float64(x)-cx)/p)]; ok {
				cv.put(x, y, c)
			}
		}
	}
}

// shuffled flies a's cells out to places of their own in the first half and
// b's cells in from others in the second.
func shuffled(cv canvas, a, b grid, w, h int, p float64) {
	g, q, seed := a, p*2, uint64(11)
	if p >= 0.5 {
		g, q, seed = b, 2-p*2, 13
	}
	var k uint64
	for y, row := range g {
		x := 0
		for _, c := range row {
			k++
			tx := unit(noise(seed, k)) * float64(w-1)
			ty := unit(noise(seed+1, k)) * float64(h-1)
			cv.put(round(float64(x)+(tx-float64(x))*q), round(float64(y)+(ty-float64(y))*q), c)
			x += c.width
		}
	}
}

// Active reports whether t is still moving between the blocks.
func (t Transition) Active() bool { return t.run.active() }

// Tick is the command for t's next frame, nil once it ended.
func (t Transition) Tick() tea.Cmd { return t.run.tick() }

// FrameAt is t's frame for at, for a test or a host that delivers frames itself.
func (t Transition) FrameAt(at time.Time) FrameMsg { return t.run.frameAt(at) }

// Owns reports whether msg is t's current frame.
func (t Transition) Owns(msg tea.Msg) bool {
	_, ok := t.run.own(msg)
	return ok
}

// Update advances on its own frames and asks for the next until it ends.
func (t Transition) Update(msg tea.Msg) (Transition, tea.Cmd) {
	r, p, ok, cmd := t.run.update(msg)
	if ok {
		t.run, t.p = r, p
	}
	return t, cmd
}

// Advance is t at now, and whether it ended.
func (t Transition) Advance(now time.Time) (Transition, bool) {
	if !t.Active() {
		return t, true
	}
	t.run, t.p = t.run.advance(now)
	return t, !t.Active()
}
