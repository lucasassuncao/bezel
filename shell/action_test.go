package shell

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
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
		h := c.a.Entry().Help()
		require.Equal(t, c.key, h.Key)
		require.Equal(t, c.desc, h.Desc)
	}
}

func TestWithKeyRebindsAPrebuiltButKeepsItsText(t *testing.T) {
	e := Commands(WithKey("ctrl+p")).Entry()
	require.Equal(t, "ctrl+p", e.Help().Key)
	require.Equal(t, "commands", e.Help().Desc)
	require.True(t, key.Matches(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}, e.Binding))
}

// Every filter in these apps is "/": it must never be split into keys.
func TestCustomBindsItsKeyLiterally(t *testing.T) {
	e := Custom("/", "filter", nil).Entry()
	require.Equal(t, []string{"/"}, e.Keys())
	require.Equal(t, "/", e.Help().Key)
}

func TestCustomWithKeyKeepsItsLabel(t *testing.T) {
	e := Custom("←/→", "fold", nil, WithKey("left", "right")).Entry()
	require.Equal(t, "←/→", e.Help().Key)
	require.Equal(t, []string{"left", "right"}, e.Keys())
}

func TestCommandHasNoKey(t *testing.T) {
	require.Empty(t, Command("goto", "jump", nil, Arg("path")).Entry().Keys())
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
	require.Equal(t, "g", Command("goto", "jump", nil, WithKey("g")).Entry().Help().Key)
}
