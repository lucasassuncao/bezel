package shell

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
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
	return New(Config{Layout: twoPane(), Tabs: tabs, Theme: theme.Resolve(theme.Plain, true), Title: "demo"})
}

func press(k string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: rune(k[0]), Text: k} }

func TestNewStartsFocusedOnTheFirstLeaf(t *testing.T) {
	s := newShell()
	require.Equal(t, "list", s.Focus())
	require.Equal(t, 0, s.ActiveTab())
}

// A value Tab is fixed at New: SetTabs is how its name and status follow the app.
func TestSetTabsReplacesWhatTheShellReads(t *testing.T) {
	s := sized(newShell(stubTab{name: "a"}, stubTab{name: "b"})).SetTab(1)
	s = s.SetTabs(stubTab{name: "a"}, stubTab{name: "renamed"})
	require.Equal(t, 1, s.ActiveTab())
	require.Contains(t, ansi.Strip(s.View(demoPanes())), "status of renamed")
}

func TestSetTabsKeepsEachTabsFocusAndClampsTheActiveOne(t *testing.T) {
	a, b, c := stubTab{name: "a"}, stubTab{name: "b"}, stubTab{name: "c"}
	s := sized(newShell(a, b)).SetFocus("detail")

	s = s.SetTabs(a, b, c)
	require.Equal(t, "detail", s.Focus())
	require.Equal(t, "list", s.SetTab(2).Focus())

	s = s.SetTab(2).SetTabs(a)
	require.Equal(t, 0, s.ActiveTab())
	require.Equal(t, "detail", s.Focus())
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
		Theme:  theme.Resolve(theme.Plain, true),
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
	require.Equal(t, 24-2-3, s.Body().H) // header and blank row, status, both legend rows
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
	one := New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.Plain, true), LegendLines: 1})
	one, _, _ = one.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	one = one.SetActions(keys...)
	require.Equal(t, 24-2-1-1, one.Body().H) // header and blank row, status, legend

	two := one.SetLegendLines(2)
	require.Equal(t, 24-2-1-2, two.Body().H)
	require.Equal(t, 2, New(Config{LegendLines: -3}).cfg.LegendLines)
}

// Moving focus to a pane with fewer keys must not grow the panes: the legend
// keeps its rows, and the ones it does not fill stay blank.
func TestTheBodyKeepsItsHeightWhenTheLegendShrinks(t *testing.T) {
	keys := make([]Action, 12)
	for i := range keys {
		keys[i] = Custom(string(rune('a'+i)), "a long action description", nil)
	}
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	long := s.SetActions(keys...)
	short := s.SetActions(keys[0])
	require.Equal(t, long.Body(), short.Body())
	require.Equal(t, long.Rect("detail"), short.Rect("detail"))

	rows := strings.Split(ansi.Strip(short.View(demoPanes())), "\n")
	require.Len(t, rows, 24)
	require.Contains(t, rows[22], "[a] a long action description")
	require.Empty(t, strings.TrimSpace(rows[23]))
}

// requireLaidOut fails unless s holds what its state resolves to now: the
// rects for the current body, and focus on a leaf that was placed.
func requireLaidOut(t *testing.T, s Shell, op string) {
	t.Helper()
	present := func(name string) bool { return s.present == nil || s.present[name] }
	want, collapsed := layout.Resolve(s.cfg.Layout, s.Body(), present)
	require.Equal(t, want, s.rects, op)
	require.Equal(t, collapsed, s.collapsed, op)
	_, placed := s.rects[s.Focus()]
	require.True(t, placed, "%s: focus on %q, which is not placed", op, s.Focus())
}

func send(msg tea.Msg) func(Shell) Shell {
	return func(s Shell) Shell {
		s, _, _ = s.Update(msg)
		return s
	}
}

// The placement is cached on the shell. Every way to change the shell must
// leave it current, or the app reads rects for a body that is gone.
func TestEveryChangeLeavesTheLayoutResolved(t *testing.T) {
	cfg := actionConfig(stubTab{name: "a"}, stubTab{name: "b"})
	cfg.Actions = []Action{Help(), Commands(), ChangeTab(), Custom("ctrl+p", "pane", nil)}
	// The header takes a second row off the list, so focus alone moves the body.
	cfg.Header = func(ctx Context, _ int) []string {
		if ctx.Focus == "list" {
			return []string{"one"}
		}
		return []string{"one", "two"}
	}
	in := textinput.New()
	ops := []struct {
		name string
		do   func(Shell) Shell
	}{
		{"SetFocus", func(s Shell) Shell { return s.SetFocus("detail") }},
		{"SetTab", func(s Shell) Shell { return s.SetTab(1) }},
		{"SetTabs", func(s Shell) Shell { return s.SetTabs(stubTab{name: "a"}) }},
		{"SetLayout", func(s Shell) Shell { return s.SetLayout(layout.Rows(layout.Fill("detail"), layout.Fill("list"))) }},
		{"SetTheme", func(s Shell) Shell { return s.SetTheme(theme.Resolve(theme.Mint, false)) }},
		{"SetSubtitle", func(s Shell) Shell { return s.SetSubtitle("sub") }},
		{"SetActions", func(s Shell) Shell { return s.SetActions(Custom("x", "x", nil)) }},
		{"SetLegendLines", func(s Shell) Shell { return s.SetLegendLines(3) }},
		{"ShowHelp", func(s Shell) Shell { return s.ShowHelp() }},
		{"OpenCommands", func(s Shell) Shell { return s.OpenCommands() }},
		{"Push", func(s Shell) Shell { return s.Push(overlay.NewText("t", []string{"x"}, theme.Resolved{})) }},
		{"Pop", func(s Shell) Shell { return s.ShowHelp().Pop() }},
		{"SetInput", func(s Shell) Shell { return s.SetInput(in) }},
		{"ClearInput", func(s Shell) Shell { return s.SetInput(in).ClearInput() }},
		{"SetStatus", func(s Shell) Shell { s, _ = s.SetStatus("hi", OK, time.Second); return s }},
		{"ClearStatus", func(s Shell) Shell { return s.ClearStatus() }},
		{"Busy", func(s Shell) Shell { s, _ = s.Busy("working"); return s }},
		{"Idle", func(s Shell) Shell { s, _ = s.Busy("working"); return s.Idle() }},
		{"resize", send(tea.WindowSizeMsg{Width: 60, Height: 20})},
		{"tab key", send(bezeltest.Key("tab"))},
		{"help key", send(bezeltest.Key("?"))},
		{"esc on help", func(s Shell) Shell { return send(overlay.CloseMsg{})(s.ShowHelp()) }},
		{"focus message", send(FocusMsg{From: "list", To: "detail"})},
	}
	base := sized(New(cfg))
	for _, from := range []Shell{base, base.SetFocus("detail")} {
		requireLaidOut(t, from, "start")
		for _, op := range ops {
			requireLaidOut(t, op.do(from), op.name)
		}
	}
}
