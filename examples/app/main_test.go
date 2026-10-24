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

// press sends key and runs what it returns, the way the program would.
func press(m tea.Model, key string) tea.Model {
	m, cmd := m.Update(bezeltest.Key(key))
	return run(m, cmd)
}

// run feeds cmd's messages back into m, following batches. A command that
// does not answer at once is a timer (a caret blink, a TTL) and is dropped.
func run(m tea.Model, cmd tea.Cmd) tea.Model {
	if cmd == nil {
		return m
	}
	got := make(chan tea.Msg, 1)
	go func() { got <- cmd() }()
	select {
	case msg := <-got:
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				m = run(m, c)
			}
			return m
		}
		if msg == nil {
			return m
		}
		m, cmd = m.Update(msg)
		return run(m, cmd)
	case <-time.After(20 * time.Millisecond):
		return m
	}
}

func sized(t *testing.T, readOnly bool) tea.Model {
	t.Helper()
	var m tea.Model = newModel(readOnly)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 90, Height: 24})
	return m
}

func TestDemoRendersAtCommonSizes(t *testing.T) {
	out := frame(t, newModel(false), 90, 24)
	require.Contains(t, out, "record the demo")
	require.Contains(t, out, "1/4 done")
	require.Contains(t, out, "write the release notes", "the detail shows the first task")

	narrow := frame(t, newModel(false), 60, 20)
	require.Contains(t, narrow, "record the demo")
	require.NotContains(t, narrow, "state", "the detail collapsed away")
}

func TestSpaceMarksTheTaskDone(t *testing.T) {
	m := press(press(sized(t, false), "down"), "space")
	require.Contains(t, ansi.Strip(m.View().Content), "2/4 done")
}

func TestNewAsksForATitleAndAddsIt(t *testing.T) {
	m := press(sized(t, false), "n")
	require.Contains(t, ansi.Strip(m.View().Content), "New task")
	for _, k := range strings.Split("ship it", "") {
		m = press(m, k)
	}
	m = press(m, "enter")
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "ship it")
	require.Contains(t, out, "1/5 done")
}

func TestDeleteAsksFirst(t *testing.T) {
	m := press(sized(t, false), "d")
	require.Contains(t, ansi.Strip(m.View().Content), "Delete task")
	m = press(m, "y")
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "deleted write the release notes")
	require.Contains(t, out, "0/3 done")
}

func TestReadOnlyHidesDelete(t *testing.T) {
	m := sized(t, true)
	require.NotContains(t, ansi.Strip(m.View().Content), "delete")
	m = press(m, "d")
	require.NotContains(t, ansi.Strip(m.View().Content), "Delete task")
}
