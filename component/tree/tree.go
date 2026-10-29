// Package tree navigates and renders a flat, depth-first node list: rows
// visible under collapsed parents, a cursor that skips headings,
// expand/collapse, scrolling. The node type is the caller's; Info describes it.
package tree

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Noder is a node the tree can read and whose expansion it can set. Expand
// returns a changed copy: the tree never writes through a shared slice.
type Noder[N any] interface {
	TreeInfo() Info
	Expand(open bool) N
}

// Info is what the tree reads from a node.
type Info struct {
	Depth    int
	Leaf     bool // nothing under it: never expands
	Expanded bool
	Section  bool // a heading or spacer: skipped by the cursor
	Hidden   bool // left out, and everything under it with it
}

// Action is what a key did.
type Action int

const (
	None      Action = iota
	Expanded         // → on a collapsed parent
	Collapsed        // ← on an expanded parent
)

// Keys are the bindings the tree answers to.
type Keys struct {
	Up, Down, Left, Right key.Binding
}

// DefaultKeys are the four arrows.
func DefaultKeys() Keys {
	return Keys{
		Up:    key.New("up").WithHelp("↑/↓", "move"),
		Down:  key.New("down"),
		Left:  key.New("left").WithHelp("←/→", "close/open"),
		Right: key.New("right"),
	}
}

// VimKeys are DefaultKeys with vim's letters added: k/j to move, h/l to
// close and open.
func VimKeys() Keys {
	k := DefaultKeys()
	k.Up, k.Down = key.New("up", "k").WithHelp("↑/↓ k/j", "move"), key.New("down", "j")
	k.Left, k.Right = key.New("left", "h").WithHelp("h/l", "close/open"), key.New("right", "l")
	return k
}

// Connectors are what Guides draws before a row, one piece per level under
// the top: Mid for a row with siblings after it, Last for the last one,
// Pipe where an outer level runs on below and Space where it ended.
type Connectors struct{ Mid, Last, Pipe, Space string }

// DefaultConnectors are thin box-drawing lines, three cells a level.
func DefaultConnectors() Connectors {
	return Connectors{Mid: "├─ ", Last: "└─ ", Pipe: "│  ", Space: "   "}
}

// Model is a value: every method returns the new tree. The zero value is
// usable and answers to DefaultKeys.
type Model[N Noder[N]] struct {
	nodes     []N
	cursor    int // position within the visible list
	offset    int // scroll offset within the visible list
	height    int
	empty     string // shown when nothing is visible
	keys      *Keys
	lines     *Connectors
	scrollOff int
	edit      func(Styles) Styles
}

// Styles are how a tree paints what it draws itself; each row is the app's.
type Styles struct {
	Empty lipgloss.Style // EmptyMessage, or "(empty)"
	More  lipgloss.Style // the "↓ N more" row
}

// DefaultStyles dim both.
func DefaultStyles(th theme.Resolved) Styles { return Styles{Empty: th.Dim, More: th.Dim} }

// New builds a tree over nodes with the cursor on the first row that is not
// a section; SetHeight gives it the rows its pane has.
func New[N Noder[N]](nodes []N) Model[N] {
	m := Model[N]{nodes: slices.Clone(nodes)}
	vis := m.Visible()
	for m.cursor < len(vis) && nodes[vis[m.cursor]].TreeInfo().Section {
		m.cursor++
	}
	return m
}

// SetHeight changes the rows the tree takes and keeps the cursor in view.
func (m Model[N]) SetHeight(h int) Model[N] {
	m.height = h
	return m.clamp()
}

// SetNodes replaces the nodes after a reload, keeping the cursor on a row
// and in view.
func (m Model[N]) SetNodes(nodes []N) Model[N] {
	m.nodes = slices.Clone(nodes)
	return m.clamp()
}

// WithEmptyMessage is shown, dimmed, when no row is visible; "  (empty)"
// until set.
func (m Model[N]) WithEmptyMessage(s string) Model[N] {
	m.empty = s
	return m
}

// Nodes is a copy of the nodes, every one, folded or not.
func (m Model[N]) Nodes() []N { return slices.Clone(m.nodes) }

// Cursor is the cursor's place among the visible rows.
func (m Model[N]) Cursor() int { return m.cursor }

// Height is the rows the tree takes.
func (m Model[N]) Height() int { return m.height }

