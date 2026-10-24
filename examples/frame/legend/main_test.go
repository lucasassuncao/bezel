package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/shell"
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

// press sends one key and feeds back the message its command emits.
func press(t *testing.T, m tea.Model, key string) tea.Model {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key(key))
	if cmd != nil {
		m, _ = m.Update(cmd())
	}
	return m
}

func TestLegendCountsWhatItCuts(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "[w] write mode")
	require.Contains(t, out, "in [?]", "the cut entries are counted")

	m = press(t, m, "w")
	m = press(t, m, "?")
	_, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "delete", "help lists the write keys once allowed")
}

func TestLegendHelpHidesWriteKeysWhenReadOnly(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = press(t, m, "?")
	_, out := frame(t, m, nil)
	require.Contains(t, out, "export")
	require.NotContains(t, out, "delete")
}

func TestLegendKeysDoNotCollide(t *testing.T) {
	require.NoError(t, shell.Check(actions()...))
}
