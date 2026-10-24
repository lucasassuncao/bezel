package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTheLightTabShowsTheCardTheDiffAndTheGhost(t *testing.T) {
	_, out := onTab(t, 1)
	t.Log("\n" + out)
	for _, want := range []string{"bezel ✦ light", "func sum", "report.pdf"} {
		require.Contains(t, out, want)
	}
}

// Every light ends, and leaves the card as it found it.
func TestEveryLightLeavesTheCardWhole(t *testing.T) {
	m, _ := onTab(t, 1)
	at := time.Unix(2000, 0)
	for range lightKinds {
		m = press(t, m, "e")
		name := lightKinds[m.(model).lt.kind].name
		var out string
		m, out = clockFor(t, m, at)
		at = at.Add(time.Hour)
		require.Contains(t, out, "only its colour moves.", name)
	}
}

func TestTheDiffMeltsTheOldLineAndUndoBringsItBack(t *testing.T) {
	m, _ := onTab(t, 1)
	m = press(t, m, "f")
	m, out := clockFor(t, m, time.Unix(2000, 0))
	require.NotContains(t, out, "return a - b", "the removed line melted away")
	require.Contains(t, out, "return a + b")

	m = press(t, m, "u")
	_, out = clockFor(t, m, time.Unix(3000, 0))
	require.Contains(t, out, "return a - b", "undo put it back")
}
