package spinner

import (
	"os"
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

func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("BEZEL_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

// fast is s turning every nanosecond, so a test that runs a tick never waits.
func fast(s Spinner) Spinner {
	s.Every = time.Nanosecond
	return s
}

func TestThePresetsKeepTheirFramesAndPaces(t *testing.T) {
	for _, c := range []struct {
		s     Spinner
		first string
		n     int
		every time.Duration
	}{
		{Line(), "|", 4, time.Second / 10}, {Dot(), "⣾ ", 8, time.Second / 10},
		{MiniDot(), "⠋", 10, time.Second / 12}, {Jump(), "⢄", 7, time.Second / 10},
		{Pulse(), "█", 4, time.Second / 8}, {Points(), "∙∙∙", 4, time.Second / 7},
		{Globe(), "🌍", 3, time.Second / 4}, {Moon(), "🌑", 8, time.Second / 8},
		{Monkey(), "🙈", 3, time.Second / 3}, {Meter(), "▱▱▱", 7, time.Second / 7},
		{Hamburger(), "☱", 4, time.Second / 3}, {Ellipsis(), "", 4, time.Second / 3},
	} {
		require.Equal(t, c.first, c.s.Frames[0])
		require.Len(t, c.s.Frames, c.n)
		require.Equal(t, c.every, c.s.Every)
	}
}

func TestAPresetIsAFreshValue(t *testing.T) {
	d := Dot()
	d.Frames[0] = "x"
	require.Equal(t, "⣾ ", Dot().Frames[0])
}

func TestStartAdvancesOnItsOwnTicks(t *testing.T) {
	m, cmd := New(fast(Line())).Start()
	require.Equal(t, "|", m.Frame())
	m, cmd = withCmd(m.Update(cmd()))
	require.Equal(t, "/", m.Frame())
	require.NotNil(t, cmd)
}

func TestStartWhileRunningStartsNothing(t *testing.T) {
	m, _ := New(fast(Line())).Start()
	_, cmd := m.Start()
	require.Nil(t, cmd)
}

func TestARestartLeavesTheOldTickSpent(t *testing.T) {
	m, old := New(fast(Line())).Start()
	m, cur := m.Stop().Start()
	m, cmd := withCmd(m.Update(old()))
	require.Equal(t, "|", m.Frame())
	require.Nil(t, cmd, "a second loop would spin twice as fast")
	m, cmd = withCmd(m.Update(cur()))
	require.Equal(t, "/", m.Frame())
	require.NotNil(t, cmd)
}

func TestASpinnerTakesNoOtherSpinnersTicks(t *testing.T) {
	_, other := New(fast(Line())).Start()
	m, _ := New(fast(Line())).Start()
	m, cmd := withCmd(m.Update(other()))
	require.Equal(t, "|", m.Frame())
	require.Nil(t, cmd)
}

func TestTheViewIsAsWideAsTheWidestFrame(t *testing.T) {
	require.Equal(t, 3, ansi.StringWidth(New(Ellipsis()).View(th)))
	require.Equal(t, 2, ansi.StringWidth(New(Globe()).View(th)))
}

func TestTheFrameIsPaintedInTheAccent(t *testing.T) {
	require.Equal(t, th.Accent.Render("|"), New(Line()).View(th))
}

func TestWithStylesHoldsUnderAnotherTheme(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New(Line()).WithStyles(func(s Styles) Styles { s.Frame = red; return s })
	require.Equal(t, red.Render("|"), m.View(theme.Resolve(theme.Plain, true)))
}

func TestASpinnerWithoutFramesDrawsNothingAndNeverTicks(t *testing.T) {
	m, cmd := New(Spinner{}).Start()
	require.Nil(t, cmd)
	require.Equal(t, "", m.View(th))
}

func TestAnEveryOfZeroIsATenthOfASecond(t *testing.T) {
	require.Equal(t, time.Second/10, New(Spinner{Frames: []string{"a"}}).Spinner().Every)
}

// With motion off the spinner runs on its first frame and asks for none.
func TestAStillSpinnerHoldsItsFirstFrame(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	m, cmd := New(Line()).Start()
	require.True(t, m.Running())
	require.Nil(t, cmd)
	require.Equal(t, "|", m.Frame())
}

// A host that delivers frames itself drives it through FrameAt.
func TestFrameAtTurnsItAndOwnsTellsItsFrames(t *testing.T) {
	m, _ := New(Line()).Start()
	other, _ := New(Line()).Start()
	f := m.FrameAt(time.Now())
	require.True(t, m.Owns(f))
	require.False(t, other.Owns(f))
	m, cmd := withCmd(m.Update(f))
	require.Equal(t, "/", m.Frame())
	require.NotNil(t, cmd)
}

func TestOwnTickIsHandled(t *testing.T) {
	m, cmd := New(fast(Line())).Start()
	_, r := m.Update(cmd())
	require.True(t, r.Handled)
	require.NotNil(t, r.Cmd)
}

func TestForeignFrameIsNotHandled(t *testing.T) {
	m, _ := New(fast(Line())).Start()
	_, other := New(fast(Line())).Start()
	_, r := m.Update(other())
	require.False(t, r.Handled)
	_, r = m.Update(bezeltest.Key("a"))
	require.False(t, r.Handled)
}

// withCmd keeps the command of an Update, for the tests that run it.
func withCmd[T any](m T, r bezel.Result) (T, tea.Cmd) { return m, r.Cmd }
