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

func TestTextboxTypesEveryKey(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "so no line comes back doubled.")

	for _, k := range strings.Split("quit", "") {
		m, _ = m.Update(bezeltest.Key(k))
	}
	_, out = frame(t, m, nil)
	require.Contains(t, out, "quit", "q is text, not the quit key")
}

func TestTextboxCountsLines(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(bezeltest.Key("ctrl+s"))
	_, out := frame(t, m, cmd())
	t.Log("\n" + out)
	require.Contains(t, out, "4 lines", "CRLF folded: three lines and the empty last one")
}
