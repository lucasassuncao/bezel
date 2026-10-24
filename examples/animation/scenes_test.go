package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Every scene plays on its own frames with its title over it.
func TestEveryScenePlaysUnderItsTitle(t *testing.T) {
	m, out := onTab(t, 6)
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
