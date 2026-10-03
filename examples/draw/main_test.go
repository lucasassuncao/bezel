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

func TestDrawScrollsThroughTheSections(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "HEADER")
	require.NotContains(t, out, "EMPTYSTATE")

	for range 40 {
		m, _ = m.Update(bezeltest.Key("down"))
	}
	_, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "nothing here")
	require.NotContains(t, out, "HEADER")
}

func TestDrawShowsBothDividers(t *testing.T) {
	var m tea.Model = newModel()
	_, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	require.Contains(t, out, "DIVIDER")
	require.Contains(t, out, "── Packages ──")
	require.Contains(t, out, strings.Repeat("─", 40), "the bare rule spans the pane")
}

func TestDrawShowsACardAndButtons(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	for range 14 {
		m, _ = m.Update(bezeltest.Key("down"))
	}
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "─ git ─")
	require.Contains(t, out, "[u] upgrade")
	require.Contains(t, out, "Save")
	require.Contains(t, out, "Cancel")
}
