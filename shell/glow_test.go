package shell

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

// glowing is a two-pane shell, sized, whose focus colour travels over 100ms.
func glowing(t *testing.T, fade time.Duration) Shell {
	t.Helper()
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
	s := New(Config{
		Layout:    layout.Columns(layout.Fill("left"), layout.Fill("right")),
		Theme:     theme.Resolve(theme.Mint(), true),
		Actions:   []Action{ChangePane()},
		FocusFade: fade,
	})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	return s
}

func TestTheFocusColourTravelsBetweenPanes(t *testing.T) {
	s := glowing(t, 100*time.Millisecond)
	s, r := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.NotNil(t, r.Cmd)
	require.Equal(t, "right", s.Focused())
	accent := lipgloss.Color(s.cfg.Theme.Colors.Accent)
	border := lipgloss.Color(s.cfg.Theme.Colors.Border)
	assert.Equal(t, rgba(border), rgba(s.glowColor("right")), "the new pane starts cool")
	assert.Equal(t, rgba(accent), rgba(s.glowColor("left")), "the old one starts lit")

	at := time.Unix(10, 0)
	s, r = s.Update(frames.NewMsg(at, s.glow.run.ID, s.glow.run.Tag))
	require.True(t, r.Handled, "the shell takes its own frames")
	s, _ = s.Update(frames.NewMsg(at.Add(time.Second), s.glow.run.ID, s.glow.run.Tag))
	assert.Nil(t, s.glowColor("right"), "settled, the theme's own styles take over")
	assert.Nil(t, s.glowColor("left"))
}

func TestWithoutAFadeFocusJumps(t *testing.T) {
	s := glowing(t, 0)
	s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, "right", s.Focused())
	assert.Nil(t, s.glowColor("right"))
}

func rgba(c interface{ RGBA() (r, g, b, a uint32) }) [4]uint32 {
	r, g, b, a := c.RGBA()
	return [4]uint32{r >> 8, g >> 8, b >> 8, a >> 8}
}
