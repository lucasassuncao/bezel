package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

func TestSoloDrawsOnlyTheKindItWasGiven(t *testing.T) {
	s, err := newSolo("typewriter")
	require.NoError(t, err)
	var m tea.Model = s
	m, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 12})
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, headline)
	require.NotContains(t, out, "Downloads", "the marquee's row is not drawn")
	require.NotContains(t, out, "password", "the shake's row is not drawn")
}

func TestSoloNamesTheKindsWhenGivenAWrongOne(t *testing.T) {
	_, err := newSolo("nope")
	require.ErrorContains(t, err, "typewriter")
	require.ErrorContains(t, err, "burn")
}

func TestEveryKindTheGalleryDrawsHasASolo(t *testing.T) {
	for _, kind := range newModel().kinds() {
		_, err := newSolo(kind)
		require.NoError(t, err, kind)
	}
}

// A recording has no use for a legend of keys, and the keys keep working.
func TestSoloDrawsNoLegendAndKeepsItsKeys(t *testing.T) {
	s, err := newSolo("typewriter")
	require.NoError(t, err)
	var m tea.Model = s
	m, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 12})
	out := ansi.Strip(m.View().Content)
	require.NotContains(t, out, "quit")
	require.NotContains(t, out, "again")
	_, cmd := m.Update(bezeltest.Key("r"))
	require.NotNil(t, cmd, "r still plays it again")
	_, cmd = m.Update(bezeltest.Key("q"))
	require.NotNil(t, cmd, "q still quits")
	require.IsType(t, tea.QuitMsg{}, cmd())
}
