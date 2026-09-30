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
