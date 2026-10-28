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

func typed(m tea.Model, s string) tea.Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

// enter submits and renders the status it sets.
func enter(t *testing.T, m tea.Model) string {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key("enter"))
	require.NotNil(t, cmd)
	return ansi.Strip(m.View().Content)
}

func TestInputSubmitsEveryField(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = typed(m, "ana")
	m, _ = m.Update(bezeltest.Key("down"))
	m = typed(m, "ro")
	m, _ = m.Update(bezeltest.Key("tab")) // completes rollback
	m, _ = m.Update(bezeltest.Key("down"))
	m = typed(m, "s3cret")
	_, out := frame(t, m, nil)
	require.NotContains(t, out, "s3cret", "the password shows only its mask")
	out = enter(t, m)
	t.Log("\n" + out)
	require.Contains(t, out, `ana runs "rollback" (password: 6 characters)`)
}

// Leaving the name empty says so under it, and enter refuses to submit.
func TestInputRequiresAName(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, out := frame(t, m, bezeltest.Key("down"))
	require.Contains(t, out, "a name is required")
	out = enter(t, m)
	require.NotContains(t, out, "runs")
}

// esc quits, since q is text.
func TestInputQuitsOnEsc(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = typed(m, "q")
	_, cmd := m.Update(bezeltest.Key("esc"))
	require.NotNil(t, cmd)
	require.Equal(t, tea.Quit(), cmd())
}
