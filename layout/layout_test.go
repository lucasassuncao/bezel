package layout

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func names(n Node) []string {
	var out []string
	for l := range Leaves(n) {
		out = append(out, l.Name)
	}
	return out
}

func TestLeavesWalkInDeclarationOrder(t *testing.T) {
	tree := Columns(
		Fixed("list", Ratio(1, 3)),
		Rows(Fill("detail"), Fixed("copy", Ratio(1, 5))),
	)
	require.Equal(t, []string{"list", "detail", "copy"}, names(tree))
}

func TestInfoMarksALeaf(t *testing.T) {
	require.True(t, Fixed("hint", Lines(8)).Info().IsInfo())
	require.False(t, Fill("detail").IsInfo())
}

func TestLeavesYieldsDepthFirstAndStopsWhenAsked(t *testing.T) {
	tree := Columns(Fixed("list"), Rows(Fill("detail"), Fixed("hint").Info()), Fill("log"))

	var names []string
	for l := range Leaves(tree) {
		names = append(names, l.Name)
		if l.Name == "hint" {
			break
		}
	}
	require.Equal(t, []string{"list", "detail", "hint"}, names)
}

func TestCollapseIsRecordedOnTheBox(t *testing.T) {
	b := Columns(Fixed("a"), Fill("b")).Collapse(72, "b")
	require.Equal(t, 72, b.collapseLimit)
	require.Equal(t, "b", b.collapseKeep)
}

// A box shared by two layouts, as SetLayout swaps them, must not change under either.
func TestCollapseLeavesTheReceiverAsItWas(t *testing.T) {
	base := Columns(Fixed("a", Lines(10)), Fill("b"))
	narrow := base.Collapse(72, "b")

	require.Zero(t, base.collapseLimit)
	require.Empty(t, base.collapseKeep)
	require.Equal(t, 72, narrow.collapseLimit)

	rects, collapsed := Resolve(base, Rect{W: 40, H: 5}, nil)
	require.False(t, collapsed)
	require.Equal(t, 10, rects["a"].W)
}

// The tree keeps what it was built from, whatever the caller does after.
func TestConstructorsCopyTheirArguments(t *testing.T) {
	cs := []Constraint{Lines(3)}
	children := []Node{Fixed("a", cs...), Fill("b")}
	tree := Rows(children...)
	cs[0], children[1] = Lines(9), Fill("c")

	rects, _ := Resolve(tree, Rect{W: 10, H: 10}, nil)
	require.Equal(t, 3, rects["a"].H)
	require.Equal(t, []string{"a", "b"}, names(tree))
}
