package overlay

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

func TestHelpListsEverySectionAndCloses(t *testing.T) {
	h := NewHelp("Keys", []HelpSection{
		{Name: "list", Entries: []legend.Entry{legend.New("↑/↓", "move"), legend.New("enter", "open")}},
		{Name: "detail", Entries: []legend.Entry{legend.New("pgdn", "page")}},
	}, lipgloss.NewStyle(), legend.Style{})
	plain := ansi.Strip(h.View(layout.Rect{W: 60, H: 20}))
	require.Contains(t, plain, "list")
	require.Contains(t, plain, "enter")
	require.Contains(t, plain, "detail")
	_, cmd := h.Update(press("?"))
	require.IsType(t, CloseMsg{}, cmd())
}

func TestHelpAlignsWideKeysByCells(t *testing.T) {
	h := NewHelp("", []HelpSection{
		{Name: "s", Entries: []legend.Entry{legend.New("漢字", "wide"), legend.New("ab", "narrow")}},
	}, lipgloss.NewStyle(), legend.Style{})
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
