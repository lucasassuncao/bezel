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
// command is not run: the count's timer would make the test wait on it.
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

func TestProgressCountsUpAndRestarts(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	for range 10 {
		m, _ = m.Update(tickMsg{})
	}
	m, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "10/40")
	require.Contains(t, out, "█")

	m = key(t, m, "r")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "0/40")
}

func TestProgressStepperMovesAndFolds(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = key(t, m, "n")
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "✓ Template ─ ● Features ─ ○ Ports ─ ○ Review")
	require.Contains(t, out, "Step 2 of 4 · Features")

	for range 5 {
		m = key(t, m, "n")
	}
	_, out = frame(t, m, nil)
	require.Contains(t, out, "All 4 steps done")
}
