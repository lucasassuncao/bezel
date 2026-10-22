package viewport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnUnwrappedIndexIsARowALine(t *testing.T) {
	x := newIndex([]string{"ab", "", "abcdef"})
	require.Equal(t, 3, x.total())
	require.Equal(t, 6, x.longest)
	line, part := x.at(2)
	require.Equal(t, []int{2, 0}, []int{line, part})
}

func TestAWrappedIndexCountsEveryRow(t *testing.T) {
	x := newIndex([]string{"ab", "", "abcdef"}).wrapped(4)
	require.Equal(t, 4, x.total(), "an empty line still takes a row")
	require.Equal(t, 2, x.row(2))
	line, part := x.at(3)
	require.Equal(t, []int{2, 1}, []int{line, part})
	require.Equal(t, 3, x.unwrapped().total())
}

func TestAtPastTheEndIsTheLineCount(t *testing.T) {
	line, _ := newIndex([]string{"a"}).wrapped(4).at(9)
	require.Equal(t, 1, line)
}

func TestWideGraphemesCountTheirCells(t *testing.T) {
	require.Equal(t, 2, newIndex([]string{"🌍🌍"}).wrapped(2).total())
}

func TestAWideGraphemeIsNeverSplitAcrossRows(t *testing.T) {
	x := newIndex([]string{"全角全"}).wrapped(3)
	require.Equal(t, 3, x.total(), "each two-cell character takes a row of three")
	require.Equal(t, []int{0, 2, 4}, breaks("全角全", 3))
	require.Equal(t, []int{0, 3, 6}, breaks("abcdefgh", 3))
}
