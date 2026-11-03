package layout

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var body = Rect{0, 0, 100, 20}

func twoSplits() Box {
	return Columns(Fixed("list", Ratio(1, 4)), Rows(Fill("detail"), Fixed("log", Lines(5))).Resizable()).Resizable()
}

func TestEdgesSitWhereTheChildrenMeet(t *testing.T) {
	edges := Sizes(nil).Edges(twoSplits(), body, all)
	require.Equal(t, []Edge{
		{Box: "list", Index: 0, Cols: true, At: 25, From: 0, To: 20},
		{Box: "detail", Index: 0, Cols: false, At: 15, From: 25, To: 100},
	}, edges, "the outer box first, each keyed by its first leaf")
	require.Empty(t, Sizes(nil).Edges(Columns(Fill("a"), Fill("b")), body, all), "only Resizable boxes")
}

func TestMoveShiftsTheCellsAndResolveScalesThem(t *testing.T) {
	tree := twoSplits()
	e := Sizes(nil).Edges(tree, body, all)[0]
	sz := Sizes(nil).Move(tree, body, all, e, 15)
	rects, _ := sz.Resolve(tree, body, all)
	require.Equal(t, 40, rects["list"].W)
	require.Equal(t, Rect{40, 0, 60, 15}, rects["detail"])

	rects, _ = sz.Resolve(tree, Rect{0, 0, 200, 20}, all)
	require.Equal(t, 80, rects["list"].W, "the share holds through a resize")
	rects, _ = Resolve(tree, body, all)
	require.Equal(t, 25, rects["list"].W, "Resolve keeps the declared sizes")
}

func TestMoveIsHeldByTheFloorAndTheConstraints(t *testing.T) {
	tree := twoSplits()
	e := Sizes(nil).Edges(tree, body, all)[0]
	rects, _ := Sizes(nil).Move(tree, body, all, e, -50).Resolve(tree, body, all)
	require.Equal(t, 3, rects["list"].W, "three cells: two borders and one inside")
	rects, _ = Sizes(nil).Move(tree, body, all, e, 500).Resolve(tree, body, all)
	require.Equal(t, 97, rects["list"].W)

	capped := Columns(Fixed("list", Ratio(1, 4), Min(10), Max(30)), Fill("detail")).Resizable()
	e = Sizes(nil).Edges(capped, body, all)[0]
	rects, _ = Sizes(nil).Move(capped, body, all, e, 50).Resolve(capped, body, all)
	require.Equal(t, 30, rects["list"].W, "Max caps the drag")
	rects, _ = Sizes(nil).Move(capped, body, all, e, -50).Resolve(capped, body, all)
	require.Equal(t, 10, rects["list"].W, "Min floors it")

	before := Sizes{}
	_ = before.Move(tree, body, all, e, 5)
	require.Empty(t, before, "Move leaves sz as it was")
}

func TestGrowMovesTheEdgeAfterThePaneOrBeforeTheLast(t *testing.T) {
	tree := twoSplits()
	rects, _ := Sizes(nil).Grow(tree, body, all, "list", true, 4).Resolve(tree, body, all)
	require.Equal(t, 29, rects["list"].W)
	rects, _ = Sizes(nil).Grow(tree, body, all, "log", true, 4).Resolve(tree, body, all)
	require.Equal(t, 21, rects["list"].W, "the detail column is last: its left edge moves")
	rects, _ = Sizes(nil).Grow(tree, body, all, "log", false, 2).Resolve(tree, body, all)
	require.Equal(t, 7, rects["log"].H, "the inner box for rows")
	require.Nil(t, Sizes(nil).Grow(Columns(Fill("a"), Fill("b")), body, all, "a", true, 2))
}

func TestSizesForAnotherShapeAreIgnored(t *testing.T) {
	tree := twoSplits()
	sz := Sizes{"list": {1, 2, 3}}
	rects, _ := sz.Resolve(tree, body, all)
	require.Equal(t, 25, rects["list"].W, "three sizes for two children")
	rects, _ = Sizes{"list": {50, 50}}.Resolve(tree, body, func(n string) bool { return n != "list" })
	require.Equal(t, 100, rects["detail"].W, "without list the box is keyed by detail")
}