// WithStyles changes the styles derived from the theme through edit, under
// whatever theme View gets. It replaces an earlier edit.
func (m Model[N]) WithStyles(edit func(Styles) Styles) Model[N] {
	m.edit = edit
	return m
}

// styles is DefaultStyles of th with the app's edit applied.
func (m Model[N]) styles(th theme.Resolved) Styles {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	return st
}

// WithKeys rebinds the tree.
func (m Model[N]) WithKeys(k Keys) Model[N] {
	m.keys = &k
	return m
}

// Legend lists the keys the tree answers to.
func (m Model[N]) Legend() []legend.Entry {
	k := m.bindings()
	return legend.FromBindings([]key.Binding{k.Up, k.Down, k.Left, k.Right})
}

// WithConnectors changes the lines Guides draws.
func (m Model[N]) WithConnectors(c Connectors) Model[N] {
	m.lines = &c
	return m
}

// WithScrollOff keeps n rows in view above and below the cursor while it
// moves, where the tree has them.
func (m Model[N]) WithScrollOff(n int) Model[N] {
	m.scrollOff = max(n, 0)
	return m.clamp()
}

// Guides is the connector prefix of every visible row, in Visible's order,
// for a row function to draw before its label; a top-level row has none.
func (m Model[N]) Guides() []string {
	c := DefaultConnectors()
	if m.lines != nil {
		c = *m.lines
	}
	vis := m.Visible()
	out := make([]string, len(vis))
	var after []bool // after[d]: a row of depth d follows, no shallower one between
	for vi := len(vis) - 1; vi >= 0; vi-- {
		d := m.nodes[vis[vi]].TreeInfo().Depth
		for len(after) <= d {
			after = append(after, false)
		}
		var b strings.Builder
		for l := 1; l < d; l++ {
			if after[l] {
				b.WriteString(c.Pipe)
			} else {
				b.WriteString(c.Space)
			}
		}
		if d > 0 {
			if after[d] {
				b.WriteString(c.Mid)
			} else {
				b.WriteString(c.Last)
			}
		}
		out[vi] = b.String()
		after[d] = true
		for k := d + 1; k < len(after); k++ {
			after[k] = false
		}
	}
	return out
}

func (m Model[N]) bindings() Keys {
	if m.keys != nil {
		return *m.keys
	}
	return DefaultKeys()
}

// Visible returns the indices into Nodes that are rendered, honouring each
// parent's collapsed state.
func (m Model[N]) Visible() []int {
	vis := make([]int, 0, len(m.nodes))
	var collapsed []int // depths of collapsed or hidden ancestors: inside one, rows are skipped
	for i, n := range m.nodes {
		in := n.TreeInfo()
		for len(collapsed) > 0 && in.Depth <= collapsed[len(collapsed)-1] {
			collapsed = collapsed[:len(collapsed)-1]
		}
		if len(collapsed) > 0 {
			continue
		}
		if in.Hidden {
			collapsed = append(collapsed, in.Depth)
			continue
		}
		vis = append(vis, i)
		if !in.Leaf && !in.Expanded {
			collapsed = append(collapsed, in.Depth)
		}
	}
	return vis
}

// CurrentIdx is the Nodes index under the cursor, or -1.
func (m Model[N]) CurrentIdx() int { return At(m.cursor, m.Visible()) }

// At maps a cursor position into a visible index list, or -1 when out of
// range. For callers that already hold the Visible() result.
func At(cursor int, vis []int) int {
	if cursor >= 0 && cursor < len(vis) {
		return vis[cursor]
	}
	return -1
}

// WithNodeMutated returns the tree with a fresh Nodes slice in which
// nodes[idx] was changed by mut: copy-on-write, never in place. A nil mut
// changes nothing.
func (m Model[N]) WithNodeMutated(idx int, mut func(*N)) Model[N] {
	if idx < 0 || idx >= len(m.nodes) || mut == nil {
		return m
	}
	nodes := make([]N, len(m.nodes))
	copy(nodes, m.nodes)
	mut(&nodes[idx])
	m.nodes = nodes
	return m
}

