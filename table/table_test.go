package table

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

var cols = []Column{
	{Title: "Name", Min: 4, Max: 20, Flex: true},
	{Title: "Source", Max: 10},
	{Title: "Version"},
}

var rows = [][]string{
	{"Google Chrome", "winget", "126.0.6478.127"},
	{"VS Code", "scoop", "1.91"},
}

func TestFitSizesToContentWithinBounds(t *testing.T) {
	w := Fit(cols, rows, 200)
	require.Equal(t, []int{13, 6, 14}, w) // widest cell wins, "Version" title shorter than the value
	require.Equal(t, 10, Fit(cols, [][]string{{"x", "a-very-long-source", "1"}}, 200)[1], "capped at Max")
	require.Equal(t, 4, Fit(cols, [][]string{{"x", "s", "1"}}, 200)[0], "a short column stays the width of its title")
}

func TestFitShrinksTheFlexColumnFirstThenTheWidest(t *testing.T) {
	// 13+6+14 = 33 plus two gaps = 37; ask for 30 and the name gives 7.
	w := Fit(cols, rows, 30)
	require.Equal(t, []int{6, 6, 14}, w)
	// Below what the name can give, the widest value column shrinks.
	w = Fit(cols, rows, 22)
	require.Equal(t, 4, w[0])
	require.Equal(t, 8, w[2])
	require.Equal(t, 22-Gap*2, w[0]+w[1]+w[2])
}

func TestRowPadsAllButTheLastCell(t *testing.T) {
	w := []int{5, 3, 4}
	require.Equal(t, "ab     cd   efgh", Row(w, []string{"ab", "cd", "efgh"}))
	got := Row(w, []string{"abcdefg", "c", "efghijkl"})
	require.Equal(t, 5+2+3+2+4, ansi.StringWidth(got))
	require.Equal(t, "Name   So…  Ver…", ansi.Strip(Titles(cols, w)))
}
