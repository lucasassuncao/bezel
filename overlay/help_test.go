package overlay

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

func TestHelpListsEverySectionAndCloses(t *testing.T) {
	h := NewHelp([]HelpSection{
		{Name: "list", Entries: []legend.Entry{legend.New("↑/↓", "move"), legend.New("enter", "open")}},
		{Name: "detail", Entries: []legend.Entry{legend.New("pgdn", "page")}},
	}, theme.Resolved{}).WithTitle("Keys")
	plain := ansi.Strip(h.View(layout.Rect{W: 60, H: 20}))
	require.True(t, strings.HasPrefix(strings.TrimSpace(plain), "Keys"), "the title leads the box")
	require.Contains(t, plain, "list")
	require.Contains(t, plain, "enter")
	require.Contains(t, plain, "detail")
	_, cmd := h.Update(press("?"))
	require.IsType(t, CloseMsg{}, cmd())
}

func TestHelpAlignsWideKeysByCells(t *testing.T) {
	h := NewHelp([]HelpSection{
		{Name: "s", Entries: []legend.Entry{legend.New("漢字", "wide"), legend.New("ab", "narrow")}},
	}, theme.Resolved{})
	col := map[string]int{}
	for _, l := range strings.Split(ansi.Strip(h.View(layout.Rect{W: 60, H: 20})), "\n") {
		for _, d := range []string{"wide", "narrow"} {
			if i := strings.Index(l, d); i >= 0 {
				col[d] = ansi.StringWidth(l[:i])
			}
		}
	}
	require.Equal(t, col["wide"], col["narrow"])
}