// Update moves the cursor and expands or collapses. Other keys return None
// and leave the tree unchanged, so the caller layers its own on top.
func (m Model[N]) Update(msg tea.KeyPressMsg) (Model[N], Action) {
	vis := m.Visible()
	if len(vis) == 0 {
		return m, None
	}
	keys := m.bindings()
	switch {
	case key.Matches(msg, keys.Up):
		return m.move(-1, vis), None
	case key.Matches(msg, keys.Down):
		return m.move(1, vis), None
	case key.Matches(msg, keys.Right):
		return m.expand(vis)
	case key.Matches(msg, keys.Left):
		return m.collapse(vis)
	}
	return m, None
}

// move steps the cursor over section rows, clamped at both ends.
func (m Model[N]) move(dir int, vis []int) Model[N] {
	return m.withWindow(m.window().Move(dir, len(vis), m.isSection(vis)), len(vis))
}

// Move walks delta rows, stepping over section rows and stopping at the
// last row rather than running into a trailing heading.
func (m Model[N]) Move(delta int) Model[N] {
	vis := m.Visible()
	if len(vis) == 0 {
		return m
	}
	return m.move(delta, vis)
}

// isSection reports the visible rows the cursor passes over.
func (m Model[N]) isSection(vis []int) func(int) bool {
	return func(i int) bool { return m.nodes[vis[i]].TreeInfo().Section }
}

// window is the tree's scroll as a layout.Window. Cursor, Offset and Height
// stay fields of their own: callers read and set them.
func (m Model[N]) window() layout.Window {
	return layout.Window{Cursor: m.cursor, Offset: m.offset, Height: m.height}
}

// Scroll is where the tree shows its visible rows from, for a pane's scrollbar.
func (m Model[N]) Scroll() layout.Scroll {
	n := len(m.Visible())
	return m.window().Clamp(n).Scroll(n)
}

// withWindow takes w's cursor and offset; n is how many rows are visible.
func (m Model[N]) withWindow(w layout.Window, n int) Model[N] {
	if m.scrollOff > 0 && w.Height > 0 {
		w = m.keepOff(w, n)
	}
	m.cursor, m.offset = w.Cursor, w.Offset
	return m
}

// keepOff scrolls w so scrollOff rows show either side of the cursor, as far
// as n rows and the "more" row allow.
func (m Model[N]) keepOff(w layout.Window, n int) layout.Window {
	off := min(m.scrollOff, (w.Height-1)/2)
	start, end, _ := w.Bounds(n)
	shown := end - start // less the "more" row
	switch {
	case w.Cursor-w.Offset < off:
		w.Offset = max(w.Cursor-off, 0)
	case w.Offset+shown-1-w.Cursor < off:
		w.Offset = min(w.Cursor+off-shown+1, max(n-w.Height, 0))
	}
	return w.Clamp(n) // and off the "more" row
}

// Click is a left click on line y of View: on a row it moves the cursor
// there, and on the row already under the cursor it opens or closes it,
// Expanded or Collapsed. A heading or the "more" row does nothing.
func (m Model[N]) Click(y int) (Model[N], Action) {
	vis := m.Visible()
	w := m.window().Clamp(len(vis))
	start, end, _ := w.Bounds(len(vis))
	vi := start + y
	switch {
	case y < 0 || vi >= end || m.nodes[vis[vi]].TreeInfo().Section:
		return m, None
	case vi != w.Cursor:
		m.cursor = vi
		return m.clampOffset(len(vis)), None
	}
	if in := m.nodes[vis[vi]].TreeInfo(); !in.Leaf && in.Expanded {
		return m.WithNodeMutated(vis[vi], func(n *N) { *n = (*n).Expand(false) }), Collapsed
	}
	return m.expand(vis)
}

// MoveTo jumps to a visible row, snapping backwards over section rows: the
// last row of a grouped list is a node only by coincidence.
func (m Model[N]) MoveTo(i int) Model[N] {
	vis := m.Visible()
	if len(vis) == 0 {
		return m
	}
	i = min(max(i, 0), len(vis)-1)
	for i > 0 && m.nodes[vis[i]].TreeInfo().Section {
		i--
	}
	for i < len(vis)-1 && m.nodes[vis[i]].TreeInfo().Section {
		i++
	}
	m.cursor = i
	return m.clampOffset(len(vis))
}

// Expand opens the row under the cursor when it is a collapsed parent.
func (m Model[N]) Expand() (Model[N], Action) { return m.expand(m.Visible()) }

