// Package tree navigates and renders a flat, depth-first node list: rows
// visible under collapsed parents, a cursor that skips headings,
// expand/collapse, scrolling. The node type is the caller's; Info describes it.
package tree

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
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

func DefaultKeys() Keys {
	return Keys{
		Up:    key.NewBinding(key.WithKeys("up")),
		Down:  key.NewBinding(key.WithKeys("down")),
		Left:  key.NewBinding(key.WithKeys("left")),
		Right: key.NewBinding(key.WithKeys("right")),
	}
}

// Model is a value: every method returns the new tree. Nodes is exported for
// wholesale rebuilds, but never write through a shared backing array (see
// WithNodeMutated). The zero value is usable and answers to DefaultKeys.
type Model[N Noder[N]] struct {
	Nodes  []N
	Cursor int // position within the visible list
	Offset int // scroll offset within the visible list
	Height int

	// EmptyMessage is shown when nothing is visible.
	EmptyMessage string
	keys         *Keys
}

// New builds a tree over nodes, height rows tall, with the cursor on the
// first row that is not a section.
func New[N Noder[N]](nodes []N, height int) Model[N] {
	m := Model[N]{Nodes: nodes, Height: height}
	vis := m.Visible()
	for m.Cursor < len(vis) && nodes[vis[m.Cursor]].TreeInfo().Section {
		m.Cursor++
	}
	return m
}

