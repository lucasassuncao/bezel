package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTheFlowTabPlaysToItsEnd(t *testing.T) {
	m, out := onTab(t, 5)
	t.Log("\n" + out)
	for _, want := range []string{"timeline", "keyframes", "stagger"} {
		require.Contains(t, out, want)
	}
	m = press(t, m, "r")
	require.True(t, m.(model).fl.moving())
	m, out = clockFor(t, m, time.Unix(10000, 0))
	require.Equal(t, 5, strings.Count(out, strings.Repeat("█", track)), "every staggered bar is full")
	require.InDelta(t, 20, m.(model).fl.keys.Value(), 1e-9, "the dial rests on its last stop")
}
