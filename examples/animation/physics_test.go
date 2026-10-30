package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPhysicsSettleOnTheClock(t *testing.T) {
	m, out := onTab(t, 3)
	t.Log("\n" + out)
	for _, want := range []string{"drop", "pendulum", "orbit", "flock", "rubber", "◉"} {
		require.Contains(t, out, want)
	}
	for _, k := range []string{"d", "f", "b"} {
		m = press(t, m, k)
	}
	require.Greater(t, m.(model).ph.rubber.Value(), float64(limit), "pulled past its top")
	m = press(t, m, "b")
	m, out = clockFor(t, m, time.Unix(6000, 0))
	require.Contains(t, out, "●▐", "the ball lies against the floor")
	require.InDelta(t, limit, m.(model).ph.rubber.Value(), 0.01, "let go, it sprang back to its top")
	require.True(t, strings.Contains(out, "●"))
}
