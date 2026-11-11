package canvas

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLineRunsBetweenItsEnds(t *testing.T) {
	m := grid().Draw(Line{X0: 0, Y0: 7, X1: 7, Y1: 0})
	require.Equal(t, []string{"⠑⢄  ", "  ⠑⢄"}, plain(m))
}

// A line from far away costs no more than one inside, and keeps its slope.
func TestALineFromFarAwayIsCutAtTheEdge(t *testing.T) {
	m := grid().Draw(Line{X0: -1e12, Y0: 0, X1: 1e12, Y1: 0})
	require.Equal(t, []string{"    ", "⣀⣀⣀⣀"}, plain(m))
	m = grid().Draw(Line{X0: -7, Y0: 14, X1: 14, Y1: -7}) // the same diagonal, longer
	require.Equal(t, plain(grid().Draw(Line{X0: 0, Y0: 7, X1: 7, Y1: 0})), plain(m))
}

func TestABrokenLineIsSkipped(t *testing.T) {
	m := grid().Draw(Line{X0: math.NaN(), X1: 7}, Line{X1: math.Inf(1)})
	require.Equal(t, []string{"    ", "    "}, plain(m))
}

func TestRectIsAnOutline(t *testing.T) {
	m := grid().Draw(Rect{X: 0, Y: 0, W: 7, H: 7})
	require.Equal(t, []string{"⡏⠉⠉⢹", "⣇⣀⣀⣸"}, plain(m))
}

func TestCircleIsClosedAndCentred(t *testing.T) {
	m := New(-10, 10, -10, 10).SetWidth(11).SetHeight(6).Draw(Circle{R: 9})
	rows := plain(m)
	require.Len(t, rows, 6)
	for _, r := range rows {
		require.NotEqual(t, strings.Repeat(" ", 11), r, "every row crosses the outline")
		require.Len(t, []rune(r), 11)
	}
	// Left and right halves mirror each other in how many cells they light.
	left, right := 0, 0
	for _, r := range rows {
		rs := []rune(r)
		for i, c := range rs {
			if c != ' ' && i < 5 {
				left++
			}
			if c != ' ' && i > 5 {
				right++
			}
		}
	}
	require.Equal(t, left, right)
}

func TestABadCircleIsSkipped(t *testing.T) {
	m := grid().Draw(Circle{X: 3, Y: 3, R: 0}, Circle{X: 3, Y: 3, R: -1}, Circle{X: math.NaN(), R: 1})
	require.Equal(t, []string{"    ", "    "}, plain(m))
}

func TestPointsOutsideAreDropped(t *testing.T) {
	m := grid().Draw(Points{Coords: []Point{{-1, 0}, {8, 0}, {0, 8}, {math.NaN(), 0}, {0, 0}}})
	require.Equal(t, []string{"    ", "⡀   "}, plain(m))
}

func TestLabelWritesOverTheDotsFromItsCell(t *testing.T) {
	m := grid().Draw(Line{X0: 0, Y0: 0, X1: 7, Y1: 0}, Label{X: 2, Y: 0, Text: "\x1b[1mhi\x1b[0m there"})
	require.Equal(t, []string{"    ", "⣀hi "}, plain(m))
}

func TestALabelOutsideIsDropped(t *testing.T) {
	m := grid().Draw(Label{X: -1, Y: 0, Text: "x"}, Label{X: 0, Y: 9, Text: "y"}, Label{X: math.NaN(), Text: "z"})
	require.Equal(t, []string{"    ", "    "}, plain(m))
}
