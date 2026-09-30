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

func TestTableDescriptionGivesWayFirst(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 120, Height: 24}, 120)
	t.Log("\n" + out)
	require.Contains(t, out, "A framework for terminal apps, after The Elm Architecture")

	m, cmd := m.Update(bezeltest.Key("w"))
	_, out = frame(t, m, cmd(), narrowWidth)
	t.Log("\n" + out)
	require.NotContains(t, out, "The Elm Architecture", "the description was cut")
	require.Contains(t, out, "v1.11.1", "versions stay whole")
	require.Contains(t, out, "BSD-3", "licences stay whole")
}
