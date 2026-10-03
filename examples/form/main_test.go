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

// press sends keys and delivers what each answers at once; a caret blink or a
// status timer does not answer and is dropped.
func press(m tea.Model, keys ...string) tea.Model {
	for _, k := range keys {
		var cmd tea.Cmd
		m, cmd = m.Update(bezeltest.Key(k))
		if cmd == nil {
			continue
		}
		got := make(chan tea.Msg, 1)
		go func() { got <- cmd() }()
		select {
		case msg := <-got:
			if msg != nil {
				m, _ = m.Update(msg)
			}
		case <-time.After(20 * time.Millisecond):
		}
	}
	return m
}

func TestFormIsFilledInAndSubmitted(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "› Name")
	require.Contains(t, out, "‹ ubuntu ›")

	m = press(m, "a", "p", "i")
	m = press(m, "tab", "enter", "down", "down", "enter") // image: alpine
	m = press(m, "tab", "right")                          // shell: zsh
	m = press(m, "tab", "space")                          // docker on
	m = press(m, "tab", "space")                          // forward the port
	m = press(m, "tab", "enter")                          // Create
	_, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "created api: alpine, zsh, docker true, port 3000 true")
}

func TestTypingQDoesNotQuit(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(bezeltest.Key("q"))
	if cmd != nil {
		require.NotEqual(t, tea.QuitMsg{}, cmd(), "q is a letter in the name, not quit")
	}
	_, out := frame(t, m, nil)
	require.Contains(t, out, "[q")
}
