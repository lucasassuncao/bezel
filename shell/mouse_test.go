package shell

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

func sizedShell() Shell {
	s, _, _ := newShell().Update(tea.WindowSizeMsg{Width: 120, Height: 30})
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
	s := sizedShell().Push(overlay.NewAlert(overlay.Info, "x", "y", theme.Resolve(theme.Plain, true)))
	r := s.Rect("list")
	_, ok := s.PaneAt(r.X+1, r.Y+1)
	require.False(t, ok)
}
