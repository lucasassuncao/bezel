package shell

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

// mintShell is a sized shell on a theme with a background to blend toward.
func mintShell(t *testing.T) Shell {
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("BEZEL_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
	th := theme.Resolve(theme.Mint(), true)
	return sized(New(Config{Layout: twoPane(), Theme: th, Title: "demo"}))
}

// cellAt is the cell at column x of row y of the screen s draws.
func cellAt(s Shell, x, y int) cells.Cell {
	return cells.Parse(s.render(demoPanes()))[y][x]
}

// openConfirm opens a dialog the way an action does, through Update.
func openConfirm(s Shell) (Shell, tea.Cmd) {
	s, r := s.Update(overlay.Open(overlay.NewConfirm("Delete", "sure?", nil, s.cfg.Theme))())
	return s, r.Cmd
}

func TestTheBackdropFadesInBehindAModalOpenedInUpdate(t *testing.T) {
	s := mintShell(t)
	th := s.cfg.Theme
	require.InDelta(t, 0.4, th.Backdrop, 1e-9, "on by default on a dark theme")
	before := cellAt(s, 1, 0) // the title, far from the dialog

	s, cmd := openConfirm(s)
	require.NotNil(t, cmd, "the fade asks for frames")
	require.Equal(t, before.Fg, cellAt(s, 1, 0).Fg, "no frame yet: nothing has stepped back")

	start := time.Now()
	s, r := s.Update(s.fade.FrameAt(start))
	require.True(t, r.Handled)
	s, _ = s.Update(s.fade.FrameAt(start.Add(fadeIn)))
	require.False(t, s.fade.Active(), "settled")
	want := cells.Mix(before.Fg, th.Background, th.Backdrop)
	require.Equal(t, want, cellAt(s, 1, 0).Fg, "blended toward the background, not black")
}

func TestABackdropBlendsOnceUnderTheTopModal(t *testing.T) {
	s := mintShell(t)
	plain := cellAt(s, 1, 0).Fg
	s, _ = openConfirm(s)
	s, cmd := openConfirm(s)
	require.Nil(t, cmd, "a second modal starts no fade")
	s = settle(s)
	once := cells.Mix(plain, s.cfg.Theme.Background, s.cfg.Theme.Backdrop)
	require.Equal(t, once, cellAt(s, 1, 0).Fg)
}

// settle runs the backdrop's fade to its end.
func settle(s Shell) Shell {
	s, _ = s.Update(s.fade.FrameAt(time.Now()))
	s, _ = s.Update(s.fade.FrameAt(time.Now().Add(time.Hour)))
	return s
}

func TestPushShowsTheBackdropAtOnceAndClosingClearsIt(t *testing.T) {
	s := mintShell(t)
	plain := cellAt(s, 1, 0).Fg
	s = s.Push(overlay.NewConfirm("Delete", "sure?", nil, s.cfg.Theme))
	require.NotEqual(t, plain, cellAt(s, 1, 0).Fg, "a Push outside Update has no frames to fade with")

	s, _ = s.Update(overlay.CloseMsg{})
	require.Equal(t, plain, cellAt(s, 1, 0).Fg)
	require.False(t, s.fade.Active())
}

func TestABackdropOfZeroOrWithoutABackgroundLeavesTheScreen(t *testing.T) {
	s := mintShell(t)
	th := s.cfg.Theme
	th.Backdrop = 0
	s = s.SetTheme(th)
	plain := s.render(demoPanes())
	s = s.Push(overlay.NewConfirm("Delete", "sure?", nil, th))
	require.Equal(t, strings.Split(plain, "\n")[0], strings.Split(s.render(demoPanes()), "\n")[0])

	require.Zero(t, theme.Resolve(theme.Plain(), true).Backdrop, "no background to blend toward")
	require.InDelta(t, 0.3, theme.Resolve(theme.Mint(), false).Backdrop, 1e-9, "lighter on a light theme")
}
