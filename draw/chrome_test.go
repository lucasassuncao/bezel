package draw

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
)

func TestHeaderFillsWidthWithRightAlignedInfo(t *testing.T) {
	got := Header(40, "myapp", "prod", "v1.2", ChromeStyle{})
	require.Equal(t, 40, ansi.StringWidth(got))
	plain := ansi.Strip(got)
	require.True(t, strings.HasPrefix(strings.TrimSpace(plain), "myapp"))
	require.True(t, strings.HasSuffix(strings.TrimSpace(plain), "v1.2"))
	require.Equal(t, 10, ansi.StringWidth(Header(10, "a very long title", "", "", ChromeStyle{})))
}

func TestTabsMarkTheActiveOne(t *testing.T) {
	got := Tabs(40, []string{"secrets", "policies"}, 1, ChromeStyle{TabActive: lipgloss.NewStyle().Bold(true)})
	require.Equal(t, 40, ansi.StringWidth(got))
	require.Contains(t, ansi.Strip(got), "secrets")
	require.Contains(t, ansi.Strip(got), "policies")
}

func TestStatusLineIsExactlyWidth(t *testing.T) {
	require.Equal(t, 12, ansi.StringWidth(StatusLine(12, "ok", lipgloss.NewStyle())))
	require.Equal(t, 5, ansi.StringWidth(StatusLine(5, "a long status", lipgloss.NewStyle())))
}

func TestEmptyStateCentresWithinTheRect(t *testing.T) {
	got := EmptyState(layout.Rect{W: 20, H: 5}, "nothing", "press a", ChromeStyle{})
	lines := strings.Split(got, "\n")
	require.Len(t, lines, 5)
	require.Contains(t, ansi.Strip(got), "nothing")
	require.Contains(t, ansi.Strip(got), "press a")
}

func TestDividerFillsTheWidth(t *testing.T) {
	plain := lipgloss.NewStyle()
	for _, width := range []int{1, 8, 40} {
		require.Equal(t, width, ansi.StringWidth(Divider(width, "", plain, plain)), "bare rule at %d", width)
		require.Equal(t, width, ansi.StringWidth(Divider(width, "Packages", plain, plain)), "labelled rule at %d", width)
	}
	require.Equal(t, "── Packages ──────", Divider(18, "Packages", plain, plain))
	require.Empty(t, Divider(0, "x", plain, plain))
}

func TestDividerMeasuresAPaddedLabelStyle(t *testing.T) {
	padded := lipgloss.NewStyle().Padding(0, 1)
	for _, width := range []int{8, 18, 40} {
		require.Equal(t, width, ansi.StringWidth(Divider(width, "Packages", lipgloss.NewStyle(), padded)), "width %d", width)
	}
}

func TestButtonKeepsItsLabel(t *testing.T) {
	ink := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	require.Equal(t, "    Save    ", ansi.Strip(Button("Save", ink, true)))
	require.Equal(t, ansi.StringWidth(Button("Save", ink, true)), ansi.StringWidth(Button("Save", ink, false)), "focus does not move it")
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