func (m Model[N]) WithKeys(k Keys) Model[N] {
	m.keys = &k
	return m
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
	var vis []int
	var collapsed []int // depths of collapsed ancestors: inside one, rows are skipped
	for i, n := range m.Nodes {
		in := n.TreeInfo()
		for len(collapsed) > 0 && in.Depth <= collapsed[len(collapsed)-1] {
			collapsed = collapsed[:len(collapsed)-1]
		}
		if len(collapsed) > 0 {
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
func (m Model[N]) CurrentIdx() int { return At(m.Cursor, m.Visible()) }

// At maps a cursor position into a visible index list, or -1 when out of
// range. For callers that already hold the Visible() result.
func At(cursor int, vis []int) int {
	if cursor >= 0 && cursor < len(vis) {
		return vis[cursor]
	}
	return -1
}

// WithNodeMutated returns the tree with a fresh Nodes slice in which
// nodes[idx] was changed by mut: copy-on-write, never in place.
func (m Model[N]) WithNodeMutated(idx int, mut func(*N)) Model[N] {
	if idx < 0 || idx >= len(m.Nodes) {
		return m
	}
	nodes := make([]N, len(m.Nodes))
	copy(nodes, m.Nodes)
	mut(&nodes[idx])
	m.Nodes = nodes
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
	if m.Cursor >= len(vis) {
		m.Cursor = len(vis) - 1
	}
	start := m.Cursor
	for next := m.Cursor + dir; next >= 0 && next < len(vis); next += dir {
		m.Cursor = next
		if !m.Nodes[vis[next]].TreeInfo().Section {
			break
		}
	}
	if m.Nodes[vis[m.Cursor]].TreeInfo().Section {
		m.Cursor = start
	}
	return m.ClampOffset()
}

// Move walks delta rows, stepping over section rows and stopping at the
// last row rather than running into a trailing heading.
func (m Model[N]) Move(delta int) Model[N] {
	vis := m.Visible()
	if len(vis) == 0 {
		return m
	}
	dir := 1
	if delta < 0 {
		dir, delta = -1, -delta
	}
	for ; delta > 0; delta-- {
		next := m.Cursor + dir
		for next >= 0 && next < len(vis) && m.Nodes[vis[next]].TreeInfo().Section {
			next += dir
		}
		if next < 0 || next >= len(vis) {
			break
		}
		m.Cursor = next
	}
	return m.ClampOffset()
}

// MoveTo jumps to a visible row, snapping backwards over section rows: the
// last row of a grouped list is a node only by coincidence.
func (m Model[N]) MoveTo(i int) Model[N] {
	vis := m.Visible()
	if len(vis) == 0 {
		return m
	}
	i = min(max(i, 0), len(vis)-1)
	for i > 0 && m.Nodes[vis[i]].TreeInfo().Section {
		i--
	}
	for i < len(vis)-1 && m.Nodes[vis[i]].TreeInfo().Section {
		i++
	}
	m.Cursor = i
	return m.ClampOffset()
}

// Expand opens the row under the cursor when it is a collapsed parent.
func (m Model[N]) Expand() (Model[N], Action) { return m.expand(m.Visible()) }

func (m Model[N]) expand(vis []int) (Model[N], Action) {
	idx := At(m.Cursor, vis)
	if idx < 0 {
		return m, None
	}
	if in := m.Nodes[idx].TreeInfo(); !in.Leaf && !in.Expanded {
		return m.WithNodeMutated(idx, func(n *N) { *n = (*n).Expand(true) }), Expanded
	}
	return m, None
}

// collapse closes an expanded parent, or moves the cursor to the parent of
// a nested row.
func (m Model[N]) collapse(vis []int) (Model[N], Action) {
	idx := At(m.Cursor, vis)
	if idx < 0 {
		return m, None
	}
	in := m.Nodes[idx].TreeInfo()
	if !in.Leaf && in.Expanded {
		return m.WithNodeMutated(idx, func(n *N) { *n = (*n).Expand(false) }), Collapsed
	}
	if in.Depth > 0 {
		for vi := m.Cursor - 1; vi >= 0; vi-- {
			if m.Nodes[vis[vi]].TreeInfo().Depth == in.Depth-1 {
				m.Cursor = vi
				return m.ClampOffset(), None
			}
		}
	}
	return m, None
}

// ClampOffset scrolls so the cursor row stays visible.
func (m Model[N]) ClampOffset() Model[N] {
	m.Offset = layout.ClampScroll(m.Cursor, m.Offset, m.Height)
	// The last row becomes "↓ N more" on overflow, so a cursor landing there
	// needs one more line of offset to stay visible.
	if m.hasMore(len(m.Visible())) && m.Cursor >= m.Offset+m.Height-1 {
		m.Offset = m.Cursor - m.Height + 2
	}
	return m
}

// hasMore reports whether n visible rows overflow below the window. A one-row
// window never trades its only row for the indicator.
func (m Model[N]) hasMore(n int) bool {
	return m.Height > 1 && m.Offset+m.Height < n
}

// ClampCursor forces the cursor back into the visible range, for state
// restores whose cursor may no longer be valid against the restored nodes.
func (m Model[N]) ClampCursor() Model[N] {
	vis := m.Visible()
	m.Cursor = min(max(m.Cursor, 0), max(len(vis)-1, 0))
	return m
}

// CursorTo moves the cursor to the first visible row match accepts, or
// leaves it when none does. Reports whether it moved.
func (m Model[N]) CursorTo(match func(N) bool) (Model[N], bool) {
	for vi, ni := range m.Visible() {
		if match(m.Nodes[ni]) {
			m.Cursor = vi
			return m.ClampOffset(), true
		}
	}
	return m, false
}

// Reveal is CursorTo for a row that may sit under collapsed parents: the
// ancestors of the first match in Nodes are expanded first. Ancestors are
// the nearer rows above it with a smaller depth.
func (m Model[N]) Reveal(match func(N) bool) (Model[N], bool) {
	if m2, ok := m.CursorTo(match); ok {
		return m2, true
	}
	target := -1
	for i, n := range m.Nodes {
		if match(n) {
			target = i
			break
		}
	}
	if target < 0 {
		return m, false
	}
	nodes := make([]N, len(m.Nodes))
	copy(nodes, m.Nodes)
	depth := nodes[target].TreeInfo().Depth
	for i := target - 1; i >= 0 && depth > 0; i-- {
		if in := nodes[i].TreeInfo(); in.Depth < depth {
			if !in.Leaf && !in.Expanded {
				nodes[i] = nodes[i].Expand(true)
			}
			depth = in.Depth
		}
	}
	m.Nodes = nodes
	return m.CursorTo(match)
}

// View renders the visible rows through row, which draws one node knowing
// whether it is under the cursor. The last row becomes "↓ N more" when rows
// overflow; short trees are padded to Height. A tree not yet sized
// (Height <= 0) renders nothing.
func (m Model[N]) View(th theme.Resolved, row func(n N, idx int, selected bool) string) string {
	if m.Height <= 0 {
		return ""
	}
	vis := m.Visible()
	if len(vis) == 0 {
		msg := m.EmptyMessage
		if msg == "" {
			msg = "  (empty)"
		}
		return th.Dim.Render(msg)
	}
	// Nodes may have shrunk under a scrolled tree; never scroll past the end.
	m = m.ClampCursor()
	m.Offset = min(m.Offset, max(len(vis)-m.Height, 0))
	m = m.ClampOffset()
	visible := m.Height
	more := m.hasMore(len(vis))
	if more {
		visible--
	}
	end := min(m.Offset+visible, len(vis))

	var sb strings.Builder
	for vi := m.Offset; vi < end; vi++ {
		sb.WriteString(row(m.Nodes[vis[vi]], vis[vi], vi == m.Cursor))
		sb.WriteByte('\n')
	}
	if more {
		sb.WriteString(th.Dim.Render(fmt.Sprintf("  ↓ %d more", len(vis)-end)))
	} else {
		for i := end - m.Offset; i < m.Height; i++ {
			sb.WriteByte('\n')
		}
	}
	return strings.TrimSuffix(sb.String(), "\n")
}
