package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTheTextTabShowsEveryRow(t *testing.T) {
	_, out := onTab(t, 4)
	t.Log("\n" + out)
	for _, want := range []string{"swell", "rainbow", "decrypt", "caret", "Saving", "ticker", searched, chore, "copy link"} {
		require.Contains(t, out, want)
	}
}

func TestTextKeysPlayOnTheClock(t *testing.T) {
	m, _ := onTab(t, 4)
	for _, k := range []string{"d", "t", "l", "k", "c"} {
		m = press(t, m, k)
	}
	tx := m.(model).tx
	require.True(t, tx.decrypt.Active() && tx.ticker.Active() && tx.mark.Active() && tx.strike.Active() && tx.rise.Active())

	m, out := clockFor(t, m, time.Unix(1000, 0))
	require.Contains(t, out, boarding)
	require.Contains(t, out, "1302", "the ticker moved on to the next price")
	require.False(t, m.(model).tx.moving())
}
