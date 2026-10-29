package browser

import (
	"charm.land/lipgloss/v2"
	"github.com/lucasassuncao/bezel/legend"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

// sel is the item under the cursor, the zero Item when there is none.
func sel(m Model) Item {
	it, _ := m.Selected()
	return it
}

func items(names ...string) []Item {
	out := make([]Item, len(names))
	for i, n := range names {
		out[i] = Item{Label: n, Detail: func() string { return "detail of " + n + "\nline 2\nline 3\nline 4" }}
	}
	return out
}

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func TestWithCurrentPutsTheCursorOnTheLabel(t *testing.T) {
	b := New(items("a", "b", "c")).WithCurrent("b")
	require.Equal(t, 1, b.Cursor())
	require.Equal(t, "b", sel(b).Label)
	require.Equal(t, 0, New(items("a")).WithCurrent("zzz").Cursor())
	require.Equal(t, Item{}, sel(New(nil)))
}

func TestArrowsMoveTheCursorAndClamp(t *testing.T) {
	b := New(items("a", "b", "c"))
	b, act := b.Update(press(tea.KeyDown))
	require.Equal(t, None, act)
	require.Equal(t, 1, b.Cursor())
	b, _ = b.Update(press(tea.KeyUp))
	b, _ = b.Update(press(tea.KeyUp)) // clamped at the top
	require.Equal(t, 0, b.Cursor())
}

func TestEnterChoosesEscDismissesTabMovesFocus(t *testing.T) {
	b := New(items("a", "b"))
	_, act := b.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)

	b, act = b.Update(press(tea.KeyTab))
	require.Equal(t, None, act)
	require.True(t, b.PreviewFocused())
	_, act = b.Update(press(tea.KeyEnter))
	require.Equal(t, None, act, "enter on the preview does nothing")

	b, act = b.Update(press(tea.KeyEscape))
	require.Equal(t, None, act, "esc on the preview only returns to the list")
	require.False(t, b.PreviewFocused())
	_, act = b.Update(press(tea.KeyEscape))
	require.Equal(t, Dismissed, act)
}

func TestPreviewScrollsWhenFocused(t *testing.T) {
	b := New(items("a")).SetPreviewHeight(2)
	require.Equal(t, "detail of a\nline 2", b.SetPreviewHeight(2).PreviewView())
	b, _ = b.Update(press(tea.KeyTab))
	b, _ = b.Update(press(tea.KeyDown))
	require.Equal(t, "line 2\nline 3", b.SetPreviewHeight(2).PreviewView())
	b, _ = b.Update(press(tea.KeyPgDown))
	require.Equal(t, "line 3\nline 4", b.SetPreviewHeight(2).PreviewView(), "clamped at the end")
	b, _ = b.Update(press(tea.KeyPgUp))
	b, _ = b.Update(press(tea.KeyPgUp))
	b, _ = b.Update(press(tea.KeyPgUp))
	require.Equal(t, "detail of a\nline 2", b.SetPreviewHeight(2).PreviewView())
}

func TestListViewMarksTheCursorAndScrollsToIt(t *testing.T) {
	b := New(items("a", "b", "c", "d")).WithCurrent("d")
	got := ansi.Strip(b.SetListHeight(2).ListView(theme.Resolve(theme.Plain, true)))
	require.Contains(t, got, "▶  d")
	require.NotContains(t, got, "a")
}

func TestSetItemsKeepsTheCursorInRange(t *testing.T) {
	b := New(items("a", "b", "c")).WithCurrent("c")
	b = b.SetItems(items("x"))
	require.Equal(t, 0, b.Cursor())
	require.Equal(t, "x", sel(b).Label)
}

// Paging past the end must not bank scroll that the way back has to undo.
func TestPreviewScrollStopsAtTheEnd(t *testing.T) {
	b := New(items("a")).SetPreviewHeight(2)
	b, _ = b.Update(press(tea.KeyTab))
	for range 5 {
		b, _ = b.Update(press(tea.KeyPgDown))
	}
	require.Equal(t, "line 3\nline 4", b.SetPreviewHeight(2).PreviewView())
	b, _ = b.Update(press(tea.KeyUp))
	require.Equal(t, "line 2\nline 3", b.SetPreviewHeight(2).PreviewView())
}

// Stepping back up from the bottom moves the cursor, not the whole window.
func TestListViewDoesNotPinTheCursorToTheBottom(t *testing.T) {
	th := theme.Resolve(theme.Plain, true)
	b := New(items("a", "b", "c", "d", "e", "f", "g", "h")).WithCurrent("h")
	b, _ = b.Update(press(tea.KeyUp))
	lines := strings.Split(ansi.Strip(b.SetListHeight(4).ListView(th)), "\n")
	require.Len(t, lines, 4)
	require.NotContains(t, lines[3], "▶", "cursor is not on the last row")
}

func TestNewAndSetItemsCopyTheSlice(t *testing.T) {
	xs := items("a", "b")
	b := New(xs)
	xs[0].Label = "changed"
	require.Equal(t, "a", sel(b).Label)

	ys := items("x")
	b = b.SetItems(ys)
	ys[0].Label = "changed"
	require.Equal(t, "x", sel(b).Label)
}

func TestWithStylesRecoloursTheCursorLabel(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	b := New(items("a", "b")).WithStyles(func(s Styles) Styles { s.Cursor = red; return s })
	require.Equal(t, red.Render("▶  a"), strings.Split(b.SetListHeight(2).ListView(theme.Resolve(theme.Mint, true)), "\n")[0])
}

func TestLegendDropsEnterWhileThePreviewHasTheKeys(t *testing.T) {
	keys := func(es []legend.Entry) (out []string) {
		for _, e := range es {
			out = append(out, e.Help().Key)
		}
		return out
	}
	b := New(items("a"))
	require.Equal(t, []string{"↑/↓", "pgup/pgdn", "tab", "enter", "esc"}, keys(b.Legend()))
	b = b.SetPreviewFocus(true)
	require.NotContains(t, keys(b.Legend()), "enter")
}
