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

// deploy logs n lines and hands them to the pane, as the program would.
func deploy(m tea.Model, n int) tea.Model {
	mm := m.(model)
	for i := range n {
		mm.emit(i)
	}
	next, _ := mm.Update(mm.pane.Init()())
	return next
}

func TestLogPaneFollowsAndHoldsStill(t *testing.T) {
	var m tea.Model = newModel()
	require.NotNil(t, m.Init())
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 16})

	m = deploy(m, 32) // four rounds: the last is "traffic shifted"
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "traffic shifted", "the newest line shows")
	require.Contains(t, out, "following")

	m, _ = frame(t, m, bezeltest.Key("pgup"))
	m = deploy(m, 3)
	m, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "held")
	require.Contains(t, out, "new lines", "what arrived meanwhile is counted")

	_, out = frame(t, m, bezeltest.Key("end"))
	require.Contains(t, out, "following")
}

func TestLogPaneStepsTheDeploy(t *testing.T) {
	m := newModel()
	msg := m.emit(len(events) - 1)
	require.Equal(t, step{len(events)}, msg)
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "the next line is due")
}
