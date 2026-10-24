package effect

import (
	"image/color"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/internal/noise"
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
	run     frames.Run
}

func newTransition(kind transitionKind, dur time.Duration) Transition {
	t := Transition{kind: kind, run: frames.NewRun(dur)}
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
func (t Transition) WithEase(e animation.Ease) Transition {
	t.run.Ease = e
	return t
}

// WithDelay holds t on the old block for d after its first frame.
func (t Transition) WithDelay(d time.Duration) Transition {
	t.run.Delay = max(d, 0)
	return t
}

// Apply is the frame between from and to: all from before the first frame,
// all to once t ends.
func (t Transition) Apply(from, to string) string {
	a, b := cells.Parse(from), cells.Parse(to)
	w, h := max(a.Width(), b.Width()), max(len(a), len(b))
	a, b = padGrid(a, w, h), padGrid(b, w, h)
	p := frames.Or(t.run.Ease)(t.p)
	switch {
	case p <= 0:
		return a.String()
	case p >= 1:
		return b.String()
	}
	cv := cells.NewCanvas(w, h)
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
	return cv.Grid().String()
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
func padGrid(g cells.Grid, w, h int) cells.Grid {
	out := make(cells.Grid, h)
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
func draw(cv cells.Canvas, g cells.Grid, dx, dy int) {
	for y, row := range g {
		x := 0
		for _, c := range row {
			cv.Put(x+dx, y+dy, c)
			x += c.Width
		}
	}
}

// columns maps the column each cell of row starts at to the cell.
func columns(row []cells.Cell) map[int]cells.Cell {
	m := make(map[int]cells.Cell, len(row))
	x := 0
	for _, c := range row {
		m[x] = c
		x += c.Width
	}
	return m
}

// crossFaded is b's text past halfway, a's before, every cell's colours
// blended between the two cells that share its column.
func crossFaded(a, b cells.Grid, p float64) cells.Grid {
	out := make(cells.Grid, len(a))
	for y := range a {
		src, other := a[y], columns(b[y])
		if p >= 0.5 {
			src, other = b[y], columns(a[y])
		}
		row := make([]cells.Cell, len(src))
		x := 0
		for i, c := range src {
			if o, ok := other[x]; ok {
				from, to := o, c // the colours run from a's cell to b's
				if p < 0.5 {
					from, to = c, o
				}
				c.Fg, c.Bg = blendOr(from.Fg, to.Fg, p), blendOr(from.Bg, to.Bg, p)
			}
			row[i] = c
			x += c.Width
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
	return animation.Blend(a, b, p)
}

// wiped puts on cv each cell of b the pattern has reached, and a's elsewhere.
func (t Transition) wiped(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
	pattern := t.pattern
	if pattern == nil {
		pattern = Sweep(Right)
	}
	for y := range a {
		x := 0
		for _, c := range a[y] {
			if pattern(x, y, w, h) > p {
				cv.Put(x, y, c)
			}
			x += c.Width
		}
		x = 0
		for _, c := range b[y] {
			if pattern(x, y, w, h) <= p {
				cv.Put(x, y, c)
			}
			x += c.Width
		}
	}
}

// zoomed puts a on cv, then b scaled by p about the middle over it.
func zoomed(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
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
				cv.Put(x, y, c)
			}
		}
	}
}

// shuffled flies a's cells out to places of their own in the first half and
// b's cells in from others in the second.
func shuffled(cv cells.Canvas, a, b cells.Grid, w, h int, p float64) {
	g, q, seed := a, p*2, uint64(11)
	if p >= 0.5 {
		g, q, seed = b, 2-p*2, 13
	}
	var k uint64
	for y, row := range g {
		x := 0
		for _, c := range row {
			k++
			tx := noise.Unit(noise.Hash(seed, k)) * float64(w-1)
			ty := noise.Unit(noise.Hash(seed+1, k)) * float64(h-1)
			cv.Put(round(float64(x)+(tx-float64(x))*q), round(float64(y)+(ty-float64(y))*q), c)
			x += c.Width
		}
	}
}

// Active reports whether t is still moving between the blocks.
func (t Transition) Active() bool { return t.run.Active() }

// Tick is the command for t's next frame, nil once it ended.
func (t Transition) Tick() tea.Cmd { return t.run.Tick() }

// FrameAt is t's frame for at, for a test or a host that delivers frames itself.
func (t Transition) FrameAt(at time.Time) animation.FrameMsg { return t.run.FrameAt(at) }

// Owns reports whether msg is t's current frame.
func (t Transition) Owns(msg tea.Msg) bool {
	_, ok := t.run.Own(msg)
	return ok
}

// Update advances on its own frames and asks for the next until it ends.
func (t Transition) Update(msg tea.Msg) (Transition, tea.Cmd) {
	r, p, ok, cmd := t.run.Update(msg)
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
	t.run, t.p = t.run.Advance(now)
	return t, !t.Active()
}
