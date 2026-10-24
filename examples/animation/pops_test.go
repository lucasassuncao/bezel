package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestThePopTabShowsEveryRow(t *testing.T) {
	_, out := onTab(t, 6)
	t.Log("\n" + out)
	for _, want := range []string{"burst", "donut", "race", "squash", "25%", "db"} {
		require.Contains(t, out, want)
	}
}

func TestPopKeysPlayOnTheClock(t *testing.T) {
	m, _ := onTab(t, 6)
	for _, k := range []string{"x", "o", "v", "d"} {
		m = press(t, m, k)
	}
	pp := m.(model).pp
	require.True(t, pp.burst.Active() && pp.donut.Active() && pp.race.Active() && pp.ball.Active())
	_, out := clockFor(t, m, time.Unix(11000, 0))
	require.Contains(t, out, "80%", "the donut swept to its next share")
	require.Less(t, strings.Index(out, "api"), strings.Index(out, "db"), "api overtook db")
	require.Contains(t, out, "●▐", "the ball came to rest against the wall")
}
