package browser

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

func items(names ...string) []Item {
	out := make([]Item, len(names))
	for i, n := range names {
		out[i] = Item{Label: n, Detail: func() string { return "detail of " + n + "\nline 2\nline 3\nline 4" }}
	}
	return out
}

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func TestNewSelectsCurrentAndClampsCursor(t *testing.T) {
	b := New(items("a", "b", "c"), "b")
	require.Equal(t, 1, b.Cursor())
	require.Equal(t, "b", b.Selected().Label)
	require.Equal(t, 0, New(items("a"), "zzz").Cursor())
	require.Equal(t, Item{}, New(nil, "").Selected())
}

func TestArrowsMoveTheCursorAndClamp(t *testing.T) {
	b := New(items("a", "b", "c"), "")
	b, act := b.Update(press(tea.KeyDown))
	require.Equal(t, None, act)
	require.Equal(t, 1, b.Cursor())
	b, _ = b.Update(press(tea.KeyUp))
	b, _ = b.Update(press(tea.KeyUp)) // clamped at the top
	require.Equal(t, 0, b.Cursor())
}

func TestEnterChoosesEscDismissesTabMovesFocus(t *testing.T) {
	b := New(items("a", "b"), "")
	_, act := b.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)

	b, act = b.Update(press(tea.KeyTab))
	require.Equal(t, None, act)
	require.True(t, b.PreviewFocus)
	_, act = b.Update(press(tea.KeyEnter))
	require.Equal(t, None, act, "enter on the preview does nothing")

	b, act = b.Update(press(tea.KeyEscape))
	require.Equal(t, None, act, "esc on the preview only returns to the list")
	require.False(t, b.PreviewFocus)
	_, act = b.Update(press(tea.KeyEscape))
	require.Equal(t, Dismissed, act)
}

func TestPreviewScrollsWhenFocused(t *testing.T) {
	b := New(items("a"), "").SetPreviewHeight(2)
	require.Equal(t, "detail of a\nline 2", b.PreviewView(2))
	b, _ = b.Update(press(tea.KeyTab))
	b, _ = b.Update(press(tea.KeyDown))
	require.Equal(t, "line 2\nline 3", b.PreviewView(2))
	b, _ = b.Update(press(tea.KeyPgDown))
	require.Equal(t, "line 3\nline 4", b.PreviewView(2), "clamped at the end")
	b, _ = b.Update(press(tea.KeyPgUp))
	b, _ = b.Update(press(tea.KeyPgUp))
	b, _ = b.Update(press(tea.KeyPgUp))
	require.Equal(t, "detail of a\nline 2", b.PreviewView(2))
}

func TestListViewMarksTheCursorAndScrollsToIt(t *testing.T) {
	b := New(items("a", "b", "c", "d"), "d")
	got := ansi.Strip(b.ListView(theme.Resolve(theme.ThemePlain, true), 2))
	require.Contains(t, got, "▶  d")
	require.NotContains(t, got, "a")
}

func TestSetItemsKeepsTheCursorInRange(t *testing.T) {
	b := New(items("a", "b", "c"), "c")
	b = b.SetItems(items("x"))
	require.Equal(t, 0, b.Cursor())
	require.Equal(t, "x", b.Selected().Label)
}

// Paging past the end must not bank scroll that the way back has to undo.
func TestPreviewScrollStopsAtTheEnd(t *testing.T) {
	b := New(items("a"), "").SetPreviewHeight(2)
	b, _ = b.Update(press(tea.KeyTab))
	for range 5 {
		b, _ = b.Update(press(tea.KeyPgDown))
	}
	require.Equal(t, "line 3\nline 4", b.PreviewView(2))
	b, _ = b.Update(press(tea.KeyUp))
	require.Equal(t, "line 2\nline 3", b.PreviewView(2))
}

// Stepping back up from the bottom moves the cursor, not the whole window.
func TestListViewDoesNotPinTheCursorToTheBottom(t *testing.T) {
	th := theme.Resolve(theme.ThemePlain, true)
	b := New(items("a", "b", "c", "d", "e", "f", "g", "h"), "h")
	b, _ = b.Update(press(tea.KeyUp))
	lines := strings.Split(ansi.Strip(b.ListView(th, 4)), "\n")
	require.Len(t, lines, 4)
	require.NotContains(t, lines[3], "▶", "cursor is not on the last row")
}

func TestNewAndSetItemsCopyTheSlice(t *testing.T) {
	xs := items("a", "b")
	b := New(xs, "")
	xs[0].Label = "changed"
	require.Equal(t, "a", b.Selected().Label)

	ys := items("x")
	b = b.SetItems(ys)
	ys[0].Label = "changed"
	require.Equal(t, "x", b.Selected().Label)
}
