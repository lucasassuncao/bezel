package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTheDataTabDrawsEveryChart(t *testing.T) {
	m, out := onTab(t, 4)
	t.Log("\n" + out)
	for _, want := range []string{"bars", "cpu", "spark", "area", "gauge", "heat", "▲"} {
		require.Contains(t, out, want)
	}
	for _, k := range []string{"g", "p", "n", "w"} {
		m = press(t, m, k)
	}
	require.True(t, m.(model).dt.moving())
	m, _ = clockFor(t, m, time.Unix(7000, 0))
	require.False(t, m.(model).dt.moving())
	require.Len(t, m.(model).dt.spark.Values(), sparkKeep)
}
