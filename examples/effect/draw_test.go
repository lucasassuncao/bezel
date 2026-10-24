package main

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

// onDraw is the gallery, sized, started and settled, on its first tab.
func onDraw(t *testing.T) (tea.Model, string) {
	t.Helper()
	m, _ := onTab(t, 0)
	return clockFor(t, m, time.Unix(1000, 0))
}

func TestTheDrawTabShowsTheCard(t *testing.T) {
	_, out := onDraw(t)
	t.Log("\n" + out)
	require.Contains(t, out, "fade in")
	require.Contains(t, out, "bezel ✦ effects")
	require.Contains(t, out, "drawn again.")
}

// Every effect, the ones that take the card away included, leaves it whole.
func TestEveryEffectLeavesTheCardWhole(t *testing.T) {
	m, _ := onDraw(t)
	at := time.Unix(2000, 0)
	for range effectKinds {
		m = press(t, m, "e")
		name := effectKinds[m.(model).dr.kind].name
		var out string
		m, out = clockFor(t, m, at)
		at = at.Add(time.Hour)
		require.Contains(t, out, "drawn again.", name)
		require.Contains(t, out, name)
	}
}

func TestEveryPatternPlaysTheEffect(t *testing.T) {
	m, _ := onDraw(t)
	at := time.Unix(2000, 0)
	for range patterns {
		m = press(t, m, "p")
		mid := m.(model)
		require.True(t, mid.dr.moving(), "a new order plays again")
		var out string
		m, out = clockFor(t, m, at)
		at = at.Add(time.Hour)
		require.Contains(t, out, "drawn again.", patterns[m.(model).dr.order].name)
	}
}

// Half way through a dissolve, part of the card is gone and part is left.
func TestADissolveHalfWayLeavesPartOfTheCard(t *testing.T) {
	m, _ := onDraw(t)
	for m.(model).dr.kind != 2 {
		m = press(t, m, "e")
	}
	at := time.Unix(2000, 0)
	m, _ = frame(t, m, m.(model).clock.FrameAt(at))
	_, out := frame(t, m, m.(model).clock.FrameAt(at.Add(play/2)))
	require.NotContains(t, out, "drawn again.")
	require.True(t, strings.Contains(out, "─") || strings.Contains(out, "│"), "some of the border is left")
}
