package stopwatch

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Mint, true)

// fast steps a nanosecond a tick, so a test that runs a tick never waits.
func fast() Model { return New().WithEvery(time.Nanosecond) }

func TestANewStopwatchWaitsAtZero(t *testing.T) {
	m := New()
	require.False(t, m.Running())
	require.Equal(t, time.Duration(0), m.Elapsed())
}

func TestEachTickAddsOneStep(t *testing.T) {
	m, cmd := fast().Start()
	m, cmd = withCmd(m.Update(cmd()))
	m, _ = m.Update(cmd())
	require.Equal(t, 2*time.Nanosecond, m.Elapsed())
}

func TestARestartLeavesTheOldTickSpent(t *testing.T) {
	m, old := fast().Start()
	m, _ = m.Stop().Start()
	m, cmd := withCmd(m.Update(old()))
	require.Equal(t, time.Duration(0), m.Elapsed())
	require.Nil(t, cmd)
}

func TestAStopwatchTakesNoOtherStopwatchesTicks(t *testing.T) {
	_, other := fast().Start()
	m, _ := fast().Start()
	m, cmd := withCmd(m.Update(other()))
	require.Equal(t, time.Duration(0), m.Elapsed())
	require.Nil(t, cmd)
}

func TestResetKeepsItRunning(t *testing.T) {
	m, cmd := fast().Start()
	m, cmd = withCmd(m.Update(cmd()))
	m = m.Reset()
	require.Equal(t, time.Duration(0), m.Elapsed())
	require.True(t, m.Running())
	m, _ = m.Update(cmd())
	require.Equal(t, time.Nanosecond, m.Elapsed(), "the loop it had goes on")
}

func TestToggleStartsAndStops(t *testing.T) {
	m, cmd := fast().Toggle()
	require.True(t, m.Running())
	require.NotNil(t, cmd)
	m, _ = m.Toggle()
	require.False(t, m.Running())
}

func TestAnEveryOfZeroIsASecond(t *testing.T) {
	require.Equal(t, time.Second, New().WithEvery(-1).Every())
}

func TestTheViewIsStyledByState(t *testing.T) {
	require.Equal(t, th.Dim.Render("0s"), New().View(th))
	m, _ := New().Start()
	require.Equal(t, lipgloss.NewStyle().Render("0s"), m.View(th))
}

func TestWithStylesHoldsUnderAnotherTheme(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New().WithStyles(func(s Styles) Styles { s.Stopped = red; return s })
	require.Equal(t, red.Render("0s"), m.View(theme.Resolve(theme.Plain, true)))
}

func TestAZeroStopwatchStepsASecond(t *testing.T) {
	require.Equal(t, time.Second, Model{}.Every())
}

func TestTwoZeroStopwatchesTakeOnlyTheirOwnTicks(t *testing.T) {
	_, other := Model{}.WithEvery(time.Nanosecond).Start()
	m, _ := Model{}.WithEvery(time.Nanosecond).Start()
	m, cmd := withCmd(m.Update(other()))
	require.Equal(t, time.Duration(0), m.Elapsed())
	require.Nil(t, cmd)
}

func TestOwnTickIsHandled(t *testing.T) {
	m, cmd := fast().Start()
	_, r := m.Update(cmd())
	require.True(t, r.Handled)
	require.NotNil(t, r.Cmd)
}

func TestForeignFrameIsNotHandled(t *testing.T) {
	_, other := fast().Start()
	m, _ := fast().Start()
	_, r := m.Update(other())
	require.False(t, r.Handled)
	_, r = m.Update(bezeltest.Key("a"))
	require.False(t, r.Handled)
}

// withCmd keeps the command of an Update, for the tests that run it.
func withCmd[T any](m T, r bezel.Result) (T, tea.Cmd) { return m, r.Cmd }
