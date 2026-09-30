package tree

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

type node struct {
	label    string
	depth    int
	leaf     bool
	expanded bool
	section  bool
}

func (n node) TreeInfo() Info {
	return Info{Depth: n.depth, Leaf: n.leaf, Expanded: n.expanded, Section: n.section}
}

func (n node) Expand(open bool) node {
	n.expanded = open
	return n
}

// ADDED / server / ├ host / └ port / AVAILABLE / cache (collapsed) / └ ttl
func nodes() []node {
	return []node{
		{label: "ADDED", section: true},
		{label: "server", expanded: true},
		{label: "host", depth: 1, leaf: true},
		{label: "port", depth: 1, leaf: true},
		{label: "AVAILABLE", section: true},
		{label: "cache"},
		{label: "ttl", depth: 1, leaf: true},
	}
}

func newTree(h int) Model[node] { return New(nodes(), h) }

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func label(m Model[node]) string { return m.Nodes[m.CurrentIdx()].label }

func TestVisibleHidesCollapsedChildrenAndCursorSkipsSections(t *testing.T) {
	m := newTree(10)
	require.Equal(t, []int{0, 1, 2, 3, 4, 5}, m.Visible(), "ttl is under a collapsed parent")
	require.Equal(t, "server", label(m), "cursor starts past the heading")

	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown)) // hops AVAILABLE
	require.Equal(t, "cache", label(m))
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "cache", label(m), "clamped at the end")
	m, _ = m.Update(press(tea.KeyUp))
	m, _ = m.Update(press(tea.KeyUp))
	m, _ = m.Update(press(tea.KeyUp))
	m, _ = m.Update(press(tea.KeyUp))
	require.Equal(t, "server", label(m), "clamped above the first heading")
}

func TestRightExpandsLeftCollapsesOrGoesToParent(t *testing.T) {
	m := newTree(10)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "cache", label(m))

	m, act := m.Update(press(tea.KeyRight))
	require.Equal(t, Expanded, act)
	require.Len(t, m.Visible(), 7)
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "ttl", label(m))

	m, act = m.Update(press(tea.KeyLeft))
	require.Equal(t, None, act)
	require.Equal(t, "cache", label(m), "left on a child goes to its parent")
	m, act = m.Update(press(tea.KeyLeft))
	require.Equal(t, Collapsed, act)
	require.Len(t, m.Visible(), 6)

	_, act = m.Update(press(tea.KeyRight))
	require.Equal(t, Expanded, act)
	m, _ = m.Update(press(tea.KeyUp)) // "port", a leaf
	_, act = m.Update(press(tea.KeyRight))
	require.Equal(t, None, act, "a leaf does not expand")
}

func TestMutationsAreCopyOnWrite(t *testing.T) {
	m := newTree(10)
	before := m.Nodes
	m2 := m.WithNodeMutated(1, func(n *node) { n.expanded = false })
	require.True(t, before[1].expanded, "the old slice is untouched")
	require.False(t, m2.Nodes[1].expanded)
	require.Equal(t, m.Nodes, m.WithNodeMutated(99, func(n *node) { n.label = "x" }).Nodes, "stale index is a no-op")
}

func TestRevealExpandsAncestorsAndCursorToMatches(t *testing.T) {
	m := newTree(10)
	m, ok := m.Reveal(func(n node) bool { return n.label == "ttl" })
	require.True(t, ok)
	require.Equal(t, "ttl", label(m))
	require.True(t, m.Nodes[5].expanded)

	m, ok = m.CursorTo(func(n node) bool { return n.label == "host" })
	require.True(t, ok)
	require.Equal(t, "host", label(m))
	_, ok = m.CursorTo(func(n node) bool { return n.label == "nope" })
	require.False(t, ok)
}

func TestClampCursorAfterNodesShrink(t *testing.T) {
	m := newTree(10)
	m.Cursor = 5
	m.Nodes = m.Nodes[:2]
	require.Equal(t, 1, m.ClampCursor().Cursor)
	m.Nodes = nil
	require.Equal(t, 0, m.ClampCursor().Cursor)
}

func TestViewPadsShortTreesAndMarksOverflow(t *testing.T) {
	th := theme.Resolve(theme.ThemePlain, true)
	row := func(n node, _ int, selected bool) string {
		prefix := "  "
		if selected {
			prefix = "▶ "
		}
		return prefix + strings.Repeat("  ", n.depth) + n.label
	}
	m := newTree(3)
	got := ansi.Strip(m.View(th, row))
	lines := strings.Split(got, "\n")
	require.Len(t, lines, 3)
	require.Equal(t, "▶ server", lines[1])
	require.Contains(t, lines[2], "↓ 4 more")

	tall := newTree(10)
	require.Len(t, strings.Split(tall.View(th, row), "\n"), 10, "padded to the height")
	require.Contains(t, ansi.Strip(tall.View(th, row)), "    host")

	empty := New[node](nil, 5)
	empty.EmptyMessage = "  (no fields)"
	require.Contains(t, empty.View(th, row), "(no fields)")
}

func TestMoveStepsOverSectionsAndMoveToSnaps(t *testing.T) {
	m := newTree(10)
	m = m.Move(3) // server, host, port, (AVAILABLE skipped) cache
	require.Equal(t, "cache", label(m))
	m = m.Move(-10)
	require.Equal(t, "server", label(m), "clamped above the first heading")
	m = m.MoveTo(4) // AVAILABLE: snaps back onto port
	require.Equal(t, "port", label(m))
	m = m.MoveTo(99)
	require.Equal(t, "cache", label(m))
}

func leaves(n int) []node {
	out := make([]node, n)
	for i := range out {
		out[i] = node{label: string(rune('a' + i)), leaf: true}
	}
	return out
}

func labelRow(n node, _ int, selected bool) string {
	if selected {
		return "> " + n.label
	}
	return "  " + n.label
}

func TestViewIsEmptyBeforeSized(t *testing.T) {
	m := New(leaves(3), 0)
	require.Empty(t, m.View(theme.Resolved{}, labelRow))
}

func TestCursorNeverHidesBehindMoreIndicator(t *testing.T) {
	m := New(leaves(5), 3)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Contains(t, m.View(theme.Resolved{}, labelRow), "> c")
}

func TestOneRowTreeShowsTheCursorNotTheIndicator(t *testing.T) {
	m := New(leaves(3), 1)
	require.Equal(t, "> a", m.View(theme.Resolved{}, labelRow))
}

// Nodes is exported, so an app can shrink it under a scrolled tree.
func TestViewStaysHeightTallAfterNodesShrink(t *testing.T) {
	m := New(leaves(20), 5).MoveTo(15)
	m.Nodes = leaves(3)
	require.Len(t, strings.Split(m.View(theme.Resolved{}, labelRow), "\n"), 5)
}
