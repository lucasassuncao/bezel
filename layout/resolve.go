package layout

// Resolve places the tree inside body. Leaves that present rejects are pruned
// first, so their space goes to their siblings. collapsed reports whether any
// Collapse fired, which is what a legend needs to change its wording.
func Resolve(n Node, body Rect, present func(name string) bool) (map[string]Rect, bool) {
	return Sizes(nil).Resolve(n, body, present)
}

// prune drops absent leaves and boxes left empty. Returns nil when nothing
// is left.
func prune(n Node, present func(string) bool) Node {
	switch v := n.(type) {
	case Leaf:
		if present(v.Name) {
			return v
		}
		return nil
	case Box:
		kept := make([]Node, 0, len(v.children))
		for _, c := range v.children {
			if p := prune(c, present); p != nil {
				kept = append(kept, p)
			}
		}
		if len(kept) == 0 {
			return nil
		}
		v.children = kept
		return v
	}
	return nil
}

// placing is one resolve: the rects so far, whether a Collapse fired, and
// the resizable boxes as placed, outer before inner.
type placing struct {
	sz        Sizes
	rects     map[string]Rect
	collapsed bool
	splits    []split
}

func (p *placing) place(n Node, r Rect) {
	switch v := n.(type) {
	case Leaf:
		p.rects[v.Name] = r
	case Box:
		length := r.W
		if v.dir == vertical {
			length = r.H
		}
		if v.collapseLimit > 0 && length < v.collapseLimit {
			p.collapsed = true
			p.place(v.survivor(), r)
			return
		}
		sizes := v.sizes(length)
		if v.resizable && len(v.children) > 1 {
			key := v.key()
			sizes = scaled(p.sz[key], length, sizes)
			p.splits = append(p.splits, split{key: key, box: v, r: r, sizes: sizes})
		}
		offset := 0
		for i, c := range v.children {
			child := r
			if v.dir == horizontal {
				child.X, child.W = r.X+offset, sizes[i]
			} else {
				child.Y, child.H = r.Y+offset, sizes[i]
			}
			p.place(c, child)
			offset += sizes[i]
		}
	}
}

// key is the box's first leaf, which names it in Sizes.
func (b Box) key() string {
	for l := range Leaves(b) {
		return l.Name
	}
	return ""
}

// survivor is the child that stays when the box collapses: the one holding
// collapseKeep, or the first when no child does.
func (b Box) survivor() Node {
	for _, c := range b.children {
		if contains(c, b.collapseKeep) {
			return c
		}
	}
	return b.children[0]
}

// sizes splits length among the children. Fixed leaves are sized by their
// constraints; everything else shares the rest evenly.
func (b Box) sizes(length int) []int {
	n := len(b.children)
	sizes := make([]int, n)
	fixed := make([]bool, n)
	sumFixed, flex := 0, 0

	for i, c := range b.children {
		leaf, ok := c.(Leaf)
		if !ok || len(leaf.cs) == 0 {
			flex++
			continue
		}
		fixed[i] = true
		sizes[i] = leaf.size(length)
		sumFixed += sizes[i]
	}

	// Fixed sizes that leave no room for the flexible ones shrink together,
	// each flexible child keeping at least one cell.
	if room := length - flex; sumFixed > room {
		room = max(room, 0)
		scaled := 0
		for i := range sizes {
			if fixed[i] && sumFixed > 0 {
				sizes[i] = sizes[i] * room / sumFixed
				scaled += sizes[i]
			}
		}
		sumFixed = scaled
	}

	if flex > 0 {
		rest := max(length-sumFixed, 0)
		each := rest / flex
		given := 0
		lastFlex := -1
		for i := range sizes {
			if !fixed[i] {
				sizes[i] = each
				given += each
				lastFlex = i
			}
		}
		sizes[lastFlex] += rest - given
	} else if n > 0 && sumFixed < length {
		sizes[n-1] += length - sumFixed
	}
	return sizes
}

// size applies a leaf's constraints in order: ratio or exact first, then the
// clamps. Never negative.
func (l Leaf) size(length int) int {
	size := 0
	for _, c := range l.cs {
		switch c.kind {
		case ratio:
			if c.den > 0 {
				size = length * c.num / c.den
			}
		case exact:
			size = c.num
		}
	}
	for _, c := range l.cs {
		switch c.kind {
		case minimum:
			size = max(size, c.num)
		case maximum:
			size = min(size, c.num)
		}
	}
	return max(size, 0)
}
