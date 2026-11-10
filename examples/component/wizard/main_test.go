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

func TestWizardIsFilledInPageByPage(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	t.Log("\n" + out)
	require.Contains(t, out, "1 Project ── 2 Tools ── 3 Review")

	m = press(m, "pgdown")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "the container needs a name")

	m = press(m, "a", "p", "i", "tab", "enter", "down", "down", "enter", "pgdown")
	m = press(m, "enter", "g", "i", "space", "backspace", "backspace", "d", "o", "c", "space")
	_, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "/ doc")
	m = press(m, "enter", "pgdown")
	m = press(m, "right", "tab", "space", "pgdown")
	_, out = frame(t, m, nil)
	t.Log("\n" + out)
	require.Contains(t, out, "api: alpine with git+docker, zsh, starts now")
}

func TestToolsWillNotBeSkipped(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = press(m, "x", "pgdown", "pgdown")
	_, out := frame(t, m, nil)
	require.Contains(t, out, "check at least one tool")

	m = press(m, "pgup")
	_, out = frame(t, m, nil)
	require.Contains(t, out, "[x")
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
