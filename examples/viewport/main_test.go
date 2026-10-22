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

// key presses k and delivers the message its action sends.
func key(t *testing.T, m tea.Model, k string) tea.Model {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key(k))
	if cmd != nil {
		m, _ = m.Update(cmd())
	}
	return m
}

func sized(t *testing.T) tea.Model {
	t.Helper()
	m, _ := frame(t, newModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

func TestViewportShowsTheLogWithLineNumbers(t *testing.T) {
	_, out := frame(t, sized(t), nil)
	t.Log("\n" + out)
	require.Contains(t, out, "   1 │ 12:00:00 GET")
}

func TestViewportNJumpsToTheNextError(t *testing.T) {
	_, out := frame(t, key(t, sized(t), "n"), nil)
	require.Contains(t, out, "  74 │")
	require.Contains(t, out, "ERROR")
}

func TestViewportWTogglesWrap(t *testing.T) {
	m := key(t, sized(t), "w")
	require.True(t, m.(model).vp.Wrap())
	m = key(t, m, "w")
	require.False(t, m.(model).vp.Wrap())
}
