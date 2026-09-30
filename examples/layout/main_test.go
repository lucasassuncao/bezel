package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

// frame sends msg and renders; every row must be exactly w columns wide.
func frame(t *testing.T, m tea.Model, msg tea.Msg, w int) (tea.Model, string) {
	t.Helper()
	m, _ = m.Update(msg)
	out := ansi.Strip(m.View().Content)
	for i, l := range strings.Split(out, "\n") {
		require.Equal(t, w, ansi.StringWidth(l), "row %d", i)
	}
	return m, out
}

func TestLayoutCollapsesWhenNarrow(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24}, 80)
	t.Log("\n" + out)
	require.Contains(t, out, "side ")
	require.Contains(t, out, "main ")

	m, cmd := m.Update(bezeltest.Key("w"))
	m, out = frame(t, m, cmd(), narrowWidth)
	t.Log("\n" + out)
	require.NotContains(t, out, "side ")
	require.Contains(t, out, "main ")

	m, cmd = m.Update(bezeltest.Key("w"))
	_, out = frame(t, m, cmd(), 80)
	require.Contains(t, out, "side ")
}

// A terminal smaller than any pane's minimum still draws, one row per line.
func TestLayoutTinyTerminal(t *testing.T) {
	var m tea.Model = newModel()
	require.NotPanics(t, func() { frame(t, m, tea.WindowSizeMsg{Width: 30, Height: 8}, 30) })
}
