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

func typeKeys(m tea.Model, keys ...string) tea.Model {
	for _, k := range keys {
		m, _ = m.Update(bezeltest.Key(k))
	}
	return m
}

func TestBrowserPreviewFollowsTheCursor(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 16})
	t.Log("\n" + out)
	require.Contains(t, out, "PANCAKES")

	m = typeKeys(m, "down")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "GUACAMOLE")
}

func TestBrowserTabScrollsTheRecipe(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 16})
	m = typeKeys(m, "tab", "down", "down", "down", "down")
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.NotContains(t, out, "PANCAKES", "the recipe scrolled past its title")
	require.Contains(t, out, "pancakes", "the list kept its place")
}
