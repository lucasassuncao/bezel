package braille

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestDotsAreTwoByFourACell(t *testing.T) {
	w, h := New(3, 2).Dots()
	require.Equal(t, 6, w)
	require.Equal(t, 8, h)
}

func TestSetLightsTheRightBrailleDot(t *testing.T) {
	g := New(1, 1)
	g.Set(0, 0, 0)
	require.Equal(t, "⠁", g.Row(0, nil))
	g.Set(1, 3, 0)
	require.Equal(t, "⢁", g.Row(0, nil))
}

func TestSetOutsideTheGridIsDropped(t *testing.T) {
	g := New(1, 1)
	for _, p := range [][2]int{{-1, 0}, {0, -1}, {2, 0}, {0, 4}} {
		g.Set(p[0], p[1], 0)
	}
	require.Equal(t, " ", g.Row(0, nil))
}

func TestLineLightsBothEndsAndEveryStep(t *testing.T) {
	g := New(2, 1)
	g.Line(0, 0, 3, 0, 0)
	require.Equal(t, "⠉⠉", g.Row(0, nil))

	g = New(1, 1)
	g.Line(0, 3, 0, 0, 0) // drawn upwards
	require.Equal(t, "⡇", g.Row(0, nil))

	g = New(2, 1)
	g.Line(0, 0, 3, 3, 0)
	require.Equal(t, "⠑⢄", g.Row(0, nil))
}

// Every slope must reach its far end and light no more dots than its run.
func TestLineEndsForEverySlope(t *testing.T) {
	for x1 := range 8 {
		for y1 := range 8 {
			g := New(4, 2)
			g.Line(0, 7, x1, 7-y1, 0)
			lit := 0
			for _, b := range g.bits {
				for ; b != 0; b &= b - 1 {
					lit++
				}
			}
			require.Equal(t, max(x1, y1)+1, lit, "to %d,%d", x1, y1)
		}
	}
}

func TestTextCoversTheDotsAndIsCutAtTheEdge(t *testing.T) {
	g := New(4, 1)
	g.Line(0, 0, 7, 0, 0)
	g.Text(2, 0, "abc", 0)
	require.Equal(t, "⠉⠉ab", g.Row(0, nil))
	g.Text(0, 5, "x", 0) // a row outside is dropped
	require.Equal(t, "⠉⠉ab", g.Row(0, nil))
}

func TestTextKeepsAWideRuneInTwoCells(t *testing.T) {
	g := New(3, 1)
	g.Text(0, 0, "日x", 0)
	require.Equal(t, "日x", g.Row(0, nil))
	require.Equal(t, 3, ansi.StringWidth(g.Row(0, nil)))

	g = New(1, 1)
	g.Text(0, 0, "日", 0) // does not fit
	require.Equal(t, " ", g.Row(0, nil))
}

func TestRowColoursEachOwnersRun(t *testing.T) {
	g := New(2, 1)
	g.Set(0, 0, 0)
	g.Set(2, 0, 1)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	blue := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff"))
	row := g.Row(0, []lipgloss.Style{red, blue})
	require.Equal(t, red.Render("⠁")+blue.Render("⠁"), row)
	// An owner past the styles wraps round them.
	g.Set(2, 0, 3)
	require.Equal(t, red.Render("⠁")+blue.Render("⠁"), g.Row(0, []lipgloss.Style{red, blue}))
}

func TestRowOutsideTheGridIsEmpty(t *testing.T) {
	require.Empty(t, New(1, 1).Row(1, nil))
	require.Empty(t, New(0, 0).Row(0, nil))
}
