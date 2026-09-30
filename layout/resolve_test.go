package layout

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func all(string) bool { return true }

func TestTwoColumnsSplitByRatioWithClamps(t *testing.T) {
	tree := Columns(Fixed("list", Ratio(1, 3), Min(26), Max(52)), Fill("detail"))

	rects, collapsed := Resolve(tree, Rect{0, 0, 120, 20}, all)
	require.False(t, collapsed)
	require.Equal(t, Rect{0, 0, 40, 20}, rects["list"])
	require.Equal(t, Rect{40, 0, 80, 20}, rects["detail"])

	rects, _ = Resolve(tree, Rect{0, 0, 60, 20}, all) // a third is 20: floored at 26
	require.Equal(t, 26, rects["list"].W)
	require.Equal(t, 34, rects["detail"].W)

	rects, _ = Resolve(tree, Rect{0, 0, 300, 20}, all) // a third is 100: capped at 52
	require.Equal(t, 52, rects["list"].W)
	require.Equal(t, 248, rects["detail"].W)
}

func TestNestedRowsInsideAColumn(t *testing.T) {
	tree := Columns(
		Fixed("list", Ratio(1, 3)),
		Rows(Fill("detail"), Fixed("copy", Lines(6))),
	)
	rects, _ := Resolve(tree, Rect{0, 0, 90, 30}, all)
	require.Equal(t, Rect{0, 0, 30, 30}, rects["list"])
	require.Equal(t, Rect{30, 0, 60, 24}, rects["detail"])
	require.Equal(t, Rect{30, 24, 60, 6}, rects["copy"])
}

func TestAbsentLeavesGiveTheirSpaceBack(t *testing.T) {
	tree := Columns(Fixed("list", Ratio(1, 3)), Rows(Fill("detail"), Fixed("copy", Lines(6))))
	present := func(name string) bool { return name != "copy" }

	rects, _ := Resolve(tree, Rect{0, 0, 90, 30}, present)
	_, ok := rects["copy"]
	require.False(t, ok)
	require.Equal(t, 30, rects["detail"].H)
}

func TestCollapseKeepsOneLeafBelowTheLimit(t *testing.T) {
	tree := Columns(Fixed("list", Ratio(1, 3), Min(26)), Fill("detail")).Collapse(72, "detail")

	rects, collapsed := Resolve(tree, Rect{0, 0, 60, 20}, all)
	require.True(t, collapsed)
	require.Equal(t, Rect{0, 0, 60, 20}, rects["detail"])
	_, ok := rects["list"]
	require.False(t, ok)
}

func TestMinimumsThatDoNotFitShrinkProportionally(t *testing.T) {
	tree := Columns(Fixed("a", Lines(30)), Fixed("b", Lines(30)), Fill("c"))
	rects, _ := Resolve(tree, Rect{0, 0, 31, 5}, all)

	total := rects["a"].W + rects["b"].W + rects["c"].W
	require.Equal(t, 31, total)
	require.GreaterOrEqual(t, rects["c"].W, 1)
	require.Equal(t, rects["a"].X+rects["a"].W, rects["b"].X)
	require.Equal(t, rects["b"].X+rects["b"].W, rects["c"].X)
}

func TestSeveralFillsShareEvenly(t *testing.T) {
	tree := Rows(Fill("a"), Fill("b"), Fill("c"))
	rects, _ := Resolve(tree, Rect{0, 0, 10, 10}, all)
	require.Equal(t, 3, rects["a"].H)
	require.Equal(t, 3, rects["b"].H)
	require.Equal(t, 4, rects["c"].H) // remainder goes to the last
	require.Equal(t, 6, rects["c"].Y)
}

func TestResolveNeverOverflowsOrGoesNegative(t *testing.T) {
	tree := Columns(
		Fixed("nav", Ratio(1, 5), Min(20)),
		Fixed("list", Ratio(1, 3), Min(26), Max(52)),
		Rows(Fill("detail"), Fixed("copy", Lines(6))),
	).Collapse(72, "detail")

	for w := 0; w <= 200; w += 7 {
		for h := 0; h <= 60; h += 5 {
			rects, _ := Resolve(tree, Rect{0, 0, w, h}, all)
			for name, r := range rects {
				require.GreaterOrEqual(t, r.W, 0, "%s at %dx%d", name, w, h)
				require.GreaterOrEqual(t, r.H, 0, "%s at %dx%d", name, w, h)
				require.LessOrEqual(t, r.X+r.W, w, "%s at %dx%d", name, w, h)
				require.LessOrEqual(t, r.Y+r.H, h, "%s at %dx%d", name, w, h)
			}
		}
	}
}
