package layout

import (
	"maps"
	"slices"
)

// floor is the fewest cells a child keeps when an edge moves: a border on
// each side and one cell between them.
const floor = 3

// Sizes are the cells a user gave the children of Resizable boxes, each box
// keyed by its first leaf. Resolve scales them to the box's length, so they
// hold their share through a terminal resize. nil is the boxes as declared.
type Sizes map[string][]int

// split is a Resizable box as placed: its rect and its children's sizes.
type split struct {
	key   string
	box   Box
	r     Rect
	sizes []int
}

// Edge is where two children of a Resizable box meet. At is the first cell
// of the one after: a column when Cols, between children side by side, else
// a row. From and To are the span it runs along, To excluded.
type Edge struct {
	Box      string // the box's key in Sizes
	Index    int    // the edge follows child Index
	Cols     bool
	At       int
	From, To int
}

// Resolve is the package's Resolve with sz giving the Resizable boxes their
// sizes.
func (sz Sizes) Resolve(n Node, body Rect, present func(name string) bool) (map[string]Rect, bool) {
	p := sz.placed(n, body, present)
	return p.rects, p.collapsed
}

func (sz Sizes) placed(n Node, body Rect, present func(name string) bool) *placing {
	p := &placing{sz: sz, rects: map[string]Rect{}}
	if present == nil {
		present = func(string) bool { return true }
	}
	if n = prune(n, present); n != nil {
		p.place(n, body)
	}
	return p
}

// scaled is saved stretched to length, or declared when nothing fits it.
func scaled(saved []int, length int, declared []int) []int {
	total := 0
	for _, s := range saved {
		total += s
	}
	if len(saved) != len(declared) || total <= 0 {
		return declared
	}
	out := make([]int, len(saved))
	given := 0
	for i, s := range saved {
		out[i] = s * length / total
		given += out[i]
	}
	out[len(out)-1] += length - given
	return out
}

// Edges are the edges of every Resizable box on screen, outer boxes first.
func (sz Sizes) Edges(n Node, body Rect, present func(name string) bool) []Edge {
	var out []Edge
	for _, s := range sz.placed(n, body, present).splits {
		out = append(out, s.edges()...)
	}
	return out
}

func (s split) edges() []Edge {
	cols := s.box.dir == horizontal
	at, from, to := s.r.X, s.r.Y, s.r.Y+s.r.H
	if !cols {
		at, from, to = s.r.Y, s.r.X, s.r.X+s.r.W
	}
	out := make([]Edge, 0, len(s.sizes)-1)
	for i, size := range s.sizes[:len(s.sizes)-1] {
		at += size
		out = append(out, Edge{Box: s.key, Index: i, Cols: cols, At: at, From: from, To: to})
	}
	return out
}

// Move is sz with edge e moved by delta cells, towards the end when
// positive, held where a child would drop below three cells or past the
// Min and Max of its constraints. sz is left as it was.
func (sz Sizes) Move(n Node, body Rect, present func(name string) bool, e Edge, delta int) Sizes {
	for _, s := range sz.placed(n, body, present).splits {
		if s.key != e.Box || e.Index < 0 || e.Index >= len(s.sizes)-1 {
			continue
		}
		i := e.Index
		lo, hi := s.box.bounds(i), s.box.bounds(i+1)
		delta = min(delta, s.sizes[i+1]-hi.min, lo.max-s.sizes[i])
		delta = max(delta, lo.min-s.sizes[i], s.sizes[i+1]-hi.max)
		if delta == 0 {
			return sz
		}
		sizes := slices.Clone(s.sizes)
		sizes[i] += delta
		sizes[i+1] -= delta
		out := maps.Clone(sz)
		if out == nil {
			out = Sizes{}
		}
		out[s.key] = sizes
		return out
	}
	return sz
}

// Grow is sz with leaf's pane delta cells longer, or shorter when negative,
// along columns when cols and rows when not: the innermost Resizable box of
// that axis holding leaf moves the edge after it, or before it when last.
func (sz Sizes) Grow(n Node, body Rect, present func(name string) bool, leaf string, cols bool, delta int) Sizes {
	splits := sz.placed(n, body, present).splits
	for _, s := range slices.Backward(splits) {
		if (s.box.dir == horizontal) != cols {
			continue
		}
		i := slices.IndexFunc(s.box.children, func(c Node) bool { return contains(c, leaf) })
		if i < 0 {
			continue
		}
		e := Edge{Box: s.key, Index: i}
		if i == len(s.box.children)-1 {
			e.Index, delta = i-1, -delta
		}
		return sz.Move(n, body, present, e, delta)
	}
	return sz
}

// span is the fewest and most cells a child of a box may take.
type span struct{ min, max int }

// bounds is child i's span: three cells at least, and its Min and Max when
// it is a leaf that has them.
func (b Box) bounds(i int) span {
	sp := span{min: floor, max: int(^uint(0) >> 1)}
	leaf, ok := b.children[i].(Leaf)
	if !ok {
		return sp
	}
	for _, c := range leaf.cs {
		switch c.kind {
		case minimum:
			sp.min = max(sp.min, c.num)
		case maximum:
			sp.max = max(c.num, sp.min)
		}
	}
	return sp
}
