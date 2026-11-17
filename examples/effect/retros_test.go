package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/examples/internal/look"
	"github.com/lucasassuncao/bezel/theme"
)

func TestTheRetroTabShowsTheCardAndTheLoadingPanel(t *testing.T) {
	m, out := onTab(t, 3)
	t.Log("\n" + out)
	for _, want := range []string{"retro", "glitch", "bezel ✦ retro", "skeleton", "▒"} {
		require.Contains(t, out, want)
	}
	m = press(t, m, "l")
	_, out = clockFor(t, m, time.Unix(8000, 0))
	require.Contains(t, out, "Weekly report", "done loading, the panel shows")
}

func TestEveryRetroEffectPlaysOutOnTheClock(t *testing.T) {
	m, _ := onTab(t, 3)
	for range retroKinds {
		m = press(t, m, "e")
		var out string
		m, out = clockFor(t, m, time.Unix(9000, 0))
		require.False(t, m.(model).rt.effect.Active(), "%s ended", retroKinds[m.(model).rt.kind].name)
		require.NotEmpty(t, out)
	}
}

// An effect lit with no colour glows black, and on a black terminal that is nothing.
func TestThePowerEffectsGlowInTheTextColour(t *testing.T) {
	th := theme.Resolve(look.Plain(), true)
	at := time.Unix(5000, 0)
	for name, ms := range map[string]int{"power off": 700, "power on": 200} {
		r := newRetros(th)
		for i, k := range retroKinds {
			if k.name == name {
				r.kind = i
			}
		}
		r = r.again(th).advance(at).advance(at.Add(time.Duration(ms) * time.Millisecond))
		require.Greater(t, look.Brightest(r.stage(th)), 120, name)
	}
}

// The skeleton's shine is the text colour: brighter than the bars it runs over.
func TestTheSkeletonShinesBrighterThanItsBars(t *testing.T) {
	th := theme.Resolve(look.Plain(), true)
	r := newRetros(th)
	at := time.Unix(5000, 0)
	best := 0
	for ms := 0; ms <= 1000; ms += 50 {
		r.shape, _ = r.shape.Advance(at.Add(time.Duration(ms) * time.Millisecond))
		best = max(best, look.Brightest(r.shape.Apply(report(th))))
	}
	require.Greater(t, best, 150)
}
