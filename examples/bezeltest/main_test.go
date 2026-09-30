package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

// press sends each key the way a terminal would and returns the last command.
func press(m tea.Model, keys ...string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	for _, k := range keys {
		m, cmd = m.Update(bezeltest.Key(k))
	}
	return m, cmd
}

func TestCounterCountsKeys(t *testing.T) {
	m, _ := press(model{}, "+", "+", "up", "-")
	require.Equal(t, 2, m.(model).n)
	require.Contains(t, m.View().Content, "count: 2")
}

func TestCounterQuits(t *testing.T) {
	for _, k := range []string{"q", "ctrl+c"} {
		_, cmd := press(model{}, k)
		require.NotNil(t, cmd, k)
		require.IsType(t, tea.QuitMsg{}, cmd(), k)
	}
}

// Key reads back as the name it was given, so a binding on "ctrl+s" matches.
func TestKeyReadsBackAsItsName(t *testing.T) {
	for _, k := range []string{"+", "ctrl+s", "enter", "shift+tab", "?"} {
		require.Equal(t, k, bezeltest.Key(k).String())
	}
}
