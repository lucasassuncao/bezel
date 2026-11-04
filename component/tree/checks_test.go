package tree

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

var space = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}

func TestSpaceChecksOnlyWithWithMulti(t *testing.T) {
	_, r := newTree(10).Update(space)
	require.False(t, r.Handled)

	m, r := newTree(10).WithMulti(true).Update(space)
	require.Equal(t, Toggled[node]{Nodes: []node{nodes()[1]}}, r.Event)
	require.True(t, m.IsChecked(1))
	require.Equal(t, []node{nodes()[1]}, m.Checked())

	m, _ = m.Update(space)
	require.Empty(t, m.Checked(), "a second press unchecks")
}

func TestCheckAllChecksWhatIsShownThenClearsIt(t *testing.T) {
	m := newTree(10).WithMulti(true)
	m, r := m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	require.Len(t, r.Event.(Toggled[node]).Nodes, 4, "server, host, port, cache; not the headings, not ttl")
	require.Len(t, m.Checked(), 4)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	require.Empty(t, m.Checked())
}

func TestABoxIsDrawnBeforeEachRow(t *testing.T) {
	m, _ := newTree(4).WithMulti(true).Update(space)
	out := ansi.Strip(m.View(theme.Resolved{}, func(n node, _ int, _ bool) string { return n.label }))
	require.Equal(t, "    ADDED\n[x] server\n[ ] host\n  ↓ 3 more", out)
	require.Empty(t, m.SetNodes(nodes()[:3]).Checked(), "a reload of another length clears the checks")
	require.Len(t, m.ClearChecks().Checked(), 0)
}
