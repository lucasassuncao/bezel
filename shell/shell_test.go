package shell

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

type stubTab struct {
	name    string
	actions []Action
}

func (t stubTab) Name() string              { return t.name }
func (t stubTab) Status(ctx Context) string { return "status of " + t.name + " in " + ctx.Focus }
func (t stubTab) Actions(Context) []Action  { return t.actions }

func twoPane() layout.Node {
	return layout.Columns(layout.Fixed("list", layout.Ratio(1, 3), layout.Min(26)), layout.Fill("detail")).Collapse(72, "detail")
}

func newShell(tabs ...Tab) Shell {
	return New(Config{Layout: twoPane(), Tabs: tabs, Theme: theme.Resolve(theme.ThemePlain, true), Title: "demo"})
}

func press(k string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: rune(k[0]), Text: k} }

func TestNewStartsFocusedOnTheFirstLeaf(t *testing.T) {
	s := newShell()
	require.Equal(t, "list", s.Focus())
	require.Equal(t, 0, s.ActiveTab())
}

func TestStatusTTLDefaultsToFiveSeconds(t *testing.T) {
	require.Equal(t, 5*time.Second, New(actionConfig()).cfg.StatusTTL)
}

func TestFocusMovesAlongTheLayoutOrderAndWraps(t *testing.T) {
	s := newShell()
	s = s.SetFocus("detail")
	require.Equal(t, "detail", s.Focus())
	s = s.nextPane(1)
	require.Equal(t, "list", s.Focus())
	s = s.nextPane(-1)
	require.Equal(t, "detail", s.Focus())
	require.Equal(t, "detail", s.SetFocus("nope").Focus()) // unknown leaf: unchanged
}

// An Info pane is skipped by tab both ways and never gets the first focus,
// but SetFocus still reaches it, so a long hint can be scrolled.
func TestTabSkipsInfoPanes(t *testing.T) {
	s := New(Config{
		Layout: layout.Columns(layout.Fixed("hint", layout.Ratio(1, 4)).Info(), layout.Fill("list"), layout.Fill("detail")),
		Theme:  theme.Resolve(theme.ThemePlain, true),
	})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Equal(t, "list", s.Focus())
	require.Equal(t, "detail", s.nextPane(1).Focus())
	require.Equal(t, "list", s.nextPane(1).nextPane(1).Focus())
	require.Equal(t, "detail", s.nextPane(-1).Focus())
	require.Equal(t, "hint", s.SetFocus("hint").Focus())
	// Leaving an Info pane goes back to the ring pane that had focus, either way.
	s = s.SetFocus("detail").SetFocus("hint")
	require.Equal(t, "detail", s.nextPane(1).Focus())
	require.Equal(t, "detail", s.nextPane(-1).Focus())
}

func TestEachTabRemembersItsFocus(t *testing.T) {
	s := newShell(stubTab{name: "a"}, stubTab{name: "b"})
	s = s.SetFocus("detail")
	s = s.SetTab(1)
	require.Equal(t, "list", s.Focus())
	s = s.SetTab(0)
	require.Equal(t, "detail", s.Focus())
	require.Equal(t, 0, s.SetTab(5).ActiveTab()) // out of range: unchanged
}

func TestNarrowIsReportedWhenTheLayoutCollapsed(t *testing.T) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	require.True(t, s.Narrow())
	require.Equal(t, layout.Rect{}, s.Rect("list"))
	require.Equal(t, 60, s.Rect("detail").W)
	require.Equal(t, "detail", s.Focus()) // the vanished pane cannot keep focus
}

func TestSetActionsFeedsATablessShell(t *testing.T) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s = s.SetActions(Custom("x", "do x", nil), Custom("y", "do y", nil))
	plain := ansi.Strip(s.View(demoPanes()))
	require.Contains(t, plain, "[x] do x")
	require.Equal(t, 24-2-2, s.Body().H) // header and blank row, status, one legend row
}

func TestSetLayoutReplacesTheTree(t *testing.T) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 90, Height: 24})
	s = s.SetLayout(layout.Columns(layout.Fill("a"), layout.Fill("b"), layout.Fill("c")))
	require.Equal(t, 30, s.Rect("a").W)
	require.Equal(t, layout.Rect{}, s.Rect("list"))
	require.Equal(t, "a", s.Focus()) // old focus vanished with the old tree
}

func TestSubtitleStatusAndShowHelp(t *testing.T) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s = s.SetSubtitle("config.yaml ● modified")
	s, _ = s.SetStatus("hello", OK, 0)
	require.Equal(t, "hello", s.Status())
	plain := ansi.Strip(s.View(demoPanes()))
	require.Contains(t, plain, "config.yaml ● modified")

	s = s.ShowHelp()
	require.IsType(t, overlay.Help{}, s.TopOverlay())
}

func TestLegendLinesIsConfigurable(t *testing.T) {
	keys := make([]Action, 12)
	for i := range keys {
		keys[i] = Custom(string(rune('a'+i)), "a long action description", nil)
	}
	one := New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.ThemePlain, true), LegendLines: 1})
	one, _, _ = one.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	one = one.SetActions(keys...)
	require.Equal(t, 24-2-1-1, one.Body().H) // header and blank row, status, legend

	two := one.SetLegendLines(2)
	require.Equal(t, 24-2-1-2, two.Body().H)
	require.Equal(t, 2, New(Config{LegendLines: -3}).cfg.LegendLines)
}
