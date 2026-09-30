package palette

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

func view(m Model) string { return ansi.Strip(m.View(layout.Rect{W: 80, H: 20})) }

func TestNothingIsPickedUntilTheUserPicks(t *testing.T) {
	require.NotContains(t, view(open()), "›")
}

func TestThePanelPrintsTheKeyBesideTheName(t *testing.T) {
	require.Regexp(t, `:copy\s+y\s+copy the secret`, view(open()))
}

func TestUnavailableCommandsAreNotListed(t *testing.T) {
	require.NotContains(t, view(open()), ":delete")
}

func TestNoMatchSaysSo(t *testing.T) {
	require.Contains(t, view(typed(open(), "zz")), "no command matches")
}

// A terminal mid-resize can hand the panel any width, zero included.
func TestItDrawsAtAnyWidth(t *testing.T) {
	for _, w := range []int{0, 1, 5} {
		require.NotPanics(t, func() { open().View(layout.Rect{W: w, H: 10}) })
	}
}

func TestTheStatusLineIsTheTypedLine(t *testing.T) {
	require.True(t, strings.HasPrefix(ansi.Strip(typed(open(), "cr").StatusLine(40)), ":cr"))
}

// The panel never grows past the body, or it covers the typed line below it.
func TestThePanelFitsAShortBody(t *testing.T) {
	var its []Item
	for _, n := range []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9"} {
		its = append(its, Item{Name: n, Available: true})
	}
	m := New(its, theme.Resolve(theme.ThemePlain, true))
	out := ansi.Strip(m.View(layout.Rect{W: 60, H: 5}))
	require.LessOrEqual(t, len(strings.Split(out, "\n")), 5)
	require.Contains(t, out, "more")
}

// The columns hold still while the window scrolls: a long name below the fold
// still sets the width of the rows above it.
func TestTheColumnsDoNotShiftAsTheListScrolls(t *testing.T) {
	its := []Item{{Name: "a", Desc: "first", Available: true}}
	for _, n := range []string{"b", "c", "d", "e", "f", "g", "z-much-longer-name"} {
		its = append(its, Item{Name: n, Desc: "x", Available: true})
	}
	out := ansi.Strip(New(its, theme.Resolve(theme.ThemePlain, true)).View(layout.Rect{W: 80, H: 5}))
	require.NotContains(t, out, "z-much-longer-name")
	require.Contains(t, out, ":a"+strings.Repeat(" ", len(":z-much-longer-name")-len(":a")+2))
}
