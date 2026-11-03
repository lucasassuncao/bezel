package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

// frame sends msg and renders; every row must be exactly the window wide.
func frame(t *testing.T, m tea.Model, msg tea.Msg) (tea.Model, string) {
	t.Helper()
	m, _ = m.Update(msg)
	out := ansi.Strip(m.View().Content)
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		require.Equal(t, ansi.StringWidth(lines[0]), ansi.StringWidth(l), "row %d", i)
	}
	return m, out
}

func press(t *testing.T, m tea.Model, keys ...string) tea.Model {
	t.Helper()
	for _, k := range keys {
		m, _ = frame(t, m, bezeltest.Key(k))
	}
	return m
}

func TestEnterGoesDownAndThePathGoesBackUp(t *testing.T) {
	var m tea.Model = newModel()
	require.Nil(t, m.Init())
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 16})
	t.Log("\n" + out)
	require.Contains(t, out, "apps/")

	m = press(t, m, "enter", "enter", "enter", "enter")
	_, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "kv › apps › billing › production › database")
	require.Contains(t, out, "password")

	m = press(t, m, "enter")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "opened  kv/apps/billing/production/database/password")

	m = press(t, m, "tab", "left", "left", "left", "enter")
	_, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "kv › apps", "back up to apps")
	require.NotContains(t, out, "billing ›")
	require.Contains(t, out, "search/")
}

func TestANarrowWindowGivesThePathsMiddleWay(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 36, Height: 16})
	m = press(t, m, "enter", "enter", "enter", "enter")
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "kv › … › production › database")
}

// click is a left click at screen cell (x, y), through the shell as a
// program would deliver it: a click on a pane without focus also focuses it.
func click(t *testing.T, m tea.Model, x, y int) tea.Model {
	t.Helper()
	m, cmd := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	require.NotNil(t, cmd, "the shell turns it into a ClickMsg")
	msgs := []tea.Msg{cmd()}
	if b, ok := msgs[0].(tea.BatchMsg); ok {
		msgs = msgs[:0]
		for _, c := range b {
			msgs = append(msgs, c())
		}
	}
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m
}

func TestAClickOnAPartThenAgainGoesUpToIt(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 16})
	m = press(t, m, "enter", "enter")
	m = click(t, m, 2, 3) // "kv", the first part
	m = click(t, m, 2, 3)
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.NotContains(t, out, "kv › apps")
	require.Contains(t, out, "infra/")
}