func (m Model[N]) expand(vis []int) (Model[N], Action) {
	idx := At(m.cursor, vis)
	if idx < 0 {
		return m, None
	}
	if in := m.nodes[idx].TreeInfo(); !in.Leaf && !in.Expanded {
		return m.WithNodeMutated(idx, func(n *N) { *n = (*n).Expand(true) }), Expanded
	}
	return m, None
}

// collapse closes an expanded parent, or moves the cursor to the parent of
// a nested row.
func (m Model[N]) collapse(vis []int) (Model[N], Action) {
	idx := At(m.cursor, vis)
	if idx < 0 {
		return m, None
	}
	in := m.nodes[idx].TreeInfo()
	if !in.Leaf && in.Expanded {
		return m.WithNodeMutated(idx, func(n *N) { *n = (*n).Expand(false) }), Collapsed
	}
	if in.Depth > 0 {
		for vi := m.cursor - 1; vi >= 0; vi-- {
			if m.nodes[vis[vi]].TreeInfo().Depth == in.Depth-1 {
				m.cursor = vi
				return m.clampOffset(len(vis)), None
			}
		}
	}
	return m, None
}

// clamp puts the cursor back on a visible row and scrolls it into view.
func (m Model[N]) clamp() Model[N] { return m.clampOffset(len(m.Visible())) }

// clampOffset is clamp for a caller already holding n visible rows.
func (m Model[N]) clampOffset(n int) Model[N] { return m.withWindow(m.window().Clamp(n), n) }

// CursorTo moves the cursor to the first visible row match accepts, or
// leaves it when none does, as a nil match. Reports whether it moved.
func (m Model[N]) CursorTo(match func(N) bool) (Model[N], bool) {
	if match == nil {
		return m, false
	}
	vis := m.Visible()
	for vi, ni := range vis {
		if match(m.nodes[ni]) {
			m.cursor = vi
			return m.clampOffset(len(vis)), true
		}
	}
	return m, false
}

// Reveal is CursorTo for a row that may sit under collapsed parents: the
// ancestors of the first match in Nodes are expanded first. Ancestors are
// the nearer rows above it with a smaller depth.
func (m Model[N]) Reveal(match func(N) bool) (Model[N], bool) {
	if match == nil {
		return m, false
	}
	if m2, ok := m.CursorTo(match); ok {
		return m2, true
	}
	target := -1
	for i, n := range m.nodes {
		if match(n) {
			target = i
			break
		}
	}
	if target < 0 {
		return m, false
	}
	nodes := make([]N, len(m.nodes))
	copy(nodes, m.nodes)
	depth := nodes[target].TreeInfo().Depth
	for i := target - 1; i >= 0 && depth > 0; i-- {
		if in := nodes[i].TreeInfo(); in.Depth < depth {
			if !in.Leaf && !in.Expanded {
				nodes[i] = nodes[i].Expand(true)
			}
			depth = in.Depth
		}
	}
	m.nodes = nodes
	return m.CursorTo(match)
}

// View renders the visible rows through row, which draws one node knowing
// whether it is under the cursor. The last row becomes "↓ N more" when rows
// overflow; short trees are padded to Height. A tree not yet sized
// (Height <= 0) renders nothing.
func (m Model[N]) View(th theme.Resolved, row func(n N, idx int, selected bool) string) string {
	if row == nil {
		row = func(N, int, bool) string { return "" }
	}
	if m.height <= 0 {
		return ""
	}
	vis := m.Visible()
	if len(vis) == 0 {
		msg := m.empty
		if msg == "" {
			msg = "  (empty)"
		}
		return m.styles(th).Empty.Render(msg)
	}
	// Nodes may have shrunk under a scrolled tree; Clamp never scrolls past the end.
	w := m.window().Clamp(len(vis))
	start, end, below := w.Bounds(len(vis))

	var sb strings.Builder
	for vi := start; vi < end; vi++ {
		sb.WriteString(row(m.nodes[vis[vi]], vis[vi], vi == w.Cursor))
		sb.WriteByte('\n')
	}
	if below > 0 {
		sb.WriteString(draw.More(below, 2, m.styles(th).More))
	} else {
		for i := end - start; i < m.height; i++ {
			sb.WriteByte('\n')
		}
	}
	return strings.TrimSuffix(sb.String(), "\n")
}
