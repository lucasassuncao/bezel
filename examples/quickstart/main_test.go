package main

import (
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

// frame is m at w×h as text, every row checked to fill the width.
func frame(t *testing.T, m tea.Model, w, h int) string {
	t.Helper()
	m, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	out := ansi.Strip(m.View().Content)
	for i, l := range strings.Split(out, "\n") {
		require.Equal(t, w, ansi.StringWidth(l), "row %d", i)
	}
	return out
}

func TestQuickstartShowsTheFirstItem(t *testing.T) {
	out := frame(t, newModel(), 70, 14)
	require.Contains(t, out, "Fruit")
	require.Contains(t, out, "▶ apple")
	require.Contains(t, out, "red and crisp")
	require.Contains(t, out, "[q] quit")
}

func TestQuickstartDownMovesTheDetail(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 14})
	m, _ = m.Update(bezeltest.Key("down"))
	out := frame(t, m, 70, 14)
	require.Contains(t, out, "▶ banana")
	require.Contains(t, out, "yellow and soft")
}

func TestQuickstartQuits(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 14})
	_, cmd := m.Update(bezeltest.Key("q"))
	require.NotNil(t, cmd)
	require.IsType(t, tea.QuitMsg{}, cmd())
}

// The README shows this program whole; this keeps its copy the same as
// main.go, comments above `package main` aside.
func TestReadmeShowsThisProgram(t *testing.T) {
	src, err := os.ReadFile("main.go")
	require.NoError(t, err)
	readme, err := os.ReadFile("../../README.md")
	require.NoError(t, err)
	code := strings.ReplaceAll(string(src), "\r\n", "\n")
	_, after, found := strings.Cut(code, "package main")
	require.True(t, found)
	code = "package main" + after
	doc := strings.ReplaceAll(string(readme), "\r\n", "\n")
	require.Contains(t, doc, "```go\n"+code+"```", "README.md's quickstart differs from examples/quickstart/main.go")
}
