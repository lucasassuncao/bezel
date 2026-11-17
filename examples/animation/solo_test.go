package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

func TestNoTwoAnimationsShareAName(t *testing.T) {
	seen := map[string]bool{}
	for _, k := range newModel().kinds() {
		require.False(t, seen[k], k)
		seen[k] = true
	}
}

func TestEveryAnimationTheGalleryDrawsHasASolo(t *testing.T) {
	for _, k := range newModel().kinds() {
		s, err := newSolo(k)
		require.NoError(t, err, k)
		var m tea.Model = s
		m, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 30})
		require.NotEmpty(t, ansi.Strip(m.View().Content), k)
	}
}

func TestSoloDrawsOnlyTheAnimationItWasGiven(t *testing.T) {
	s, err := newSolo("count-up")
	require.NoError(t, err)
	var m tea.Model = s
	m, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 20})
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "files")
	require.NotContains(t, out, "waiting for the server", "the pulse's row is not drawn")
}

func TestSoloNamesTheAnimationsWhenGivenAWrongOne(t *testing.T) {
	_, err := newSolo("nope")
	require.ErrorContains(t, err, "easing")
	require.ErrorContains(t, err, "squash")
}

// A recording has no use for a legend of keys, and the keys keep working.
func TestSoloDrawsNoLegendAndKeepsItsKeys(t *testing.T) {
	s, err := newSolo("fade")
	require.NoError(t, err)
	var m tea.Model = s
	m, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 20})
	out := ansi.Strip(m.View().Content)
	require.NotContains(t, out, "quit")
	require.NotContains(t, out, "again")
	_, cmd := m.Update(bezeltest.Key("r"))
	require.NotNil(t, cmd, "r still plays it again")
	_, cmd = m.Update(bezeltest.Key("q"))
	require.NotNil(t, cmd, "q still quits")
	require.IsType(t, tea.QuitMsg{}, cmd())
}
