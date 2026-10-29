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

func TestTheAccordionOpensTheSectionUnderTheCursor(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Contains(t, out, "▸ What is bezel?")
	require.NotContains(t, out, "The frame around")

	m, _ = m.Update(bezeltest.Key("down"))
	m, cmd := m.Update(bezeltest.Key("enter"))
	require.NotNil(t, cmd, "opening asks for frames")
	t.Setenv("BEZEL_MOTION", "20") // the frames come quickly and the fold lands in a few
	for i := 0; cmd != nil && i < 100; i++ {
		m, cmd = m.Update(cmd())
	}
	_, out = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "▾ Does it need a mouse?")
	require.Contains(t, out, "Every action has a key")
}
