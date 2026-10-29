package tree

import (
	"charm.land/lipgloss/v2"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

type node struct {
	label    string
	depth    int
	leaf     bool
	expanded bool
	section  bool
	hidden   bool
}

func (n node) TreeInfo() Info {
	return Info{Depth: n.depth, Leaf: n.leaf, Expanded: n.expanded, Section: n.section, Hidden: n.hidden}
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

func newTree(h int) Model[node] { return New(nodes()).SetHeight(h) }

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
	th := theme.Resolve(theme.Plain, true)
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

	empty := New[node](nil).SetHeight(5)
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
	m := New(leaves(3))
	require.Empty(t, m.View(theme.Resolved{}, labelRow))
}

func TestCursorNeverHidesBehindMoreIndicator(t *testing.T) {
	m := New(leaves(5)).SetHeight(3)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Contains(t, m.View(theme.Resolved{}, labelRow), "> c")
}

func TestScrollIsWhereTheWindowSits(t *testing.T) {
	m := New(leaves(5)).SetHeight(3)
	require.Equal(t, layout.Scroll{Visible: 2, Total: 5}, m.Scroll(), "the more row is not a row shown")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, layout.Scroll{Offset: 1, Visible: 2, Total: 5}, m.Scroll())
	require.Equal(t, layout.Scroll{}, New[node](nil).SetHeight(5).Scroll())
}

func TestOneRowTreeShowsTheCursorNotTheIndicator(t *testing.T) {
	m := New(leaves(3)).SetHeight(1)
	require.Equal(t, "> a", m.View(theme.Resolved{}, labelRow))
}

// Nodes is exported, so an app can shrink it under a scrolled tree.
func TestViewStaysHeightTallAfterNodesShrink(t *testing.T) {
	m := New(leaves(20)).SetHeight(5).MoveTo(15)
	m.Nodes = leaves(3)
	require.Len(t, strings.Split(m.View(theme.Resolved{}, labelRow), "\n"), 5)
}

// Cursor is exported, so a caller can leave it out of range.
func TestOutOfRangeCursorDoesNotPanic(t *testing.T) {
	for _, c := range []int{-3, 99} {
		m := newTree(10)
		m.Cursor = c
		require.NotPanics(t, func() {
			m, _ = m.Update(press(tea.KeyDown))
			m = m.Move(-1)
		})
	}
}

// bigTree holds parents expanded rows of 9 leaves each.
func bigTree(parents, height int) Model[node] {
	nodes := make([]node, 0, parents*10)
	for i := range parents {
		nodes = append(nodes, node{label: fmt.Sprint("p", i), expanded: true})
		for j := range 9 {
			nodes = append(nodes, node{label: fmt.Sprint("l", j), depth: 1, leaf: true})
		}
	}
	return New(nodes).SetHeight(height)
}

var benchSizes = []int{100, 10_000}

func BenchmarkUpdateDown(b *testing.B) {
	down := press(tea.KeyDown)
	for _, n := range benchSizes {
		b.Run(fmt.Sprint(n*10, "nodes"), func(b *testing.B) {
			m := bigTree(n, 40)
			for b.Loop() {
				m, _ = m.Update(down)
			}
		})
	}
}

func BenchmarkView(b *testing.B) {
	th := theme.Resolve(theme.Plain, true)
	row := func(n node, _ int, _ bool) string { return n.label }
	for _, n := range benchSizes {
		b.Run(fmt.Sprint(n*10, "nodes"), func(b *testing.B) {
			m := bigTree(n, 40)
			for b.Loop() {
				_ = m.View(th, row)
			}
		})
	}
}

// A nil func matches nothing, changes nothing and draws empty rows.
func TestNilFuncsAreHarmless(t *testing.T) {
	m := newTree(4)
	require.Equal(t, m.Nodes, m.WithNodeMutated(0, nil).Nodes)
	_, moved := m.CursorTo(nil)
	require.False(t, moved)
	_, moved = m.Reveal(nil)
	require.False(t, moved)
	require.NotPanics(t, func() { _ = m.View(theme.Resolved{}, nil) })
}

func TestAHiddenNodeTakesItsBranchWithIt(t *testing.T) {
	ns := nodes()
	ns[1].hidden = true // server, open, with host and port under it
	m := New(ns).SetHeight(10)
	var labels []string
	for _, i := range m.Visible() {
		labels = append(labels, m.Nodes[i].label)
	}
	require.Equal(t, []string{"ADDED", "AVAILABLE", "cache"}, labels)
	require.Equal(t, "cache", label(m), "the cursor skips what is gone")
}

func TestGuidesDrawTheBranches(t *testing.T) {
	ns := []node{
		{label: "root", expanded: true},
		{label: "a", depth: 1, expanded: true},
		{label: "a1", depth: 2, leaf: true},
		{label: "a2", depth: 2, leaf: true},
		{label: "b", depth: 1, expanded: true},
		{label: "b1", depth: 2, leaf: true},
	}
	m := New(ns).SetHeight(10)
	require.Equal(t, []string{"", "├─ ", "│  ├─ ", "│  └─ ", "└─ ", "   └─ "}, m.Guides())
	m = m.WithConnectors(Connectors{Mid: "+ ", Last: "` ", Pipe: "| ", Space: "  "})
	require.Equal(t, "| + ", m.Guides()[2])
}

func TestScrollOffKeepsRowsAroundTheCursor(t *testing.T) {
	ns := make([]node, 30)
	for i := range ns {
		ns[i] = node{label: fmt.Sprint(i), leaf: true}
	}
	m := New(ns).SetHeight(10).WithScrollOff(2)
	for range 7 {
		m, _ = m.Update(press(tea.KeyDown))
	}
	require.Equal(t, 7, m.Cursor)
	require.Equal(t, 1, m.Offset, "two rows and the more row below the cursor")
	for range 6 {
		m, _ = m.Update(press(tea.KeyUp))
	}
	require.Equal(t, 1, m.Cursor)
	require.Equal(t, 0, m.Offset, "the top can show no more than it has")
}

func TestTheTreeLegendNamesItsKeys(t *testing.T) {
	var keys []string
	for _, e := range newTree(5).Legend() {
		keys = append(keys, e.Help().Key)
	}
	require.Equal(t, []string{"↑/↓", "←/→"}, keys)
}

func TestAClickMovesTheCursorAndASecondOpensOrCloses(t *testing.T) {
	m := newTree(10) // ADDED, server (open), host, port, AVAILABLE, cache, ...
	m, act := m.Click(5)
	require.Equal(t, None, act)
	require.Equal(t, "cache", label(m))
	m, act = m.Click(5)
	require.Equal(t, Expanded, act)
	require.Contains(t, labelsOf(m), "ttl")
	m, act = m.Click(5)
	require.Equal(t, Collapsed, act)

	m, _ = m.Click(2)
	_, act = m.Click(2)
	require.Equal(t, None, act, "a leaf has nothing to open")
	same, _ := m.Click(0)
	require.Equal(t, "host", label(same), "a heading is not a row to land on")
}

func labelsOf(m Model[node]) []string {
	var out []string
	for _, i := range m.Visible() {
		out = append(out, m.Nodes[i].label)
	}
	return out
}

func TestWithStylesRecoloursTheEmptyMessage(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New([]node{}).SetHeight(3).WithStyles(func(s Styles) Styles { s.Empty = red; return s })
	require.Equal(t, red.Render("  (empty)"), m.View(theme.Resolve(theme.Mint, true), nil))
}
