package draw

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
)

func TestButtonDrawsInItsStyle(t *testing.T) {
	st := ButtonStyle{
		Focused: lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("2")).Padding(0, 2),
		Blurred: lipgloss.NewStyle().Background(lipgloss.Color("8")).Padding(0, 2),
	}
	require.Equal(t, "    Save    ", ansi.Strip(Button("Save", st, true)))
	require.Equal(t, st.Focused.Render("  Save  "), Button("Save", st, true))
	require.Equal(t, st.Blurred.Render("  Save  "), Button("Save", st, false))
	require.Equal(t, ansi.StringWidth(Button("Save", st, true)), ansi.StringWidth(Button("Save", st, false)), "focus does not move it")
}

func TestCardPinsTheFooterUnderARule(t *testing.T) {
	plain := PanelStyle{}
	got := ansi.Strip(Card(layout.Rect{W: 30, H: 8}, "git", "version 2.47.0\nsource winget", "[i] install", plain))
	t.Log("\n" + got)
	rows := strings.Split(got, "\n")
	require.Len(t, rows, 8)
	require.Contains(t, rows[1], "version 2.47.0")
	require.Contains(t, rows[5], strings.Repeat("─", 26), "the rule sits right above the footer")
	require.Contains(t, rows[6], "[i] install", "the footer is the last row inside the border")

	tall := ansi.Strip(Card(layout.Rect{W: 30, H: 5}, "git", "a\nb\nc\nd\ne", "[i] install", plain))
	require.Contains(t, tall, "[i] install", "a long body is cut, never the footer")
}
