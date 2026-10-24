package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTheTypeTabShowsEveryRow(t *testing.T) {
	_, out := onTab(t, 3)
	t.Log("\n" + out)
	for _, want := range []string{"stream", "morph", "typist", "flip", "neon", "OPEN 24/7", "burn", "big", "fall", "saving…"} {
		require.Contains(t, out, want)
	}
}

func TestTypeKeysPlayOnTheClock(t *testing.T) {
	m, _ := onTab(t, 3)
	for _, k := range []string{"m", "y", "f", "b", "a"} {
		m = press(t, m, k)
	}
	wd := m.(model).wd
	require.True(t, wd.morph.Active() && wd.typist.Active() && wd.flip.Active() && wd.burn.Active() && wd.fall.Active())
	m, out := clockFor(t, m, time.Unix(7000, 0))
	require.Contains(t, out, "saved", "the status morphed")
	require.Contains(t, out, sentence, "the sentence typed right in the end")
	require.Contains(t, out, "12:46", "the clock turned a minute")
	require.Contains(t, out, dropped, "the letters landed")
	require.NotContains(t, out, secret, "the message burned")
	require.False(t, m.(model).wd.moving())
}

func TestTheStreamCatchesUp(t *testing.T) {
	m, _ := onTab(t, 3)
	wd := m.(model).wd
	wd, _ = wd.more()
	start := time.Unix(10, 0)
	wd.stream, _ = wd.stream.Advance(start)
	wd.stream, _ = wd.stream.Advance(start.Add(time.Hour))
	require.Equal(t, strings.Repeat(answer, 2), wd.stream.View())
}
