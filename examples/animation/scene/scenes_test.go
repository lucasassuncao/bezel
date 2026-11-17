package main

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/examples/internal/look"
	"github.com/lucasassuncao/bezel/theme"
)

// Every scene plays on its own frames with its title over it.
func TestEveryScenePlaysUnderItsTitle(t *testing.T) {
	m, out := opened(t)
	require.Contains(t, out, "bezel ✦ rain")
	for range sceneKinds {
		m = press(t, m, "n")
		at := time.Unix(4000, 0)
		sc := m.(model).sc
		m, _ = frame(t, m, sc.scene.FrameAt(at))
		m, out = frame(t, m, m.(model).sc.scene.FrameAt(at.Add(time.Second)))
		require.Contains(t, out, "bezel ✦ "+sceneKinds[m.(model).sc.kind].name)
	}
}

// A scene given no colour draws black, and on a black terminal that is nothing.
func TestTheScenesInTheTextColourCanBeSeen(t *testing.T) {
	th := theme.Resolve(look.Plain(), true)
	blank := strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", 44)+"\n", 10), "\n")
	for _, k := range sceneKinds {
		if !slices.Contains([]string{"snow", "starfield", "static", "sea"}, k.name) {
			continue
		}
		sc := k.make(th)
		at := time.Unix(4000, 0)
		for i := range 40 {
			sc, _ = sc.Update(sc.FrameAt(at.Add(time.Duration(i) * 100 * time.Millisecond)))
		}
		require.Greater(t, look.Brightest(sc.Under(blank)), 120, k.name)
	}
}
