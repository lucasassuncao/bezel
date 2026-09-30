package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func frame(t *testing.T, m tea.Model, w, h int) string {
	t.Helper()
	m, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	out := ansi.Strip(m.View().Content)
	for i, l := range strings.Split(out, "\n") {
		require.Equal(t, w, ansi.StringWidth(l), "row %d", i)
	}
	return out
}

func TestDemoRendersAtCommonSizes(t *testing.T) {
	var m tea.Model = newModel()
	out := frame(t, m, 80, 24)
	t.Log("\n" + out)
	require.Contains(t, out, "fruits")
	require.Contains(t, out, "> banana")
	require.NotContains(t, out, "[d] delete") // read-only session hides it

	narrow := frame(t, m, 60, 20)
	require.NotContains(t, narrow, "> banana") // list collapsed away
	require.Contains(t, narrow, "selected: banana")
}

func TestDemoConfirmFloatsOverTheList(t *testing.T) {
	t.Setenv("DEMO_RW", "1") // delete needs a session that may write
	var m tea.Model = newModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	require.NotNil(t, cmd, "d runs the delete action")
	m, _ = m.Update(cmd())
	out := ansi.Strip(m.View().Content)
	t.Log("\n" + out)
	require.Contains(t, out, "Delete banana?")
	require.Contains(t, out, "> banana")
}
