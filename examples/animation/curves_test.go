package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

// onCurves is the gallery, sized and started, on its second tab.
func onCurves(t *testing.T) (tea.Model, string) {
	t.Helper()
	return onTab(t, 1)
}

func TestTheCurvesTabRacesEveryCurve(t *testing.T) {
	m, out := onCurves(t)
	t.Log("\n" + out)
	for _, f := range family {
		require.Contains(t, out, f.name)
	}
	require.Contains(t, out, "out", "the way they are played is named")

	m = press(t, m, "r")
	m, _ = frame(t, m, m.(model).clock.FrameAt(time.Unix(1000, 0)))
	m, _ = frame(t, m, m.(model).clock.FrameAt(time.Unix(1000, 0).Add(300*time.Millisecond)))
	dots := m.(model).cv.dots
	require.Greater(t, dots[6].Value(), dots[0].Value(), "expo out has run further than linear by now")

	m, _ = clockFor(t, m, time.Unix(1001, 0))
	for _, d := range m.(model).cv.dots {
		require.Equal(t, 1.0, d.Value(), "every curve ends on its end mark")
	}
}

func TestITurnsTheCurvesRound(t *testing.T) {
	m, _ := onCurves(t)
	for _, want := range []string{"in", "in-out", "out"} {
		m = press(t, m, "i")
		require.Equal(t, want, turns[m.(model).cv.turn].name)
		require.True(t, m.(model).cv.moving(), "turning races them again")
	}
}

// A curve that leaves [0,1] shows past its marks, kept on the track.
func TestADotPastItsMarksStaysOnTheTrack(t *testing.T) {
	th := newModel().th
	for _, v := range []float64{-1, -0.1, 0, 0.5, 1, 1.2, 3} {
		out := dotOn(th, v)
		require.Contains(t, out, "●")
	}
}
