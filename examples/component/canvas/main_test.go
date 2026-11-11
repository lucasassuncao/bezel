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
// command is not run: the tick's timer would make the test wait on it.
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

func TestCanvasDrawsTheSystem(t *testing.T) {
	_, out := frame(t, newModel(), tea.WindowSizeMsg{Width: 100, Height: 30})
	for _, want := range []string{"sun", "ember", "tide", "dusk", "orbits · live · zoom 13"} {
		require.Contains(t, out, want)
	}
	require.Contains(t, out, "⠁", "braille dots are drawn")
}

// The first frame, whole: a change to the canvas shows here as a diff.
func TestCanvasFirstFrame(t *testing.T) {
	_, out := frame(t, newModel(), tea.WindowSizeMsg{Width: 100, Height: 30})
	bezeltest.Golden(t, out)
}

func TestCanvasMovesPausesAndZooms(t *testing.T) {
	var m tea.Model = newModel()
	m, before := frame(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	for range 10 {
		m, _ = m.Update(tickMsg{})
	}
	_, after := frame(t, m, nil)
	require.NotEqual(t, before, after, "the planets moved")

	m = key(t, m, "space")
	m, cmd := m.Update(tickMsg{})
	require.Nil(t, cmd, "a paused system asks for no tick")
	_, out := frame(t, m, nil)
	require.Contains(t, out, "paused")

	m = key(t, m, "+")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "zoom 10")
	for range 20 {
		m = key(t, m, "-")
	}
	_, out = frame(t, m, nil)
	require.Contains(t, out, "zoom 40", "zoom stops at its widest")
}
