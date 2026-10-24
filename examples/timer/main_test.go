package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/component/timer"
)

// frame sends msg and renders; every row must be exactly 80 columns wide. A
// command is not run: the ticks would make the test wait on them.
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

func TestTimerShowsTheTimeLeft(t *testing.T) {
	_, out := frame(t, newModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "left 10.0s")
}

func TestTimerSaysWhenTimeIsUp(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	_, out := frame(t, m, timer.TimeoutMsg{ID: m.(model).tm.ID()})
	require.Contains(t, out, "time's up")
}

func TestTimerSpaceStopsIt(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(m.Init()())
	require.True(t, m.(model).tm.Running(), "Init starts it")
	m = key(t, m, "space")
	require.False(t, m.(model).tm.Running())
}
