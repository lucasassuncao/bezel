package shell

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell/internal/palette"
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

// commandsShell has twenty commands, more than any palette height lists.
func commandsShell(rows int, actions ...Action) Shell {
	for i := range 20 {
		actions = append(actions, Command(fmt.Sprintf("cmd%02d", i), "a command", nil))
	}
	s := New(Config{
		Layout:       layout.Columns(layout.Fixed("list", layout.Ratio(1, 3)), layout.Fill("detail")).Resizable(),
		Theme:        theme.Resolve(theme.Plain(), true),
		Mouse:        true,
		PaletteLines: rows,
		Actions:      actions,
	})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	return s.OpenCommands()
}

// shown is the lines the open palette's list takes, under its rule.
func shown(s Shell) int { return lipgloss.Height(s.TopOverlay().(palette.Model).View(s.Body())) - 1 }

// ruleRow is the screen row of the open palette's rule.
func ruleRow(s Shell) int { return s.Body().Y + s.Body().H - shown(s) - 1 }

func TestPaletteLinesSetsThePaletteHeight(t *testing.T) {
	require.Equal(t, palette.DefaultHeight, shown(commandsShell(0)))
	require.Equal(t, 12, shown(commandsShell(12)))
}

func TestDraggingThePaletteRuleResizesIt(t *testing.T) {
	s := commandsShell(0)
	rule := ruleRow(s)
	s, r := s.Update(click(10, rule))
	require.True(t, r.Handled)
	s, _ = s.Update(tea.MouseMotionMsg{X: 10, Y: rule - 4, Button: tea.MouseLeft})
	require.Equal(t, palette.DefaultHeight+4, shown(s))
	s, _ = s.Update(tea.MouseReleaseMsg{X: 10, Y: rule + 2})
	require.Equal(t, palette.DefaultHeight-2, shown(s))
	require.Equal(t, rule+2, ruleRow(s), "the rule sits where it was let go")

	s, _ = s.Update(tea.MouseMotionMsg{X: 10, Y: 3, Button: tea.MouseLeft})
	require.Equal(t, palette.DefaultHeight-2, shown(s), "released: motion moves nothing")

	s, r = s.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	s = deliver(t, s, r.Cmd)
	require.Nil(t, s.TopOverlay())
	require.Equal(t, palette.DefaultHeight-2, shown(s.OpenCommands()), "it reopens at the chosen height")
}

func TestResizeKeysResizeThePaletteAndResetSizesUndoesIt(t *testing.T) {
	s := commandsShell(0, Resize(), ResetSizes())
	s, r := s.Update(ctrlShift(tea.KeyDown))
	require.True(t, r.Handled)
	require.Equal(t, palette.DefaultHeight+1, shown(s), "down is taller, as for a pane")
	s, _ = s.Update(ctrlShift(tea.KeyUp))
	s, _ = s.Update(ctrlShift(tea.KeyUp))
	require.Equal(t, palette.DefaultHeight-1, shown(s))

	s, r = s.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	s = deliver(t, s, r.Cmd)
	s, _ = s.Update(press("="))
	require.Equal(t, palette.DefaultHeight, shown(s.OpenCommands()))
}

func TestWithoutResizeThePaletteKeepsItsHeight(t *testing.T) {
	s := commandsShell(0)
	s, _ = s.Update(ctrlShift(tea.KeyDown))
	require.Equal(t, palette.DefaultHeight, shown(s))
}

func TestThePaletteNeverOutgrowsTheBody(t *testing.T) {
	s := commandsShell(100)
	require.Equal(t, 20, shown(s), "every command and no more line")
	s = commandsShell(100)
	for i := range 40 {
		s = s.withCfg(func(c *Config) { c.Actions = append(c.Actions, Command(fmt.Sprintf("x%02d", i), "", nil)) })
	}
	s, _ = s.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	require.Equal(t, s.Body().H-1, shown(s), "all but the rule")
}
