package overlay

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

// A dialog's padding counts against the body too, and the cut spares its last
// line: that is where the buttons are, and with them how to leave.
func TestAClippedDialogFitsAndKeepsItsButtons(t *testing.T) {
	var lines []string
	for i := range 40 {
		lines = append(lines, fmt.Sprintf("row %d", i))
	}
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3)
	c := themed(NewChoice("Apply?", strings.Join(lines, "\n"), []Option{Opt("enter/y", "apply", nil)}), theme.Resolved{Modal: style})
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

// Every dialog frames its lines here, and a title or message from data may
// hold what Clean drops: drawn, it would move the dialog's right edge.
func TestBoxDrawsOnlyWhatItMeasured(t *testing.T) {
	body := layout.Rect{W: 60, H: 20}
	for _, bad := range []string{"ab\rX", "a\xffb", "a\x1b_Gf=1", "a\x1b[38;2;1"} {
		out := Box([]string{lipgloss.NewStyle().Bold(true).Render(bad), bad}, body, lipgloss.NewStyle().Border(lipgloss.RoundedBorder()))
		for _, l := range strings.Split(out, "\n") {
			require.Equal(t, draw.Clean(l), l, "%q", bad)
		}
	}
}
