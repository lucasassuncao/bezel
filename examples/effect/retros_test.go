package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTheRetroTabShowsTheCardAndTheLoadingPanel(t *testing.T) {
	m, out := onTab(t, 3)
	t.Log("\n" + out)
	for _, want := range []string{"retro", "glitch", "bezel ✦ retro", "skeleton", "▒"} {
		require.Contains(t, out, want)
	}
	m = press(t, m, "l")
	_, out = clockFor(t, m, time.Unix(8000, 0))
	require.Contains(t, out, "Weekly report", "done loading, the panel shows")
}

func TestEveryRetroEffectPlaysOutOnTheClock(t *testing.T) {
	m, _ := onTab(t, 3)
	for range retroKinds {
		m = press(t, m, "e")
		var out string
		m, out = clockFor(t, m, time.Unix(9000, 0))
		require.False(t, m.(model).rt.effect.Active(), "%s ended", retroKinds[m.(model).rt.kind].name)
		require.NotEmpty(t, out)
	}
}
