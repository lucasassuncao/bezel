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

func TestAnimationSlidesThroughMiddleHeights(t *testing.T) {
	var m tea.Model = newModel()
	m, out := frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	require.NotContains(t, out, "━ hints ━")

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
	require.Contains(t, out, "slides on a Tween")
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
	require.NotContains(t, out, "━ hints ━")
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

// The bars race on one clock: half way the linear one leads the eased-in one,
// and at the end every bar is full, the counter is there and the clock stops.
func TestTheRaceRunsOnTheSharedClock(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(startMsg{})
	require.True(t, m.(model).clock.Running())

	at := time.Unix(1000, 0)
	m, _ = frame(t, m, m.(model).clock.FrameAt(at))
	m, out := frame(t, m, m.(model).clock.FrameAt(at.Add(race/2)))
	t.Log("\n" + out)
	bars := m.(model).bars
	require.Equal(t, track/2, bars[0].Cur, "linear is half way at half time")
	require.Less(t, bars[1].Cur, bars[0].Cur, "ease in starts slower")
	require.Greater(t, bars[2].Cur, bars[0].Cur, "ease out starts faster")

	m, out = clockFor(t, m, at.Add(race))
	for _, b := range m.(model).bars {
		require.Equal(t, track, b.Cur)
	}
	require.Contains(t, out, "12,345 files")
	require.False(t, m.(model).clock.Running(), "nothing moves, so the clock stops")
}

func TestTheArrowsMoveBothSprings(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(bezeltest.Key("right"))
	m, _ = m.Update(bezeltest.Key("right"))
	require.True(t, m.(model).bouncy.Active())

	m, _ = clockFor(t, m, time.Unix(1000, 0))
	require.Equal(t, 2*nudge, m.(model).steady.Int())
	require.Equal(t, 2*nudge, m.(model).bouncy.Int())
}

func TestFadeTypeAndShakeLandAndStop(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, k := range []string{"f", "t", "s"} {
		m = press(t, m, k)
	}
	mid := m.(model)
	require.True(t, mid.fade.Active() && mid.typed.Active() && mid.shake.Active())
	require.Equal(t, "", mid.typed.View(), "the headline starts empty")

	m, out := clockFor(t, m, time.Unix(1000, 0))
	require.Contains(t, out, headline)
	require.Contains(t, out, saved)
	require.Equal(t, 0, m.(model).shake.Offset())
}

// The marquee keeps its own ticks: it moves with the shared clock stopped.
func TestTheMarqueeStepsOnItsOwnTicks(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(startMsg{})
	mm := m.(model)
	mm.clock = mm.clock.Stop()
	m = mm
	before := m.(model).marquee.View()

	at := time.Unix(1000, 0)
	m, _ = frame(t, m, m.(model).marquee.FrameAt(at))
	m, _ = frame(t, m, m.(model).marquee.FrameAt(at.Add(time.Second)))
	require.NotEqual(t, before, m.(model).marquee.View())
}

// TestMain clears the motion switches, so a CI with TERM=dumb still animates.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

func TestThePulseRunsUntilPressedAgain(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = press(t, m, "p")
	at := time.Unix(1000, 0)
	for range 200 {
		m, _ = frame(t, m, m.(model).clock.FrameAt(at))
		at = at.Add(animation.Frame)
	}
	require.True(t, m.(model).pulse.Active(), "a yoyo forever never lands")
	require.True(t, m.(model).clock.Running())

	m = press(t, m, "p")
	m, _ = clockFor(t, m, at)
	require.False(t, m.(model).clock.Running(), "with the pulse off, the clock stops")
}

// Each block starts a stagger after the one before it.
func TestTheCascadeStartsOneBlockAtATime(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = press(t, m, "r")
	at := time.Unix(1000, 0)
	m, _ = frame(t, m, m.(model).clock.FrameAt(at))
	m, _ = frame(t, m, m.(model).clock.FrameAt(at.Add(stagger+stagger/2)))
	c := m.(model).cascade
	require.Positive(t, c[0].Cur)
	require.Positive(t, c[1].Cur)
	require.Zero(t, c[2].Cur, "the third has not started yet")
	m, _ = clockFor(t, m, at.Add(stagger*2))
	for _, b := range m.(model).cascade {
		require.Equal(t, block, b.Cur)
	}
}

// With motion off every row is drawn at its end and the clock never runs.
func TestAStillGalleryLandsAtOnce(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	var m tea.Model = newModel()
	m, _ = frame(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(startMsg{})
	m = press(t, m, "t")
	m, out := frame(t, m, m.(model).clock.FrameAt(time.Unix(1000, 0)))
	require.Contains(t, out, "12,345 files")
	require.Contains(t, out, headline)
	require.Contains(t, out, "motion off")
	require.False(t, m.(model).marquee.Active())
	require.False(t, m.(model).clock.Running())
}
