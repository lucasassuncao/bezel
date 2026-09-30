package main

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/shell"
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

// typeKeys presses each key in turn and runs what it returns.
func typeKeys(m tea.Model, keys ...string) tea.Model {
	for _, k := range keys {
		var cmd tea.Cmd
		m, cmd = m.Update(bezeltest.Key(k))
		m = run(m, cmd)
	}
	return m
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

func TestTreeExpandsAndCollapses(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "▸ cmd")
	require.NotContains(t, out, "app")

	m = typeKeys(m, "right")
	m, out = frame(t, m, nil)
	require.Contains(t, out, "▾ cmd")
	require.Contains(t, out, "▸ app")

	m = typeKeys(m, "left")
	_, out = frame(t, m, nil)
	require.NotContains(t, out, "app")
}

func TestTreeRevealOpensTheWay(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = typeKeys(m, strings.Split(":reveal internal/tui/keys/keys.go", "")...)
	m = typeKeys(m, "enter")
	_, out := frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "keys.go")
	require.Contains(t, out, "internal/tui/keys/keys.go", "the detail follows the cursor")
}

func TestTreeKeysDoNotCollide(t *testing.T) {
	require.NoError(t, shell.Check(actions()...))
}
