package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

// frame sends msg and renders; every row must be exactly 80 columns wide.
func frame(t *testing.T, m tea.Model, msg tea.Msg) (tea.Model, string) {
	t.Helper()
	m, _ = m.Update(msg)
	out := ansi.Strip(m.View().Content)
	for i, l := range strings.Split(out, "\n") {
		require.Equal(t, 80, ansi.StringWidth(l), "row %d", i)
	}
	return m, out
}

func TestShellTabsAndStatus(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Contains(t, out, "> main.go")
	require.Contains(t, out, "4 items")

	m, cmd := m.Update(bezeltest.Key("s"))
	require.NotNil(t, cmd, "s runs the save action")
	_, out = frame(t, m, cmd())
	require.Contains(t, out, "saved main.go")

	_, out = frame(t, m, bezeltest.Key("tab"))
	require.Contains(t, out, "> ideas")
}

func TestShellBusyShowsText(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(bezeltest.Key("b"))
	_, out := frame(t, m, cmd())
	require.Contains(t, out, "indexing files")
}
