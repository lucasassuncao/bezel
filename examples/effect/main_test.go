package main

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
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

// turnTab presses tab and renders the new tab.
func turnTab(t *testing.T, m tea.Model) (tea.Model, string) {
	t.Helper()
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	return frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
}

// press sends key k and runs the message its action sends, if it sends one.
func press(t *testing.T, m tea.Model, k string) tea.Model {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key(k))
	if msg := run(cmd); msg != nil {
		m, _ = m.Update(msg)
	}
	return m
}

// run is the message a plain action command sends; batches, which carry the
// clock's ticks, are left alone.
func run(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if _, batch := msg.(tea.BatchMsg); batch {
		return nil
	}
	return msg
}

// clockFor delivers the shared clock's frames every 16ms from at until it
// stops, rendering each, and the last view.
func clockFor(t *testing.T, m tea.Model, at time.Time) (tea.Model, string) {
	t.Helper()
	var out string
	for i := 0; m.(model).clock.Running(); i++ {
		require.Less(t, i, 1000, "the clock never stopped")
		m, out = frame(t, m, m.(model).clock.FrameAt(at))
		at = at.Add(animation.Frame)
	}
	return m, out
}

// onTab is the gallery, sized and started, on tab i.
func onTab(t *testing.T, i int) (tea.Model, string) {
	t.Helper()
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(startMsg{})
	for range i {
		m, out = turnTab(t, m)
	}
	return m, out
}

func TestTheGalleryOpensOnTheDrawTab(t *testing.T) {
	_, out := onTab(t, 0)
	require.Contains(t, out, "every effect")
	for _, tab := range []string{"draw", "light", "swap", "retro"} {
		require.Contains(t, out, tab)
	}
}

// With motion off every effect is drawn at its end and the clock never runs.
func TestAStillGalleryLandsAtOnce(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(startMsg{})
	m, out := frame(t, m, m.(model).clock.FrameAt(time.Unix(1000, 0)))
	require.Contains(t, out, "drawn again.")
	require.Contains(t, out, "motion off")
	require.False(t, m.(model).clock.Running())
}
