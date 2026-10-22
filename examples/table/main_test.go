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
	var row string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "▶ bubbletea") {
			row = l
		}
	}
	require.NotEmpty(t, row)
	require.NotContains(t, row, "The Elm Architecture", "the table's own row was cut")
	require.Contains(t, out, "v1.11.1", "versions stay whole")
	require.Contains(t, out, "BSD-3", "licences stay whole")
}

func TestTableCursorScrollsAndTheEdgesFollow(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 100, Height: 24}, 100)
	require.Contains(t, out, " 1/18 ")

	m, out = frame(t, m, bezeltest.Key("end"), 100)
	t.Log("\n" + out)
	require.Contains(t, out, "▶ x/sync")
	require.Contains(t, out, " 18/18 ")
	require.Contains(t, out, "NAME", "the header stays put")

	_, cmd := m.Update(bezeltest.Key("enter"))
	require.NotNil(t, cmd, "enter sets a status that clears itself")
}
