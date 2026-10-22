package chart

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func barLines(bars []Bar, top float64, width int) []string {
	return strings.Split(ansi.Strip(Bars(bars, top, width, th())), "\n")
}

func TestBarsLineUpLabelsBarsAndValues(t *testing.T) {
	lines := barLines([]Bar{{Label: "api", Value: 8}, {Label: "database", Value: 4}, {Label: "web", Value: 1}}, 8, 30)
	require.Equal(t, []string{
		"api      ███████████████████ 8",
		"database █████████▌          4",
		"web      ██▍                 1",
	}, lines, "19 cells of bar: 4 of 8 is 9.5 of them, 1 of 8 is 2.375")
}

func TestBarsRightAlignTheirText(t *testing.T) {
	lines := barLines([]Bar{{Label: "cpu", Value: 12.5}, {Label: "mem", Value: 3, Text: "3 GiB"}}, 100, 20)
	for _, l := range lines {
		require.Equal(t, 20, ansi.StringWidth(l), "%q", l)
	}
	require.True(t, strings.HasSuffix(lines[0], "  12.5"), lines[0])
	require.True(t, strings.HasSuffix(lines[1], " 3 GiB"), lines[1])
}

func TestBarsCutALongLabelBeforeTheBar(t *testing.T) {
	lines := barLines([]Bar{{Label: "a-very-long-label", Value: 1}}, 1, 12)
	require.Equal(t, "a-v… █████ 1", lines[0], "a third of the width for the label")
}

func TestBarsTooNarrowForABarKeepTheLabelAndText(t *testing.T) {
	require.Equal(t, []string{"c… 100"}, barLines([]Bar{{Label: "cpu", Value: 1, Text: "100"}}, 1, 6))
}

func TestBarsClampToTheirTrack(t *testing.T) {
	lines := barLines([]Bar{{Label: "x", Value: 99}, {Label: "y", Value: -3}}, 8, 8)
	require.Equal(t, []string{"x ███ 99", "y     -3"}, lines)
	require.Equal(t, "z      1", barLines([]Bar{{Label: "z", Value: 1}}, 0, 8)[0], "no top draws no bar")
}

func TestBarsColour(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	got := Bars([]Bar{{Label: "ok", Value: 1}, {Label: "hot", Value: 1, Style: &red}}, 1, 12, th())
	rows := strings.Split(got, "\n")
	require.Contains(t, rows[0], th().Accent.Render("██████"))
	require.Contains(t, rows[1], red.Render("██████"), "a bar's own style wins")
	require.Contains(t, Bars([]Bar{{Label: "a", Value: 1}}, 1, 8, th(), InSeries(2)), th().Series[2].Render("████"))
}

func TestBarsOfNothing(t *testing.T) {
	require.Empty(t, Bars(nil, 8, 20, th()))
	require.Empty(t, Bars([]Bar{{Label: "a", Value: 1}}, 8, 0, th()))
}
