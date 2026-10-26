package main

import (
	"os"
	"path/filepath"
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

// sized is a picker over a small tree, listed, at 80x24.
func sized(t *testing.T) tea.Model {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "docs"), 0o750))
	var m tea.Model = newModel(dir)
	m, _ = m.Update(m.Init()())
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

func TestFilepickerListsTheDirectory(t *testing.T) {
	_, out := frame(t, sized(t), nil)
	t.Log("\n" + out)
	require.Contains(t, out, "docs")
	require.Contains(t, out, "8B go.mod")
}

func TestFilepickerPicksAFile(t *testing.T) {
	m, _ := frame(t, sized(t), bezeltest.Key("down"))
	_, out := frame(t, m, bezeltest.Key("enter"))
	require.Contains(t, out, "picked go.mod")
}

func TestFilepickerPShowsThePermissions(t *testing.T) {
	m, _ := sized(t).Update(permsMsg{})
	_, out := frame(t, m, nil)
	require.Contains(t, out, "drwx")
}
