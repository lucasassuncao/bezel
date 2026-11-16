package shell

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

func TestPrebuiltActionsPrintFixedText(t *testing.T) {
	cases := []struct {
		a         Action
		key, desc string
	}{
		{Help(), "?", "help"},
		{Commands(), ":", "commands"},
		{ChangeTab(), "tab", "change tab"},
		{ChangePane(), "tab", "change pane"},
		{Move(), "↑/↓", "move"},
		{Scroll(), "↑/↓", "scroll"},
		{Quit(), "q", "quit"},
	}
	for _, c := range cases {
		h := c.a.Entry().Help
		require.Equal(t, c.key, h.Key)
		require.Equal(t, c.desc, h.Desc)
	}
}

func TestWithKeyRebindsAPrebuiltButKeepsItsText(t *testing.T) {
	a := Commands(WithKey("ctrl+p"))
	e := a.Entry()
	require.Equal(t, "ctrl+p", e.Help.Key)
	require.Equal(t, "commands", e.Help.Desc)
	require.True(t, key.Matches(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}, a.binding()))
}

// Every filter in these apps is "/": it must never be split into keys.
func TestCustomBindsItsKeyLiterally(t *testing.T) {
	a := Custom("/", "filter", nil)
	require.Equal(t, []string{"/"}, a.binding().Keys())
	require.Equal(t, "/", a.Entry().Help.Key)
}

func TestCustomWithKeyKeepsItsLabel(t *testing.T) {
	a := Custom("←/→", "fold", nil, WithKey("left", "right"))
	require.Equal(t, "←/→", a.Entry().Help.Key)
	require.Equal(t, []string{"left", "right"}, a.binding().Keys())
}

func TestCommandHasNoKey(t *testing.T) {
	require.Empty(t, Command("goto", "jump", nil, Arg("path")).binding().Keys())
}

func TestNeedsReachesTheEntry(t *testing.T) {
	require.Equal(t, Capability("write"), Custom("e", "edit", nil, Needs("write")).Entry().Needs)
}

func TestSendEmitsItsMessage(t *testing.T) {
	type ping struct{}
	cmd := Custom("p", "ping", Send(ping{})).Run(ActionContext{})
	require.Equal(t, ping{}, cmd())
}

func TestQuitRunsTeaQuit(t *testing.T) {
	require.IsType(t, tea.QuitMsg{}, Quit().Run(ActionContext{})())
}

func TestCheckPassesDistinctKeys(t *testing.T) {
	require.NoError(t, Check(Help(), Commands(), Move(), Custom("r", "reveal", nil), Command("goto", "jump", nil)))
}

func TestCheckCatchesACustomOnAPrebuiltKey(t *testing.T) {
	err := Check(Help(), Custom("?", "question", nil))
	require.ErrorContains(t, err, `key "?"`)
}

// A custom "up" beside Move would print two meanings for one key.
func TestCheckCountsDisplayOnlyKeys(t *testing.T) {
	require.Error(t, Check(Move(), Custom("up", "page up", nil)))
}

func TestCheckCatchesTwoCustoms(t *testing.T) {
	require.Error(t, Check(Custom("d", "delete", nil), Custom("d", "delete field", nil)))
}

// A palette-only command given a key later prints that key, never "[]".
func TestACommandGivenAKeyPrintsIt(t *testing.T) {
	require.Equal(t, "g", Command("goto", "jump", nil, WithKey("g")).Entry().Help.Key)
}

func TestNilActionsTabsAndOptionsAreSkipped(t *testing.T) {
	require.NoError(t, Check(nil, Quit(nil), nil))
	tab := stubTab{name: "a", actions: []Action{nil, Custom("r", "reveal", nil)}}
	s := New(Config{Tabs: []Tab{nil, tab}, Actions: []Action{nil}}).SetActions(nil).SetTabs(nil, tab)
	require.NotPanics(t, func() { _ = s.Body(); _ = s.render(nil) })
	require.Equal(t, 0, s.SetTab(1).ActiveTab(), "the nil tab was left out, so there is no second")
}

func TestHelpOnlyLeavesTheLegendButStaysInHelpAndRuns(t *testing.T) {
	ran := false
	resize := Custom("z", "zoom", func(ActionContext) tea.Cmd { ran = true; return nil }, HelpOnly())
	s := New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.Plain(), true), Actions: []Action{Help(), resize, Quit()}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	plain := ansi.Strip(s.render(demoPanes()))
	require.NotContains(t, plain, "zoom")
	require.Contains(t, plain, "quit")
	require.Contains(t, ansi.Strip(s.ShowHelp().render(demoPanes())), "zoom")

	_, r := s.Update(press("z"))
	require.True(t, r.Handled)
	require.True(t, ran)
}

func TestSectionGroupsTheHelpUnderItsOwnHeadings(t *testing.T) {
	doc := Section("document")
	s := New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.Plain(), true), Title: "demo", Actions: []Action{
		Help(), Custom("ctrl+s", "save", nil, doc), Move(), Custom("=", "reset sizes", nil, Section("view")),
		Custom("ctrl+u", "undo", nil, doc), Quit(),
	}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	help := ansi.Strip(s.ShowHelp().render(demoPanes()))
	order := []string{"demo", "move", "quit", "document", "save", "undo", "view", "reset sizes"}
	at := -1
	for _, w := range order {
		i := strings.Index(help, w)
		require.Greater(t, i, at, "%q is out of place:\n%s", w, help)
		at = i
	}
	require.Contains(t, ansi.Strip(s.render(demoPanes())), "save", "the legend is not changed")
}

func TestFromLegendPrintsEachEntryAsItCame(t *testing.T) {
	entries := []legend.Entry{
		legend.New("↑/↓", "move"),
		{Help: key.Help{Key: "d", Desc: "delete"}, Needs: "write", Group: 1},
	}
	acts := FromLegend(entries)
	require.Len(t, acts, 2)
	require.Equal(t, entries[0], acts[0].Entry())
	require.Equal(t, entries[1], acts[1].Entry())
}

func TestFromLegendLeavesTheKeysToTheApp(t *testing.T) {
	c := actionConfig()
	c.Actions = FromLegend([]legend.Entry{legend.New("d", "delete")})
	s := sized(New(c))
	require.Contains(t, view(s), "delete")

	_, r := s.Update(press("d"))
	require.False(t, r.Handled, "the component that owns d still gets it")
}

func TestFromLegendSkipsWhatHasNothingToPrint(t *testing.T) {
	acts := FromLegend([]legend.Entry{
		{Help: key.Help{Desc: "no key"}},
		{Help: key.Help{Key: "x", Desc: "off"}, Disabled: true},
		legend.New("enter", "open"),
	})
	require.Len(t, acts, 1)
	require.Equal(t, "open", acts[0].Entry().Help.Desc)
}

func TestFromLegendGivesEveryActionItsOptions(t *testing.T) {
	acts := FromLegend([]legend.Entry{legend.New("↑/↓", "move"), legend.New("enter", "open")}, Group(1))
	require.Len(t, acts, 2)
	for _, a := range acts {
		require.Equal(t, 1, a.Entry().Group)
	}
}
