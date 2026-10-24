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
	m, cmd := m.Update(msg)
	if cmd != nil {
		if next := cmd(); next != nil {
			m, _ = m.Update(next)
		}
	}
	out := ansi.Strip(m.View().Content)
	for i, l := range strings.Split(out, "\n") {
		require.Equal(t, 80, ansi.StringWidth(l), "row %d", i)
	}
	return m, out
}

func TestIconShowsBothSetsAndSwapsTheList(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 40})
	t.Log("\n" + out)
	require.Contains(t, out, "OK        ✓   +")
	require.Contains(t, out, "✗ neovim")
	require.Contains(t, out, "PACKAGES · UNICODE")
	require.Contains(t, out, "├─ scoop")
	require.Contains(t, out, "│  └─ bucket extras")

	_, out = frame(t, m, bezeltest.Key("a"))
	t.Log("\n" + out)
	require.Contains(t, out, "x neovim")
	require.Contains(t, out, "PACKAGES · ASCII")
	require.Contains(t, out, "|- scoop")
	require.Contains(t, out, "`- github")
}
