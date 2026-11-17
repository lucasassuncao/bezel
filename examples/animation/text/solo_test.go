package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
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
