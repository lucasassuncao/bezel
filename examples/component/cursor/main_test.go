package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

// frame sends msg and renders; every row must be exactly 80 columns wide. A
// command is not run: the blink would make the test wait on it.
func frame(t *testing.T, m tea.Model, msg tea.Msg) (tea.Model, string) {
	t.Helper()
	m, _ = m.Update(msg)
	out := ansi.Strip(m.View().Content)
	for i, l := range strings.Split(out, "\n") {
		require.Equal(t, 80, ansi.StringWidth(l), "row %d", i)
	}
	return m, out
}

// key presses k and delivers the message its action sends.
func key(t *testing.T, m tea.Model, k string) tea.Model {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key(k))
	if cmd != nil {
		m, _ = m.Update(cmd())
	}
	return m
}

func TestCursorRowsNameTheirModes(t *testing.T) {
	_, out := frame(t, newModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	for _, mode := range []string{"blinking", "steady", "hidden"} {
		require.Contains(t, out, mode)
	}
}

func TestCursorTabMovesFocus(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = key(t, m, "tab")
	cs := m.(model).cursors
	require.False(t, cs[0].Focused())
	require.True(t, cs[1].Focused())
}

func TestCursorMCyclesTheFocusedMode(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = key(t, m, "m")
	_, out := frame(t, m, nil)
	require.Equal(t, 2, strings.Count(out, "steady"), "the first row moved on to steady")
}
