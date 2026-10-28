package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

// frame sends msg and renders; every row must be exactly 100 columns wide. A
// command is not run: the feed's timer would make the test wait on it.
func frame(t *testing.T, m tea.Model, msg tea.Msg) (tea.Model, string) {
	t.Helper()
	m, _ = m.Update(msg)
	out := ansi.Strip(m.View().Content)
	for i, l := range strings.Split(out, "\n") {
		require.Equal(t, 100, ansi.StringWidth(l), "row %d", i)
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

func TestChartDrawsEveryKind(t *testing.T) {
	var m tea.Model = newModel()
	_, out := frame(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	t.Log("\n" + out)
	for _, want := range []string{"AREA", "SPARK", "BARS", "HEAT", "DONUT", "api", "billing", "40 samples"} {
		require.Contains(t, out, want)
	}
}

// Each tick adds a sample; space stops the feed and r starts it over.
func TestChartFeedsPausesAndRestarts(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	for range 5 {
		m, _ = m.Update(tickMsg{})
	}
	_, out := frame(t, m, nil)
	require.Contains(t, out, "45 samples, live")

	m = key(t, m, "space")
	m, cmd := m.Update(tickMsg{})
	require.Nil(t, cmd, "a paused feed asks for no tick")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "45 samples, paused")

	m = key(t, m, "r")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "40 samples")
}

// The made-up load stays inside the chart's range.
func TestLoadStaysInRange(t *testing.T) {
	for n := range 1000 {
		v := load(n)
		require.True(t, v >= 0 && v <= top, "load(%d) = %v", n, v)
	}
}
