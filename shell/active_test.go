package shell

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

func actionConfig(tabs ...Tab) Config {
	return Config{Layout: twoPane(), Tabs: tabs, Theme: theme.Resolve(theme.ThemePlain, true), Title: "demo"}
}

func sized(s Shell) Shell {
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	return s
}

// legendRows is the last two screen rows: the status and a one-line legend,
// or a two-line legend.
func legendRows(s Shell) string {
	rows := strings.Split(ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}})), "\n")
	return strings.Join(rows[len(rows)-2:], "\n")
}

// The legend follows the order the app declared, help and the prebuilt keys
// included: the app decides where help sits.
func TestLegendFollowsTheDeclaredOrder(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Custom("r", "reveal", nil), Move(), Help()))
	require.Contains(t, legendRows(s), "[r] reveal  [↑/↓] move  [?] help")
}

// Group sorts an action into a legend row; the prebuilt ones stay in the first.
func TestGroupPutsActionsOnTheirOwnRow(t *testing.T) {
	cfg := actionConfig()
	cfg.LegendLines = 3
	s := sized(New(cfg).SetActions(
		Custom("ctrl+s", "save", nil, Group(2)), Help(), Move(),
		Custom("enter", "add", nil, Group(1)), Custom("esc", "back", nil),
	))
	rows := strings.Split(ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}})), "\n")
	got := make([]string, 0, 3)
	for _, r := range rows[len(rows)-3:] {
		got = append(got, strings.TrimSpace(r))
	}
	require.Equal(t, []string{"[?] help  [↑/↓] move  [esc] back", "[enter] add", "[ctrl+s] save"}, got)
}

func TestHelpShowsOnlyWhenHelpIsAnAction(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Custom("r", "reveal", nil)))
	require.NotContains(t, legendRows(s), "[?]")
}

func TestNeedsAndWhenFilterTheLegend(t *testing.T) {
	cfg := actionConfig()
	cfg.Can = func(c Capability) bool { return c != "write" }
	never := func(Context) bool { return false }
	s := sized(New(cfg).SetActions(
		Custom("e", "edit", nil, Needs("write")),
		Custom("x", "hidden", nil, When(never)),
		Custom("r", "reveal", nil),
	))
	l := legendRows(s)
	require.Contains(t, l, "[r] reveal")
	require.NotContains(t, l, "edit")
	require.NotContains(t, l, "hidden")
}

// A palette-only command has no key, so it has no business in the legend.
func TestPaletteOnlyCommandsStayOutOfTheLegend(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Command("goto", "jump to a path", nil), Custom("r", "reveal", nil)))
	l := legendRows(s)
	require.NotContains(t, l, "jump to a path")
	require.NotContains(t, l, "[]")
}

func TestOnlyTheActiveTabsActionsReachTheLegend(t *testing.T) {
	a := stubTab{name: "a", actions: []Action{Custom("r", "reveal", nil)}}
	b := stubTab{name: "b", actions: []Action{Custom("z", "other", nil)}}
	s := sized(New(actionConfig(a, b)))
	l := legendRows(s)
	require.Contains(t, l, "[r] reveal")
	require.NotContains(t, l, "other")
}

func TestChangeTabNeedsTwoTabs(t *testing.T) {
	one := actionConfig(stubTab{name: "a"})
	one.Actions = []Action{ChangeTab()}
	require.NotContains(t, legendRows(sized(New(one))), "change tab")

	two := actionConfig(stubTab{name: "a"}, stubTab{name: "b"})
	two.Actions = []Action{ChangeTab()}
	require.Contains(t, legendRows(sized(New(two))), "[tab] change tab")
}

func TestGlobalActionsJoinTheStatesOwn(t *testing.T) {
	cfg := actionConfig()
	cfg.Actions = []Action{Help(), Quit()}
	s := sized(New(cfg).SetActions(Custom("r", "reveal", nil)))
	require.Contains(t, legendRows(s), "[?] help  [q] quit  [r] reveal")
}

// The shell is a value: reusing the slice passed to SetActions must not change it.
func TestSetActionsCopiesTheSlice(t *testing.T) {
	xs := []Action{Custom("a", "alpha", nil)}
	s := sized(New(actionConfig()).SetActions(xs...))
	xs[0] = Custom("b", "beta", nil)
	require.Contains(t, legendRows(s), "alpha")
}

// Tabs with legends of different heights give the panes different bodies.
func TestSetTabRelaysOutThePanes(t *testing.T) {
	a := stubTab{name: "a"}
	b := stubTab{name: "b", actions: []Action{Custom("r", "reveal", nil)}}
	s := sized(New(actionConfig(a, b))).SetTab(1)
	require.Equal(t, s.Body().H, s.Rect("list").H)
}
