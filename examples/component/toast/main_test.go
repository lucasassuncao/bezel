package main

import (
	"strings"
	"testing"
	"time"

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

// press presses k and delivers the message its action sends.
func press(t *testing.T, m tea.Model, k string) tea.Model {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key(k))
	require.NotNil(t, cmd)
	m, _ = m.Update(cmd())
	return m
}

func TestToastsStackInTheCornerAndLeave(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = press(t, m, "s")
	m = press(t, m, "e")
	start := time.Unix(10, 0)
	toasts := m.(model).toasts
	m, _ = frame(t, m, toasts.FrameAt(start))
	m, out := frame(t, m, m.(model).toasts.FrameAt(start.Add(time.Second)))
	t.Log("\n" + out)
	require.Contains(t, out, "Saved notes.md")
	require.Contains(t, out, "Could not reach")
	require.Less(t, strings.Index(out, "Could not reach"), strings.Index(out, "Saved notes.md"), "the newest on top")
	require.Contains(t, out, "A toast says", "the page shows beside them")

	_, out = frame(t, m, m.(model).toasts.FrameAt(start.Add(10*time.Second)))
	require.NotContains(t, out, "Saved notes.md")
}
