package main

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestTheLightTabShowsTheCardTheDiffAndTheGhost(t *testing.T) {
	_, out := onTab(t, 1)
	t.Log("\n" + out)
	for _, want := range []string{"bezel ✦ light", "func sum", "report.pdf"} {
		require.Contains(t, out, want)
	}
}

// Every light ends, and leaves the card as it found it.
func TestEveryLightLeavesTheCardWhole(t *testing.T) {
	m, _ := onTab(t, 1)
	at := time.Unix(2000, 0)
	for range lightKinds {
		m = press(t, m, "e")
		name := lightKinds[m.(model).lt.kind].name
		var out string
		m, out = clockFor(t, m, at)
		at = at.Add(time.Hour)
		require.Contains(t, out, "only its colour moves.", name)
	}
}

// A light only moves cells that have a colour, and washing a colour towards
// itself shows nothing: each light must change the card's text while it plays.
func TestEveryLightChangesTheCardsTextWhileItPlays(t *testing.T) {
	th := newModel().sh.Theme()
	// text is the first letters of the card's second row. A colour set and reset
	// at once paints nothing, so the border's, which the cut carries, goes.
	dead := regexp.MustCompile(`(?:\x1b\[[0-9;]*m)+\x1b\[m`)
	text := func(card string) string {
		for _, line := range strings.Split(card, "\n") {
			if strings.Contains(ansi.Strip(line), "Every cell") {
				return dead.ReplaceAllString(ansi.Cut(line, 2, 12), "\x1b[m")
			}
		}
		return ""
	}
	for i, k := range lightKinds {
		t.Run(k.name, func(t *testing.T) {
			l := newLights(th)
			l.kind = i
			l = l.again(th)
			at := time.Unix(2000, 0)
			l = l.advance(at)
			rest := text(l.stage(th))
			moved := false
			for ms := 50; ms <= 1500 && !moved; ms += 50 {
				l = l.advance(at.Add(time.Duration(ms) * time.Millisecond))
				moved = text(l.stage(th)) != rest
			}
			require.True(t, moved)
		})
	}
}

func TestTheDiffMeltsTheOldLineAndUndoBringsItBack(t *testing.T) {
	m, _ := onTab(t, 1)
	m = press(t, m, "f")
	m, out := clockFor(t, m, time.Unix(2000, 0))
	require.NotContains(t, out, "return a - b", "the removed line melted away")
	require.Contains(t, out, "return a + b")

	m = press(t, m, "u")
	_, out = clockFor(t, m, time.Unix(3000, 0))
	require.Contains(t, out, "return a - b", "undo put it back")
}
