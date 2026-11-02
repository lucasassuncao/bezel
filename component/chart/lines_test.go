package chart

import (
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func lineRows(lines []Line, lo, hi float64, w, h int) []string {
	return strings.Split(ansi.Strip(Lines(lines, lo, hi, w, h, th())), "\n")
}

func TestLinesDrawBrailleBesideTheAxis(t *testing.T) {
	rows := lineRows([]Line{{Values: []float64{0, 4, 8}}}, 0, 8, 6, 2)
	require.Equal(t, []string{"8┤   ⡎", "0┤  ⢰⠁"}, rows, "a rise joins its dots, the newest on the right")
}

func TestLinesLabelTheEndsAndAnExactMiddle(t *testing.T) {
	rows := lineRows(nil, 0, 100, 10, 5)
	require.Equal(t, []string{"100┤", "   │", " 50┤", "   │", "  0┤"}, trimmed(rows))
	rows = lineRows(nil, 0, 100, 10, 4)
	require.Equal(t, "   │", strings.TrimRight(rows[1], " "), "no row sits halfway")
	require.Equal(t, "   │", strings.TrimRight(rows[2], " "))
}

func TestLinesPutAKeyUnderNamedLines(t *testing.T) {
	out := Lines([]Line{{Label: "cpu", Values: []float64{1}}, {Label: "mem", Values: []float64{2}}}, 0, 2, 20, 3, th())
	rows := strings.Split(out, "\n")
	require.Len(t, rows, 3, "the key takes the last row")
	require.Equal(t, "● cpu  ● mem", ansi.Strip(rows[2]))
	require.Contains(t, rows[2], th().Series[0].Render("●"))
	require.Contains(t, rows[2], th().Series[1].Render("●"))
}

func TestLinesColours(t *testing.T) {
	require.Contains(t, Lines([]Line{{Values: []float64{1, 1}}}, 0, 2, 4, 1, th()), th().Accent.Render("⠒"), "a lone line in the accent")
	require.Contains(t, Lines([]Line{{Values: []float64{1, 1}}}, 0, 2, 4, 1, th(), InSeries(2)), th().Series[2].Render("⠒"))
	two := Lines([]Line{{Values: []float64{0, 0}}, {Values: []float64{2, 2}}}, 0, 2, 4, 2, th(), InSeries(5))
	require.Contains(t, two, th().Series[5].Render("⣀"))
	require.Contains(t, two, th().Series[0].Render("⠉"), "the next line takes the next colour, round the six")
}

func TestLinesBreakAtNaNAndClampOutOfRange(t *testing.T) {
	rows := lineRows([]Line{{Values: []float64{8, math.NaN(), 8, 99, -5}}}, 0, 8, 5, 1)
	require.Equal(t, "8┤⠈⠈⢣", rows[0], "the gap at the NaN, 99 at the top, -5 at the bottom")
}

func TestLinesWithNothingToScale(t *testing.T) {
	require.Empty(t, Lines(nil, 0, 1, 0, 3, th()))
	require.Equal(t, []string{"│   ", "│   "}, lineRows([]Line{{Values: []float64{1}}}, 5, 5, 4, 2), "hi not above lo: an axis and no line")
	require.Equal(t, "│ ", lineRows(nil, 0, 1000000, 2, 1)[0], "no room for the labels: the axis alone")
}

func trimmed(rows []string) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = strings.TrimRight(r, " ")
	}
	return out
}
