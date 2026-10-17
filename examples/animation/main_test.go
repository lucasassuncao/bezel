package main

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

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

// toggle presses h and returns the model with its first frame not yet run.
func toggle(t *testing.T, m tea.Model) tea.Model {
	t.Helper()
	m, cmd := m.Update(bezeltest.Key("h"))
	m, _ = m.Update(cmd())
	return m
}

// tick delivers the tween's own frame for at, as its clock would.
func tick(t *testing.T, m tea.Model, at time.Time) (tea.Model, string) {
	t.Helper()
	return frame(t, m, m.(model).tween.FrameAt(at))
}

func TestAnimationSlidesThroughMiddleHeights(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.NotContains(t, out, "hints")

	m = toggle(t, m)
	start := time.Now()
	m, _ = tick(t, m, start) // the first frame starts the clock
	m, out = tick(t, m, start.Add(slide/2))
	t.Log("\n" + out)
	mid := m.(model).height
	require.Greater(t, mid, 0)
	require.Less(t, mid, hintsHeight, "half way, the pane is part open")

	m, out = tick(t, m, start.Add(2*slide))
	t.Log("\n" + out)
	require.Equal(t, hintsHeight, m.(model).height)
	require.Contains(t, out, "toggle this pane")
}

func TestAnimationClosesToNothing(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	for range 2 {
		m = toggle(t, m)
		start := time.Now()
		m, _ = tick(t, m, start)
		m, _ = tick(t, m, start.Add(2*slide))
	}
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Equal(t, 0, m.(model).height)
	require.NotContains(t, out, "hints")
}

// A toggle mid-slide replaces the tween: the old one's frame, still on its way,
// must not start a second clock beside the new one's.
func TestAnimationDropsTheReplacedTweensFrame(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = toggle(t, m)
	start := time.Now()
	m, _ = tick(t, m, start)
	m, _ = tick(t, m, start.Add(slide/2))
	stale := m.(model).tween.FrameAt(start.Add(slide * 3 / 4))

	m = toggle(t, m)
	require.True(t, m.(model).tween.Active(), "closing from half way is a slide of its own")
	_, cmd := m.Update(stale)
	require.Nil(t, cmd)
}
