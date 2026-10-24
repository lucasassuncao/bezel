package cursor

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

// th blinks every nanosecond, so a test that runs a tick never waits on one.
var th = func() theme.Resolved {
	t := theme.Resolve(theme.Mint, true)
	t.CaretBlink = time.Nanosecond
	return t
}()

func TestABlinkingCursorFlipsOnItsOwnTicks(t *testing.T) {
	m, cmd := New(th).Focus()
	require.True(t, m.Visible())
	m, cmd = m.Update(cmd())
	require.False(t, m.Visible(), "the first tick darkens it")
	m, _ = m.Update(cmd())
	require.True(t, m.Visible())
}

func TestACursorTakesNoOtherCursorsTicks(t *testing.T) {
	_, other := New(th).Focus()
	m, _ := New(th).Focus()
	m, cmd := m.Update(other())
	require.True(t, m.Visible())
	require.Nil(t, cmd)
}

func TestARefocusLeavesTheOldTickSpent(t *testing.T) {
	m, old := New(th).Focus()
	m, _ = m.Blur().Focus()
	m, cmd := m.Update(old())
	require.True(t, m.Visible())
	require.Nil(t, cmd, "a second loop would blink twice as fast")
}

func TestSteadyAndHiddenNeverTick(t *testing.T) {
	steady, cmd := New(th).WithMode(Steady).Focus()
	require.Nil(t, cmd)
	require.True(t, steady.Visible())
	hidden, cmd := New(th).WithMode(Hidden).Focus()
	require.Nil(t, cmd)
	require.False(t, hidden.Visible())
}

func TestSetModeToBlinkingStartsTheBlink(t *testing.T) {
	m, _ := New(th).WithMode(Steady).Focus()
	m, cmd := m.SetMode(Blinking)
	require.NotNil(t, cmd)
	m, _ = m.Update(cmd())
	require.False(t, m.Visible())
}

func TestResetLightsADarkCursor(t *testing.T) {
	m, cmd := New(th).Focus()
	m, _ = m.Update(cmd())
	m, cmd = m.Reset()
	require.True(t, m.Visible())
	require.NotNil(t, cmd)
}

func TestTheCharIsPaintedOnlyWhileVisible(t *testing.T) {
	m := New(th).SetChar("x")
	require.Equal(t, "x", m.View())
	m, _ = m.Focus()
	require.Equal(t, th.Caret.Render("x"), m.View())
}

func TestWithStylesSurvivesSetTheme(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m, _ := New(th).WithStyles(func(s Styles) Styles { s.Cursor = red; return s }).SetChar("x").Focus()
	m = m.SetTheme(theme.Resolve(theme.Plain, true))
	require.Equal(t, red.Render("x"), m.View())
}

func TestModeNames(t *testing.T) {
	require.Equal(t, []string{"blinking", "steady", "hidden"},
		[]string{Blinking.String(), Steady.String(), Hidden.String()})
}

func TestAZeroCursorDrawsWithoutPanicking(t *testing.T) {
	require.NotPanics(t, func() { _ = Model{}.View() })
}
