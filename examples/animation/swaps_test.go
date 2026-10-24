package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Every transition carries one card into the other, and back again.
func TestEveryTransitionLandsOnTheOtherCard(t *testing.T) {
	m, out := onTab(t, 7)
	require.Contains(t, out, "Inbox")
	at := time.Unix(5000, 0)
	for range swapKinds {
		m = press(t, m, "n")
		var out string
		m, out = clockFor(t, m, at)
		at = at.Add(time.Hour)
		want := "Inbox"
		if m.(model).sw.flipped {
			want = "Settings"
		}
		require.Contains(t, out, want, swapKinds[m.(model).sw.kind].name)
	}
	m = press(t, m, "r")
	require.True(t, m.(model).sw.moving())
}
