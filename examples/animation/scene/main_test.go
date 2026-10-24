package main

import (
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

// TestMain clears the motion switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

// frame sends msg and renders; every row must be exactly 80 columns wide.
func frame(t *testing.T, m tea.Model, msg tea.Msg) (tea.Model, string) {
	t.Helper()
	m, _ = m.Update(msg)
	out := ansi.Strip(m.View().Content)
	for i, l := range strings.Split(out, "\n") {
		require.Equal(t, 80, ansi.StringWidth(l), "row %d", i)
	}
	return m, out
}

// press sends key k and runs the message its action sends, if it sends one.
func press(t *testing.T, m tea.Model, k string) tea.Model {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key(k))
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, batch := msg.(tea.BatchMsg); !batch {
				m, _ = m.Update(msg)
			}
		}
	}
	return m
}

// opened is the gallery, sized.
func opened(t *testing.T) (tea.Model, string) {
	t.Helper()
	return frame(t, newModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestTheGalleryOpensOnRain(t *testing.T) {
	_, out := opened(t)
	require.Contains(t, out, "every scene")
	require.Contains(t, out, "n next")
}
