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

func typed(m tea.Model, s string) tea.Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func TestTagsFilterThePackages(t *testing.T) {
	var m tea.Model = newModel()
	require.NotNil(t, m.Init(), "the caret blinks")
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 20})
	require.Contains(t, out, "8 of 8")

	m = typed(m, "tu")
	m, _ = m.Update(bezeltest.Key("tab")) // completes tui
	m, _ = m.Update(bezeltest.Key("enter"))
	m = typed(m, "st")
	m, _ = m.Update(bezeltest.Key("tab"))
	m, out = frame(t, m, bezeltest.Key("enter"))
	t.Log("\n" + out)
	require.Contains(t, out, " tui   style ")
	require.Contains(t, out, "lipgloss")
	require.NotContains(t, out, "glamour", "style but not tui")
	require.Contains(t, out, "1 of 8")

	_, out = frame(t, m, bezeltest.Key("backspace"))
	require.Contains(t, out, "3 of 8", "back to tui alone")
}

func TestTagsQuitOnEsc(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 20})
	m = typed(m, "q")
	_, cmd := m.Update(bezeltest.Key("esc"))
	require.NotNil(t, cmd)
	require.Equal(t, tea.Quit(), cmd())
}
