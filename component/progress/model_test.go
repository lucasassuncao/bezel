package progress

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestAModelSpringsToItsPercent(t *testing.T) {
	m, cmd := New(15).SetPercent(0.5)
	require.NotNil(t, cmd, "a bar that has somewhere to go moves")
	require.Equal(t, 0.5, m.Percent())
	require.Zero(t, m.Shown(), "it starts where it was")

	at := time.Now()
	for i := 0; i < 200 && m.Active(); i++ {
		m, _ = m.Update(m.spring.FrameAt(at))
		at = at.Add(16 * time.Millisecond)
	}
	require.False(t, m.Active())
	require.InDelta(t, 0.5, m.Shown(), 1e-9)
	require.Equal(t, "█████░░░░░  50%", ansi.Strip(m.View(th())))
}

func TestAModelKeepsWithinItsEnds(t *testing.T) {
	m, _ := New(10).SetPercent(1.5)
	require.Equal(t, 1.0, m.Percent())
	m, _ = m.DecrPercent(2)
	require.Zero(t, m.Percent())
	m, _ = m.IncrPercent(0.25)
	require.Equal(t, 0.25, m.Percent())
}

func TestCharsPercentageAndFormat(t *testing.T) {
	m := New(6).WithChars('=', '.').WithPercentage(false)
	require.Equal(t, "===...", ansi.Strip(m.ViewAs(0.5, th())))
	m = New(10).WithFormat(" %.0f")
	require.Equal(t, "████░░░ 50", ansi.Strip(m.ViewAs(0.5, th())), "the bar takes what the percentage leaves")
}

func TestAGradientColoursTheFillEndToEnd(t *testing.T) {
	red, blue := lipgloss.Color("#ff0000"), lipgloss.Color("#0000ff")
	m := New(4).WithPercentage(false).WithGradient(red, blue)
	got := m.ViewAs(1, th())
	require.Contains(t, got, lipgloss.NewStyle().Foreground(red).Render("█"))
	require.Contains(t, got, lipgloss.NewStyle().Foreground(blue).Render("█"))
	require.Contains(t, got, lipgloss.NewStyle().Foreground(animation.Blend(red, blue, 1.0/3)).Render("█"))
}

func TestAStillModelJumps(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	m, cmd := New(10).SetPercent(1)
	require.Nil(t, cmd)
	require.Equal(t, 1.0, m.Shown())
}
