package overlay

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	_, r := h.Update(press("?"))
	cmd := r.Cmd
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

func TestHelpClosesOnTheKeysItWasGiven(t *testing.T) {
	keys := DefaultHelpKeys()
	keys.Close = bind("x", "close")
	h := NewHelp(nil, theme.Resolved{}).WithKeys(keys)

	_, r := h.Update(press("q"))
	cmd := r.Cmd
	require.Nil(t, cmd)
	_, r = h.Update(press("x"))
	cmd = r.Cmd
	require.Equal(t, tea.Msg(CloseMsg{}), cmd())
	require.Contains(t, ansi.Strip(h.View(layout.Rect{W: 60, H: 20})), "x closes")
}
