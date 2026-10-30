package shell

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

func sizedShell() Shell {
	s, _ := newShell().Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	return s
}

func TestPaneAtFindsThePaneUnderACell(t *testing.T) {
	s := sizedShell()
	list, detail := s.Rect("list"), s.Rect("detail")
	got, ok := s.PaneAt(list.X, list.Y)
	require.True(t, ok)
	require.Equal(t, "list", got, "its corner is its own")
	got, _ = s.PaneAt(detail.X+detail.W-1, detail.Y+detail.H-1)
	require.Equal(t, "detail", got)
	_, ok = s.PaneAt(0, 0)
	require.False(t, ok, "the header is no pane")
}

func TestMouseMovesTheMessageIntoThePanesBody(t *testing.T) {
	s := sizedShell()
	inner := draw.InnerRect(s.Rect("detail"))
	pane, msg, ok := s.Mouse(tea.MouseClickMsg{X: inner.X + 3, Y: inner.Y + 2, Button: tea.MouseLeft})
	require.True(t, ok)
	require.Equal(t, "detail", pane)
	require.Equal(t, tea.MouseClickMsg{X: 3, Y: 2, Button: tea.MouseLeft}, msg, "the same kind of message, in body cells")

	drag := s.MouseIn("detail", tea.MouseMotionMsg{X: inner.X - 2, Y: inner.Y})
	require.Equal(t, tea.MouseMotionMsg{X: -2, Y: 0}, drag, "a drag that left the pane keeps its cell")
	_, wheel, _ := s.Mouse(tea.MouseWheelMsg{X: inner.X, Y: inner.Y, Button: tea.MouseWheelDown})
	require.IsType(t, tea.MouseWheelMsg{}, wheel)
	_, release, _ := s.Mouse(tea.MouseReleaseMsg{X: inner.X, Y: inner.Y})
	require.IsType(t, tea.MouseReleaseMsg{}, release)
}

func TestADialogCoversThePanes(t *testing.T) {
	s := sizedShell().Push(overlay.NewAlert(overlay.Info, "x", "y", theme.Resolve(theme.Plain(), true)))
	r := s.Rect("list")
	_, ok := s.PaneAt(r.X+1, r.Y+1)
	require.False(t, ok)
}

func click(x, y int) tea.MouseClickMsg { return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft} }

func TestAClickOnATabShowsIt(t *testing.T) {
	s, _ := newShell(stubTab{name: "alpha"}, stubTab{name: "beta"}).Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	row := strings.Split(ansi.Strip(s.View(demoPanes())), "\n")[s.headerRows()]
	s, r := s.Update(click(strings.Index(row, "beta"), s.headerRows()))
	require.True(t, r.Handled)
	require.Equal(t, 1, s.ActiveTab())
	s, r = s.Update(click(119, s.headerRows()))
	require.True(t, r.Handled, "the tab row is the shell's even off a tab")
	require.Equal(t, 1, s.ActiveTab())
}

// clickOf is the ClickMsg cmd sends, alone or in a batch; false for none.
func clickOf(cmd tea.Cmd) (ClickMsg, bool) {
	if cmd == nil {
		return ClickMsg{}, false
	}
	switch msg := cmd().(type) {
	case ClickMsg:
		return msg, true
	case tea.BatchMsg:
		for _, c := range msg {
			if got, ok := clickOf(c); ok {
				return got, true
			}
		}
	}
	return ClickMsg{}, false
}

func TestAClickOnAPaneFocusesItAndComesAsAClickMsg(t *testing.T) {
	s := sizedShell()
	require.Equal(t, "list", s.Focused())
	inner := draw.InnerRect(s.Rect("detail"))
	s, r := s.Update(click(inner.X+3, inner.Y+2))
	require.True(t, r.Handled)
	require.Equal(t, "detail", s.Focused())
	got, ok := clickOf(r.Cmd)
	require.True(t, ok, "the focus and the click, in one go")
	require.Equal(t, "detail", got.Pane)
	require.Equal(t, 3, got.X)
	require.Equal(t, 2, got.Y)

	_, r = s.Update(click(inner.X, inner.Y))
	require.True(t, r.Handled)
	got, _ = clickOf(r.Cmd)
	require.Equal(t, ClickMsg{Pane: "detail", MouseClickMsg: tea.MouseClickMsg{Button: tea.MouseLeft}}, got, "on the focused pane, just the click")

	_, r = s.Update(tea.MouseClickMsg{X: inner.X, Y: inner.Y, Button: tea.MouseRight})
	require.False(t, r.Handled, "only the left button")
	_, r = s.Update(click(0, 0))
	require.False(t, r.Handled, "the header is no pane: the click is the app's as it came")
}

func TestScreenTurnsTheMouseOnWhenAsked(t *testing.T) {
	s := sizedShell()
	v := s.Screen(demoPanes())
	require.True(t, v.AltScreen)
	require.Equal(t, tea.MouseModeNone, v.MouseMode)
	on := New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.Plain(), true), Mouse: true})
	require.Equal(t, tea.MouseModeCellMotion, on.Screen(demoPanes()).MouseMode)
}

func TestAClickWithADialogOpenIsTheDialogs(t *testing.T) {
	s := sizedShell().Push(overlay.NewAlert(overlay.Info, "Saved", "done", theme.Resolve(theme.Plain(), true)))
	_, r := s.Update(click(0, 0))
	require.True(t, r.Handled)
	require.Nil(t, r.Cmd, "off its buttons it does nothing")
	require.Equal(t, "list", s.Focused())
}

func TestShellClickMsgIsBezelClickMsg(t *testing.T) {
	pane := func(c bezel.ClickMsg) string { return c.Pane }
	require.Equal(t, "list", pane(ClickMsg{Pane: "list"}))
}
