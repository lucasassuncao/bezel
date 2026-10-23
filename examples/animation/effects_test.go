package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

// onEffects is the gallery, sized and started, on its second tab.
func onEffects(t *testing.T) (tea.Model, string) {
	t.Helper()
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(startMsg{})
	return frame(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
}

func TestTheEffectsTabShowsEveryRow(t *testing.T) {
	_, out := onEffects(t)
	t.Log("\n" + out)
	for _, want := range []string{loading, "scramble", "odometer", title, "● eased", "○ sprung"} {
		require.Contains(t, out, want)
	}
}

func TestScrambleOdometerAndPointsLandOnTheClock(t *testing.T) {
	m, _ := onEffects(t)
	for _, k := range []string{"x", "o", "m"} {
		m = press(t, m, k)
	}
	fx := m.(model).fx
	require.True(t, fx.scramble.Active() && fx.odometer.Active() && fx.move.Active() && fx.follow.Active())

	m, out := clockFor(t, m, time.Unix(1000, 0))
	require.Contains(t, out, serial)
	require.Contains(t, out, "1284")
	fx = m.(model).fx
	require.Equal(t, corners[0], fx.move.Cur)
	require.Equal(t, corners[0], fx.follow.Pos())

	m = press(t, m, "o")
	_, out = clockFor(t, m, time.Unix(2000, 0))
	require.Contains(t, out, "98765", "the odometer rolls on to the next reading")
}

// The shimmer and the gradient loop on their own ticks, the clock stopped.
func TestTheLoopsKeepTheirOwnTicks(t *testing.T) {
	m, _ := onEffects(t)
	mm := m.(model)
	mm.clock = mm.clock.Stop()
	m = mm
	at := time.Unix(1000, 0)
	for i := range 10 {
		fx := m.(model).fx
		var cmd tea.Cmd
		m, cmd = m.Update(fx.shimmer.FrameAt(at.Add(time.Duration(i) * animation.Frame)))
		require.NotNil(t, cmd, "the shimmer asks for its next sweep frame")
		m, _ = m.Update(m.(model).fx.gradient.FrameAt(at.Add(time.Duration(i) * animation.Frame)))
	}
	require.True(t, m.(model).fx.shimmer.Active())
	require.True(t, m.(model).fx.gradient.Active())
}
