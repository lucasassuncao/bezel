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

func TestShellTabsAndStatus(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Contains(t, out, "▶ main.go")
	require.Contains(t, out, "4 items")

	m, cmd := m.Update(bezeltest.Key("s"))
	require.NotNil(t, cmd, "s runs the save action")
	_, out = frame(t, m, cmd())
	require.Contains(t, out, "saved main.go")

	_, out = frame(t, m, bezeltest.Key("tab"))
	require.Contains(t, out, "▶ ideas")
}

func TestShellBusyShowsText(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(bezeltest.Key("b"))
	_, out := frame(t, m, cmd())
	require.Contains(t, out, "indexing files")
}

// at is the screen cell where text first shows.
func at(t *testing.T, out, text string) (x, y int) {
	t.Helper()
	for y, row := range strings.Split(out, "\n") {
		if i := strings.Index(row, text); i >= 0 {
			return ansi.StringWidth(row[:i]), y
		}
	}
	t.Fatalf("%q is not on screen", text)
	return 0, 0
}

func click(x, y int) tea.MouseClickMsg { return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft} }

func TestShellTakesTheMouse(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	m, out = frame(t, m, click(at(t, out, "README.md")))
	require.Contains(t, out, "▶ README.md", "a click moves the cursor")
	m, cmd := m.Update(click(at(t, out, "▶ README.md"))) // the detail pane names it too
	require.NotNil(t, cmd, "the status's timer")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "opened README.md", "a second click opens it")

	m, _ = frame(t, m, click(at(t, out, "notes")))
	require.Equal(t, 1, m.(model).sh.ActiveTab(), "a click on a tab shows it")

	m, cmd = m.Update(bezeltest.Key("d"))
	m, cmd = m.Update(cmd()) // the action's message, which opens the confirm
	m, out = frame(t, m, cmd())
	_, cmd = m.Update(click(at(t, out, "Yes")))
	require.NotNil(t, cmd, "a click on Yes answers the confirm")
	var deleted bool
	for _, c := range cmd().(tea.BatchMsg) {
		if msg, ok := c().(deletedMsg); ok {
			deleted = msg.item == "ideas"
		}
	}
	require.True(t, deleted)
}
