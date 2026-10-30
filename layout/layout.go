// Package layout declares a screen as a tree of named leaves and resolves it
// to rectangles. It knows nothing about what is drawn in them.
package layout

import (
	"iter"
	"slices"
)

// Rect is a placed region: X,Y is the top-left corner, W,H the size.
type Rect struct{ X, Y, W, H int }

// Node is a Box or a Leaf.
type Node interface{ node() }

type axis int

const (
	horizontal axis = iota // Columns: children side by side, split W
	vertical               // Rows: children stacked, split H
)

// Box splits its rect among children along one axis.
type Box struct {
	dir           axis
	children      []Node
	collapseLimit int
	collapseKeep  string
}

// Leaf is a named region the app renders into.
type Leaf struct {
	Name string
	cs   []Constraint
	info bool
}

// Info marks a pane that only informs: tab skips it, though the app can
// still focus it on purpose (to scroll a long hint, say).
func (l Leaf) Info() Leaf {
	l.info = true
	return l
}

// IsInfo reports whether l was marked with Info.
func (l Leaf) IsInfo() bool { return l.info }

func (Box) node()  {}
func (Leaf) node() {}

// Columns lays children out left to right.
func Columns(children ...Node) Box { return Box{dir: horizontal, children: slices.Clone(children)} }

// Rows lays children out top to bottom.
func Rows(children ...Node) Box { return Box{dir: vertical, children: slices.Clone(children)} }

// Fixed is a leaf sized by its constraints. With none it behaves like Fill.
func Fixed(name string, cs ...Constraint) Leaf { return Leaf{Name: name, cs: slices.Clone(cs)} }

// Fill is a leaf that takes what the fixed siblings leave.
func Fill(name string) Leaf { return Leaf{Name: name} }

// Collapse is b keeping only the child holding leaf keep when its axis is
// shorter than limit. keep may name a leaf nested deeper in that child.
func (b Box) Collapse(limit int, keep string) Box {
	b.collapseLimit, b.collapseKeep = limit, keep
	return b
}

type constraintKind int

const (
	ratio constraintKind = iota
	minimum
	maximum
	exact
)

// Constraint shapes a Fixed leaf along its parent's axis.
type Constraint struct {
	kind     constraintKind
	num, den int
}

// Ratio is a share of the parent, before Min/Max clamp it.
func Ratio(num, den int) Constraint { return Constraint{kind: ratio, num: num, den: den} }

// Min floors the leaf's size.
func Min(n int) Constraint { return Constraint{kind: minimum, num: n} }

// Max caps the leaf's size.
func Max(n int) Constraint { return Constraint{kind: maximum, num: n} }

// Lines is an exact size: rows in a Rows box, columns in a Columns box.
func Lines(n int) Constraint { return Constraint{kind: exact, num: n} }

// Leaves yields every leaf under n, depth-first in declaration order.
func Leaves(n Node) iter.Seq[Leaf] {
	return func(yield func(Leaf) bool) { leaves(n, yield) }
}

// leaves reports false once yield has asked to stop, so no sibling runs after.
func leaves(n Node, yield func(Leaf) bool) bool {
	switch v := n.(type) {
	case Leaf:
		return yield(v)
	case Box:
		for _, c := range v.children {
			if !leaves(c, yield) {
				return false
			}
		}
	}
	return true
}

// contains reports whether leaf name is anywhere under n.
func contains(n Node, name string) bool {
	for l := range Leaves(n) {
		if l.Name == name {
			return true
		}
	}
	return false
}
