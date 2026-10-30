package timer

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Mint, true)

// fast steps a nanosecond a tick, so a test that runs a tick never waits.
func fast(d time.Duration) Model { return New(d).WithEvery(time.Nanosecond) }

func TestANewTimerWaitsForStart(t *testing.T) {
	m := New(time.Minute)
	require.False(t, m.Running())
	require.Equal(t, time.Minute, m.Left())
}

func TestEachTickTakesOneStep(t *testing.T) {
	m, cmd := fast(3 * time.Nanosecond).Start()
	m, _ = m.Update(cmd())
	require.Equal(t, 2*time.Nanosecond, m.Left())
	require.True(t, m.Running())
}

func TestTheLastTickTimesOut(t *testing.T) {
	m, cmd := fast(time.Nanosecond).Start()
	m, cmd = withCmd(m.Update(cmd()))
	require.True(t, m.TimedOut())
	require.False(t, m.Running())
	require.Equal(t, TimeoutMsg{ID: m.ID()}, cmd())
}

func TestTimeLeftNeverGoesBelowZero(t *testing.T) {
	m, cmd := New(time.Nanosecond).WithEvery(2 * time.Nanosecond).Start()
	m, _ = m.Update(cmd())
	require.Equal(t, time.Duration(0), m.Left())
}

func TestARestartLeavesTheOldTickSpent(t *testing.T) {
	m, old := fast(5 * time.Nanosecond).Start()
	m, _ = m.Stop().Start()
	m, cmd := withCmd(m.Update(old()))
	require.Equal(t, 5*time.Nanosecond, m.Left())
	require.Nil(t, cmd)
}

func TestATimerTakesNoOtherTimersTicks(t *testing.T) {
	_, other := fast(5 * time.Nanosecond).Start()
	m, _ := fast(5 * time.Nanosecond).Start()
	m, cmd := withCmd(m.Update(other()))
	require.Equal(t, 5*time.Nanosecond, m.Left())
	require.Nil(t, cmd)
}

func TestATimedOutTimerDoesNotStart(t *testing.T) {
	m, cmd := New(0).Start()
	require.Nil(t, cmd)
	require.False(t, m.Running())
	require.Equal(t, time.Duration(0), New(-time.Second).Left())
}

func TestToggleStartsAndStops(t *testing.T) {
	m, cmd := fast(5 * time.Nanosecond).Toggle()
	require.True(t, m.Running())
	require.NotNil(t, cmd)
	m, _ = m.Toggle()
	require.False(t, m.Running())
}

func TestResetStopsWithTheNewTime(t *testing.T) {
	m, _ := fast(5 * time.Nanosecond).Start()
	m = m.Reset(time.Minute)
	require.False(t, m.Running())
	require.Equal(t, time.Minute, m.Left())
}

func TestAnEveryOfZeroIsASecond(t *testing.T) {
	require.Equal(t, time.Second, New(time.Minute).WithEvery(0).Every())
}

func TestTheViewWritesTheTimeLeft(t *testing.T) {
	require.Equal(t, "1m30s", ansi.Strip(New(90*time.Second).View(th)))
	f := func(d time.Duration) string { return fmt.Sprintf("%.0fs", d.Seconds()) }
	require.Equal(t, "90s", ansi.Strip(New(90*time.Second).WithFormat(f).View(th)))
}

func TestTheViewIsStyledByState(t *testing.T) {
	require.Equal(t, th.Dim.Render("1m0s"), New(time.Minute).View(th))
	m, _ := New(time.Minute).Start()
	require.Equal(t, lipgloss.NewStyle().Render("1m0s"), m.View(th))
	require.Equal(t, th.Danger.Render("0s"), New(0).View(th))
}

func TestWithStylesHoldsUnderAnotherTheme(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New(time.Minute).WithStyles(func(s Styles) Styles { s.Stopped = red; return s })
	require.Equal(t, red.Render("1m0s"), m.View(theme.Resolve(theme.Plain, true)))
}

func TestAZeroTimerStepsASecond(t *testing.T) {
	require.Equal(t, time.Second, Model{}.Reset(time.Minute).Every())
}

func TestTwoZeroTimersTakeOnlyTheirOwnTicks(t *testing.T) {
	_, other := Model{}.Reset(5 * time.Nanosecond).WithEvery(time.Nanosecond).Start()
	m, _ := Model{}.Reset(5 * time.Nanosecond).WithEvery(time.Nanosecond).Start()
	m, cmd := withCmd(m.Update(other()))
	require.Equal(t, 5*time.Nanosecond, m.Left())
	require.Nil(t, cmd)
}

func TestOwnTickIsHandled(t *testing.T) {
	m, cmd := fast(3 * time.Nanosecond).Start()
	_, r := m.Update(cmd())
	require.True(t, r.Handled)
	require.NotNil(t, r.Cmd)
}

func TestForeignFrameIsNotHandled(t *testing.T) {
	_, other := fast(5 * time.Nanosecond).Start()
	m, _ := fast(5 * time.Nanosecond).Start()
	_, r := m.Update(other())
	require.False(t, r.Handled)
	_, r = m.Update(bezeltest.Key("a"))
	require.False(t, r.Handled)
}

// withCmd keeps the command of an Update, for the tests that run it.
func withCmd[T any](m T, r bezel.Result) (T, tea.Cmd) { return m, r.Cmd }
