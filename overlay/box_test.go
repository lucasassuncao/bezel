package overlay

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

// A dialog's padding counts against the body too, and the cut spares its last
// line: that is where the buttons are, and with them how to leave.
func TestAClippedDialogFitsAndKeepsItsButtons(t *testing.T) {
	var lines []string
	for i := range 40 {
		lines = append(lines, fmt.Sprintf("row %d", i))
	}
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3)
	c := NewChoice("Apply?", lines, []Option{Opt("enter/y", "apply", nil)}, style, legend.Style{})
	got := ansi.Strip(c.View(layout.Rect{W: 80, H: 20}))
	require.LessOrEqual(t, len(strings.Split(got, "\n")), 20, "the dialog is taller than the body")
	require.Contains(t, got, "apply (enter/y)", "the cut took the buttons")
	require.Contains(t, got, "more lines")
}

func TestBoxClipsToTheBodyAndSaysHowMuchIsMissing(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	got := Box(lines, layout.Rect{W: 40, H: 8}, lipgloss.NewStyle().Border(lipgloss.NormalBorder()))
	rows := strings.Split(got, "\n")
	require.LessOrEqual(t, len(rows), 8)
	require.Contains(t, ansi.Strip(got), "more lines")
	for _, r := range rows {
		require.LessOrEqual(t, ansi.StringWidth(r), 40)
	}
}
