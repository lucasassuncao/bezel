package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestTheRollTabShowsEveryRow(t *testing.T) {
	_, out := onTab(t, 1)
	t.Log("\n" + out)
	for _, want := range []string{loading, "scramble", "odometer", title} {
		require.Contains(t, out, want)
	}
}

func TestScrambleAndOdometerLandOnTheClock(t *testing.T) {
	m, _ := onTab(t, 1)
	for _, k := range []string{"x", "o"} {
		m = press(t, m, k)
	}
	ro := m.(model).ro
	require.True(t, ro.scramble.Active() && ro.odometer.Active())

	m, out := clockFor(t, m, time.Unix(1000, 0))
	require.Contains(t, out, serial)
	require.Contains(t, out, "1284")

	m = press(t, m, "o")
	_, out = clockFor(t, m, time.Unix(2000, 0))
	require.Contains(t, out, "98765", "the odometer rolls on to the next reading")
}

// The shimmer and the gradient loop on their own ticks, the clock stopped.
func TestTheLoopsKeepTheirOwnTicks(t *testing.T) {
	m, _ := onTab(t, 1)
	mm := m.(model)
	mm.clock = mm.clock.Stop()
	m = mm
	at := time.Unix(1000, 0)
	for i := range 10 {
		ro := m.(model).ro
		var cmd tea.Cmd
		m, cmd = m.Update(ro.shimmer.FrameAt(at.Add(time.Duration(i) * animation.Frame)))
		require.NotNil(t, cmd, "the shimmer asks for its next sweep frame")
		m, _ = m.Update(m.(model).ro.gradient.FrameAt(at.Add(time.Duration(i) * animation.Frame)))
	}
	require.True(t, m.(model).ro.shimmer.Active())
	require.True(t, m.(model).ro.gradient.Active())
}
