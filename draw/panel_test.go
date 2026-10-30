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
	got := Panel(r, Edges{Title: "Title"}, "one\ntwo\nthree\nfour\nfive\nsix", PanelStyle{})
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
	got := Panel(layout.Rect{W: 10, H: 3}, Edges{Title: "a very long title indeed"}, "", PanelStyle{})
	for _, l := range strings.Split(got, "\n") {
		require.Equal(t, 10, ansi.StringWidth(l))
	}
}

func TestPanelSurvivesTinyRects(t *testing.T) {
	require.NotPanics(t, func() { Panel(layout.Rect{W: 0, H: 0}, Edges{Title: "t"}, "x", PanelStyle{}) })
	require.NotPanics(t, func() { Panel(layout.Rect{W: -3, H: 1}, Edges{Title: "t"}, "x", PanelStyle{}) })
	got := Panel(layout.Rect{W: 8, H: 1}, Edges{Title: "t"}, "x", PanelStyle{})
	require.Len(t, strings.Split(got, "\n"), 1) // title row only
}

func TestPanelEdgeSurvivesAMonochromeTerminal(t *testing.T) {
	r := layout.Rect{W: 20, H: 4}
	idle := ansi.Strip(Panel(r, Edges{Title: "t"}, "x", PanelStyle{}))
	thick := ansi.Strip(Panel(r, Edges{Title: "t"}, "x", PanelStyle{Edge: lipgloss.ThickBorder()}))
	require.NotEqual(t, idle, thick)
	require.Contains(t, idle, "╭")
	require.Contains(t, thick, "┏")
	for _, l := range strings.Split(thick, "\n") {
		require.Equal(t, 20, ansi.StringWidth(l))
	}
}

func TestPanelPutsTheRightTextAtTheEndOfTheTopEdge(t *testing.T) {
	got := Panel(layout.Rect{W: 30, H: 3}, Edges{Title: "files", Right: "35/152"}, "", PanelStyle{})
	top := ansi.Strip(strings.Split(got, "\n")[0])
	require.Equal(t, 30, ansi.StringWidth(top))
	require.True(t, strings.HasPrefix(top, "╭─ files ─"), top)
	require.True(t, strings.HasSuffix(top, "─ 35/152 ─╮"), top)
}

// The title says what the panel is; the right text only adds to it.
func TestPanelDropsTheRightTextBeforeTheTitle(t *testing.T) {
	got := Panel(layout.Rect{W: 18, H: 3}, Edges{Title: "files", Right: "35/152"}, "", PanelStyle{})
	top := ansi.Strip(strings.Split(got, "\n")[0])
	require.Equal(t, 18, ansi.StringWidth(top))
	require.Contains(t, top, "files")
	require.NotContains(t, top, "35/152")
}

func TestFramedPanelPutsTheBottomTextAtTheEndOfTheBottomEdge(t *testing.T) {
	got := Panel(layout.Rect{W: 30, H: 4}, Edges{Title: "files", Bottom: "12/340"}, "", PanelStyle{})
	bottom := ansi.Strip(strings.Split(got, "\n")[3])
	require.Equal(t, 30, ansi.StringWidth(bottom))
	require.True(t, strings.HasPrefix(bottom, "╰─"), bottom)
	require.True(t, strings.HasSuffix(bottom, "─ 12/340 ─╯"), bottom)
}

func TestFramedPanelDropsABottomTextThatDoesNotFit(t *testing.T) {
	got := Panel(layout.Rect{W: 10, H: 3}, Edges{Bottom: "a long position"}, "", PanelStyle{})
	require.Equal(t, "╰────────╯", ansi.Strip(strings.Split(got, "\n")[2]))
}

// rightEdge is the last cell of every body row of a framed panel.
func rightEdge(t *testing.T, r layout.Rect, s layout.Scroll) string {
	t.Helper()
	lines := strings.Split(ansi.Strip(Panel(r, Edges{Scroll: s}, "", PanelStyle{})), "\n")
	require.Len(t, lines, r.H)
	var edge strings.Builder
	for _, l := range lines[1 : len(lines)-1] {
		require.Equal(t, r.W, ansi.StringWidth(l), "%q", l)
		edge.WriteString(ansi.TruncateLeft(l, r.W-1, ""))
	}
	return edge.String()
}

func TestFramedPanelDrawsTheThumbDownTheRightEdge(t *testing.T) {
	r := layout.Rect{W: 12, H: 7} // five body rows: the track
	require.Equal(t, "█││││", rightEdge(t, r, layout.Scroll{Visible: 5, Total: 25}))
	require.Equal(t, "││││█", rightEdge(t, r, layout.Scroll{Offset: 20, Visible: 5, Total: 25}))
	require.Equal(t, "││██│", rightEdge(t, r, layout.Scroll{Offset: 5, Visible: 5, Total: 12}))
	require.Equal(t, "│││││", rightEdge(t, r, layout.Scroll{Visible: 5, Total: 5}), "nothing hidden, no thumb")
}

func TestFramedPanelThumbWearsItsOwnStyle(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	got := Panel(layout.Rect{W: 12, H: 5}, Edges{Scroll: layout.Scroll{Visible: 3, Total: 9}}, "", PanelStyle{Thumb: red})
	require.Contains(t, got, red.Render("█"))
}
