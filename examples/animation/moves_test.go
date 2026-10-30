package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestTheMoveTabShowsThePointsAndThePath(t *testing.T) {
	_, out := onTab(t, 2)
	t.Log("\n" + out)
	for _, want := range []string{"● eased", "○ sprung", "path", "bezier"} {
		require.Contains(t, out, want)
	}
}

func TestThePointsAndTheMarkerLandOnTheClock(t *testing.T) {
	m, _ := onTab(t, 2)
	m = press(t, m, "m")
	m = press(t, m, "p")
	mv := m.(model).mv
	require.True(t, mv.move.Active() && mv.follow.Active() && mv.path.Active())
	m, _ = clockFor(t, m, time.Unix(3000, 0))
	mv = m.(model).mv
	require.Equal(t, corners[0], mv.move.Value())
	require.Equal(t, corners[0], mv.follow.Value())
	require.Equal(t, animation.Point{X: routeW - 1, Y: 0}, mv.path.Value(), "the marker reached the end of its route")
}
