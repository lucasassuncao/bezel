package draw

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
)

func TestPanelIsExactlyItsRect(t *testing.T) {
	r := layout.Rect{W: 20, H: 5}
	got := Panel(r, "Title", "one\ntwo\nthree\nfour\nfive\nsix", PanelStyle{})
	lines := strings.Split(got, "\n")
	require.Len(t, lines, 5)
	for _, l := range lines {
		require.Equal(t, 20, ansi.StringWidth(l), "%q", l)
	}
	require.Contains(t, ansi.Strip(lines[0]), "Title")
	require.Contains(t, ansi.Strip(lines[1]), "one")
	require.NotContains(t, ansi.Strip(got), "four")
}

func TestPanelTruncatesALongTitle(t *testing.T) {
	got := Panel(layout.Rect{W: 10, H: 3}, "a very long title indeed", "", PanelStyle{})
	for _, l := range strings.Split(got, "\n") {
		require.Equal(t, 10, ansi.StringWidth(l))
	}
}

func TestPanelSurvivesTinyRects(t *testing.T) {
	require.NotPanics(t, func() { Panel(layout.Rect{W: 0, H: 0}, "t", "x", PanelStyle{}) })
	require.NotPanics(t, func() { Panel(layout.Rect{W: -3, H: 1}, "t", "x", PanelStyle{}) })
	got := Panel(layout.Rect{W: 8, H: 1}, "t", "x", PanelStyle{})
	require.Len(t, strings.Split(got, "\n"), 1) // title row only
}

func TestPanelEdgeSurvivesAMonochromeTerminal(t *testing.T) {
	r := layout.Rect{W: 20, H: 4}
	idle := ansi.Strip(Panel(r, "t", "x", PanelStyle{}))
	thick := ansi.Strip(Panel(r, "t", "x", PanelStyle{Edge: lipgloss.ThickBorder()}))
	require.NotEqual(t, idle, thick)
	require.Contains(t, idle, "╭")
	require.Contains(t, thick, "┏")
	for _, l := range strings.Split(thick, "\n") {
		require.Equal(t, 20, ansi.StringWidth(l))
	}
}
