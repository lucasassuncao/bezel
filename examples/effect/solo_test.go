package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

func TestSlugIsTheNameAFlagTakes(t *testing.T) {
	require.Equal(t, "wipe-radial", slug("wipe radial"))
	require.Equal(t, "push", slug("push ←"))
	require.Equal(t, "iris", slug("iris"))
}

func TestNoTwoEffectsShareAName(t *testing.T) {
	seen := map[string]bool{}
	for _, k := range kinds() {
		require.False(t, seen[k.name], k.name)
		seen[k.name] = true
	}
}

func TestEveryEffectTheGalleryPlaysHasASolo(t *testing.T) {
	for _, k := range kinds() {
		s, err := newSolo(k.name)
		require.NoError(t, err, k.name)
		var m tea.Model = s
		m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
		require.NotEmpty(t, ansi.Strip(m.View().Content), k.name)
	}
}

func TestSoloDrawsOnlyTheEffectItWasGiven(t *testing.T) {
	s, err := newSolo("hue-shift")
	require.NoError(t, err)
	var m tea.Model = s
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "bezel ✦ effects")
	require.NotContains(t, out, "pattern", "no tab rows around the card")
}

func TestSoloNamesTheEffectsWhenGivenAWrongOne(t *testing.T) {
	_, err := newSolo("nope")
	require.ErrorContains(t, err, "fade-in")
	require.ErrorContains(t, err, "trace")
}

// A recording has no use for a legend of keys, and the keys keep working.
func TestSoloDrawsNoLegendAndKeepsItsKeys(t *testing.T) {
	s, err := newSolo("glitch")
	require.NoError(t, err)
	var m tea.Model = s
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	out := ansi.Strip(m.View().Content)
	require.NotContains(t, out, "quit")
	require.NotContains(t, out, "again")
	_, cmd := m.Update(bezeltest.Key("r"))
	require.NotNil(t, cmd, "r still plays it again")
	_, cmd = m.Update(bezeltest.Key("q"))
	require.NotNil(t, cmd, "q still quits")
	require.IsType(t, tea.QuitMsg{}, cmd())
}
