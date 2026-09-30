package layout

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderWalksLeavesInDeclarationOrder(t *testing.T) {
	tree := Columns(
		Fixed("list", Ratio(1, 3)),
		Rows(Fill("detail"), Fixed("copy", Ratio(1, 5))),
	)
	require.Equal(t, []string{"list", "detail", "copy"}, Order(tree))
}

func TestRingLeavesOutInfoLeaves(t *testing.T) {
	tree := Columns(
		Fixed("list", Ratio(1, 3)),
		Rows(Fill("detail"), Fixed("hint", Lines(8)).Info()),
	)
	require.Equal(t, []string{"list", "detail", "hint"}, Order(tree))
	require.Equal(t, []string{"list", "detail"}, Ring(tree))
}

func TestCollapseIsRecordedOnTheBox(t *testing.T) {
	b := Columns(Fixed("a"), Fill("b")).Collapse(72, "b")
	require.Equal(t, 72, b.collapseLimit)
	require.Equal(t, "b", b.collapseKeep)
}
