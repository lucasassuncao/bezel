package list

import (
	"image/color"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/theme"
)

const ms = time.Millisecond

func moving(t *testing.T) {
	t.Helper()
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("BEZEL_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
}

func plainRows(labels ...string) []Row[int] {
	rows := make([]Row[int], len(labels))
	for i, l := range labels {
		rows[i] = Row[int]{Label: l, Value: i}
	}
	return rows
}

// playTo is m's motion played from its first frame to d after it.
func playTo(m Model[int], d time.Duration) Model[int] {
	start := time.Unix(10, 0)
	m, _ = m.Animate(m.FrameAt(start))
	m, _ = m.Animate(m.FrameAt(start.Add(d)))
	return m
}

// marked is the line of out the cursor is drawn on.
func marked(out string) string {
	for _, l := range strings.Split(ansi.Strip(out), "\n") {
		if strings.HasPrefix(l, "▶") {
			return l
		}
	}
	return ""
}

func TestTheCursorGlidesThroughTheRowsBetween(t *testing.T) {
	moving(t)
	th := theme.Resolve(theme.Plain, true)
	m := New(plainRows("a", "b", "c", "d", "e"), 5).WithMotion(100 * ms)
	for range 4 {
		m, _ = m.Update(bezeltest.Key("down"))
	}
	require.NotNil(t, m.Tick(), "a move with motion asks for frames")
	assert.Contains(t, marked(m.View(th)), "a", "it has not left yet")
	mid := playTo(m, 50*ms)
	assert.Contains(t, marked(mid.View(th)), "c", "halfway down")
	end := playTo(m, time.Second)
	assert.False(t, end.Active())
	assert.Contains(t, marked(end.View(th)), "e")
	assert.Equal(t, "e", end.Selected().Label, "the cursor itself is where it went at once")
}

func TestAListWithoutMotionJumps(t *testing.T) {
	moving(t)
	m := New(plainRows("a", "b"), 2)
	m, _ = m.Update(bezeltest.Key("down"))
	assert.Nil(t, m.Tick())
	assert.False(t, m.Active())
}

func TestANewRowLightsUpAndALeavingRowFades(t *testing.T) {
	moving(t)
	th := theme.Resolve(theme.Plain, true)
	m := New(plainRows("keep", "old"), 4).WithMotion(100 * ms)
	m = m.SetRows(plainRows("keep", "new"))
	require.NotNil(t, m.Tick())
	out := playTo(m, 20*ms).View(th)
	lines := strings.Split(ansi.Strip(out), "\n")
	assert.Equal(t, []string{"▶ keep", "  old", "  new"}, lines, "the old row is still there, fading")
	assert.Contains(t, out, ";9m", "and struck through")
	assert.NotEqual(t, New(plainRows("keep", "new"), 4).View(th), out, "the new row is lit")
	settled := playTo(m, time.Second).View(th)
	assert.Equal(t, New(plainRows("keep", "new"), 4).View(th), settled)
}

func TestAGhostNeverMakesTheListTaller(t *testing.T) {
	moving(t)
	th := theme.Resolve(theme.Plain, true)
	m := New(plainRows("a", "b", "c"), 3).WithMotion(100 * ms)
	m = m.SetRows(plainRows("a", "c", "d"))
	out := playTo(m, 20*ms).View(th)
	assert.Len(t, strings.Split(out, "\n"), 3)
}

func TestTheListTakesOnlyItsOwnFrames(t *testing.T) {
	moving(t)
	m := New(plainRows("a", "b"), 2).WithMotion(100 * ms)
	m, _ = m.Update(bezeltest.Key("down"))
	other := New(plainRows("x", "y"), 2).WithMotion(100 * ms)
	other, _ = other.Update(bezeltest.Key("down"))
	_, cmd := m.Animate(other.FrameAt(time.Unix(1, 0)))
	assert.Nil(t, cmd)
	_, cmd = m.Animate(tea.KeyPressMsg{})
	assert.Nil(t, cmd)
	_, cmd = m.Animate(m.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, cmd)
}

func TestStillListsDoNotMove(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	m := New(plainRows("a", "b"), 2).WithMotion(100 * ms)
	m, _ = m.Update(bezeltest.Key("down"))
	assert.False(t, m.Active())
	assert.Nil(t, m.Tick())
	assert.Contains(t, marked(m.View(theme.Resolve(theme.Plain, true))), "b")
}

func TestMixBlendsAndGivesWayToAMissingColour(t *testing.T) {
	black, white := color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}
	assert.Equal(t, color.RGBA{128, 128, 128, 255}, mix(black, white, 0.5))
	assert.Equal(t, black, mix(black, nil, 0.2))
	assert.Equal(t, white, mix(nil, white, 0.2))
}
