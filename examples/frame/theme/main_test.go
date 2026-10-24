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

func TestThemeStepsThroughPresets(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "mint")
	require.Contains(t, out, "Success")

	m, cmd := m.Update(bezeltest.Key("n"))
	m, out = frame(t, m, cmd())
	require.Contains(t, out, "banana")

	m, cmd = m.Update(bezeltest.Key("p"))
	m, _ = m.Update(cmd())
	m, cmd = m.Update(bezeltest.Key("p"))
	_, out = frame(t, m, cmd())
	require.Contains(t, out, "plain", "p wraps from the first preset to the last")
}

func TestEveryPresetResolves(t *testing.T) {
	for i := range presets {
		require.NotPanics(t, func() { newModel().apply(i) })
	}
}
