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

func TestCalendarMovesAndChooses(t *testing.T) {
	var m tea.Model = newModel()
	require.Nil(t, m.Init())
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 16})
	t.Log("\n" + out)
	require.Contains(t, out, "November 2026")
	require.Contains(t, out, "Mo Tu We Th Fr Sa Su")
	require.Contains(t, out, "Monday, 2 November 2026")

	m, _ = frame(t, m, bezeltest.Key("down"))
	m, _ = frame(t, m, bezeltest.Key("pgdown"))
	m, out = frame(t, m, bezeltest.Key("enter"))
	t.Log("\n" + out)
	require.Contains(t, out, "December 2026")
	require.Contains(t, out, "chosen  2026-12-09")
	require.Contains(t, out, "37 days from today")

	m, _ = frame(t, m, bezeltest.Key("pgup"))
	_, out = frame(t, m, bezeltest.Key("pgup"))
	require.Contains(t, out, "Monday, 2 November 2026", "held at today, the range's start")
}

// click is a left click at screen cell (x, y), through the shell as a
// program would deliver it.
func click(t *testing.T, m tea.Model, x, y int) tea.Model {
	t.Helper()
	m, cmd := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	require.NotNil(t, cmd, "the shell turns it into a ClickMsg")
	m, _ = m.Update(cmd())
	return m
}

func TestCalendarTakesClicks(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 16})
	m = click(t, m, 8, 7) // Wednesday the 11th: the third week, third column
	m = click(t, m, 8, 7)
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "chosen  2026-11-11")
}
