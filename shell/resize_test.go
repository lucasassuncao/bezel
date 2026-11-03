package shell

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

func resizable(actions ...Action) Shell {
	s := New(Config{
		Layout:  layout.Columns(layout.Fixed("list", layout.Ratio(1, 3)), layout.Fill("detail")).Resizable(),
		Theme:   theme.Resolve(theme.Plain(), true),
		Title:   "demo",
		Mouse:   true,
		Actions: actions,
	})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	return s
}

func TestDraggingAnEdgeResizesThePanes(t *testing.T) {
	s := resizable()
	list := s.Rect("list")
	y := list.Y + 3
	s, r := s.Update(click(list.X+list.W-1, y))
	require.True(t, r.Handled)
	require.Nil(t, r.Cmd, "a press on an edge is no click on a pane")
	s, r = s.Update(tea.MouseMotionMsg{X: list.X + list.W + 9, Y: y, Button: tea.MouseLeft})
	require.Equal(t, list.W+10, s.Rect("list").W)
	require.Equal(t, ResizedMsg{}, r.Cmd(), "the app sizes its models again")
	_, r = s.Update(tea.MouseMotionMsg{X: list.X + list.W + 9, Y: y, Button: tea.MouseLeft})
	require.Nil(t, r.Cmd, "the edge did not move")
	s, r = s.Update(tea.MouseReleaseMsg{X: list.X + list.W + 19, Y: y})
	require.True(t, r.Handled)
	require.Equal(t, list.W+20, s.Rect("list").W)

	s, r = s.Update(tea.MouseMotionMsg{X: 5, Y: y, Button: tea.MouseLeft})
	require.False(t, r.Handled, "released: motion is the app's again")
	require.Equal(t, list.W+20, s.Rect("list").W)
}

func TestAClickInsideAPaneIsStillAClick(t *testing.T) {
	s := resizable()
	list := s.Rect("list")
	_, r := s.Update(click(list.X+4, list.Y+3))
	require.NotNil(t, r.Cmd)
	require.IsType(t, ClickMsg{}, r.Cmd())
}

func ctrlShift(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Mod: tea.ModCtrl | tea.ModShift}
}

func TestResizeGrowsTheFocusedPaneAndResetSizesUndoesIt(t *testing.T) {
	s := resizable(Resize(), ResetSizes())
	w := s.Rect("list").W
	s, r := s.Update(ctrlShift(tea.KeyRight))
	require.True(t, r.Handled)
	require.Equal(t, w+2, s.Rect("list").W)
	s, _ = s.Update(ctrlShift(tea.KeyLeft))
	s, _ = s.Update(ctrlShift(tea.KeyLeft))
	require.Equal(t, w-2, s.Rect("list").W)

	require.Equal(t, ResizedMsg{}, r.Cmd())

	s = s.SetFocus("detail")
	s, _ = s.Update(ctrlShift(tea.KeyRight))
	require.Equal(t, w-4, s.Rect("list").W, "the last pane grows by its left edge")
	s, r = s.Update(ctrlShift(tea.KeyDown))
	require.Equal(t, w-4, s.Rect("list").W, "no Resizable rows: nothing moves")
	require.Nil(t, r.Cmd)

	s, _ = s.Update(press("="))
	require.Equal(t, w, s.Rect("list").W)
}
