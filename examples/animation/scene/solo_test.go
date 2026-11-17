package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

func TestSoloDrawsOnlyTheSceneItWasGiven(t *testing.T) {
	s, err := newSolo("snow")
	require.NoError(t, err)
	var m tea.Model = s
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 14})
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "bezel ✦ snow")
	require.NotContains(t, out, "n next", "no way to change the scene")
}

func TestSoloNamesTheScenesWhenGivenAWrongOne(t *testing.T) {
	_, err := newSolo("nope")
	require.ErrorContains(t, err, "rain")
	require.ErrorContains(t, err, "life")
}

func TestEverySceneTheGalleryPlaysHasASolo(t *testing.T) {
	for _, k := range sceneKinds {
		_, err := newSolo(k.name)
		require.NoError(t, err, k.name)
	}
}

// A recording has no use for a legend of keys, and the keys keep working.
func TestSoloDrawsNoLegendAndKeepsItsKeys(t *testing.T) {
	s, err := newSolo("snow")
	require.NoError(t, err)
	var m tea.Model = s
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 14})
	out := ansi.Strip(m.View().Content)
	require.NotContains(t, out, "quit")
	require.NotContains(t, out, "again")
	_, cmd := m.Update(bezeltest.Key("q"))
	require.NotNil(t, cmd, "q still quits")
	require.IsType(t, tea.QuitMsg{}, cmd())
}
