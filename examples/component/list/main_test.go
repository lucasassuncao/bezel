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

func TestListChoosesAcrossASection(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "Fruit")
	require.Contains(t, out, "enter shows")

	// Four downs: past pineapple, over the spacer and heading, onto carrot.
	m = typeKeys(m, "down", "down", "down", "down", "enter")
	_, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "41 kcal")
}

func TestListFilterTakesEveryKey(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = typeKeys(m, "/", "q", "u")
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.NotContains(t, out, "banana", "q went to the filter, not to quit")

	m = typeKeys(m, "esc", "esc")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "banana")
}

func TestListTurnsASpinnerOnABusyRow(t *testing.T) {
	var m tea.Model = newModel()
	require.NotNil(t, m.Init(), "orange starts busy")
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Contains(t, out, "  ⠋  [ ] orange", "the spinner takes the mark")

	m, cmd := m.Update(bezeltest.Key("b"))
	_, out = frame(t, m, cmd())
	t.Log("\n" + out)
	require.Contains(t, out, "▶ ⠋  [ ] banana", "b marks the cursor row")
}

func TestListChecksRowsForABatch(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = frame(t, m, bezeltest.Key("space"))
	m, _ = frame(t, m, bezeltest.Key("down"))
	_, out := frame(t, m, bezeltest.Key("space"))
	require.Contains(t, out, "[x] banana")
	require.Contains(t, out, "checked: banana, mango")
	require.Contains(t, out, "2 checked")
}
