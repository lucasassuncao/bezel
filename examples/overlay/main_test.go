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

// press sends one key and runs what it returns, the way the program would.
func press(t *testing.T, m tea.Model, key string) tea.Model {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key(key))
	return run(m, cmd)
}

// run feeds cmd's messages back into m, following batches. A command that
// does not answer at once is a timer (a cursor blink, a TTL) and is dropped.
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

func TestOverlayConfirmAnswersTheApp(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = press(t, m, "down")
	m = press(t, m, "enter")
	m, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "Delete notes.txt?")

	m = press(t, m, "y")
	_, out = frame(t, m, nil)
	require.NotContains(t, out, "Delete notes.txt?")
	require.Contains(t, out, "deleted notes.txt")
}

func TestOverlayPromptRefusesAnEmptyName(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = press(t, m, "down")
	m = press(t, m, "down")
	m = press(t, m, "enter")
	m, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "a name is required")

	for _, k := range []string{"a", ".", "t", "x", "t", "enter"} {
		m = press(t, m, k)
	}
	_, out = frame(t, m, nil)
	require.Contains(t, out, "created a.txt")
}

func TestOverlayEscClosesAlert(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = press(t, m, "enter")
	m, out := frame(t, m, nil)
	require.Contains(t, out, "Disk almost full")
	m = press(t, m, "esc")
	_, out = frame(t, m, nil)
	require.NotContains(t, out, "Disk almost full")
}

func TestOverlayPagerScrollsALongOutput(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	for range 4 {
		m = press(t, m, "down")
	}
	m = press(t, m, "enter")
	m, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "go test ./...")
	require.Contains(t, out, "TestCase01")
	require.NotContains(t, out, "ok  ")

	m = press(t, m, "end")
	m, out = frame(t, m, nil)
	require.Contains(t, out, "example.com/app", "end reaches the last line")

	m = press(t, m, "esc")
	_, out = frame(t, m, nil)
	require.NotContains(t, out, "go test ./...")
}
