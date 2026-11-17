package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A crossfade blends colours: between the two cards there are frames that are
// neither, which a card without colours of its own would cut through.
func TestCrossfadeShowsFramesBetweenTheTwoCards(t *testing.T) {
	th := newModel().sh.Theme()
	s := newSwaps()
	for i, k := range swapKinds {
		if slug(k.name) == "crossfade" {
			s.kind = i
		}
	}
	s = s.again()
	at := time.Unix(5000, 0)
	s = s.advance(at)
	seen := map[string]bool{s.stage(th): true}
	for ms := 50; ms <= 1000; ms += 50 {
		s = s.advance(at.Add(time.Duration(ms) * time.Millisecond))
		seen[s.stage(th)] = true
	}
	require.Greater(t, len(seen), 2)
}

// Every transition carries one card into the other, and back again.
func TestEveryTransitionLandsOnTheOtherCard(t *testing.T) {
	m, out := onTab(t, 2)
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
