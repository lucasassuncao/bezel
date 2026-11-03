package breadcrumb

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain(), true)

var path = []string{"kv", "apps", "billing", "prod", "db"}

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func text(m Model) string { return ansi.Strip(m.View(th)) }

func click(x int) bezel.ClickMsg {
	return bezel.ClickMsg{MouseClickMsg: tea.MouseClickMsg{X: x, Button: tea.MouseLeft}}
}

func TestViewJoinsThePartsWithTheCursorOnTheLast(t *testing.T) {
	m := New(path)
	require.Equal(t, "kv › apps › billing › prod › db", text(m))
	require.Equal(t, 4, m.Cursor())
	got, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, "db", got)
	require.Equal(t, "kv / apps", text(New(path[:2]).WithSeparator(" / ")))
}

func TestTheKeysMoveAndEnterChoosesThePathUpToTheCursor(t *testing.T) {
	m, r := New(path).Update(press(tea.KeyLeft))
	require.True(t, r.Handled)
	m, _ = m.Update(press(tea.KeyLeft))
	require.Equal(t, []string{"kv", "apps", "billing"}, m.Path())
	_, r = m.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen{Path: []string{"kv", "apps", "billing"}}, r.Event)

	m, _ = m.Update(press(tea.KeyHome))
	m, _ = m.Update(press(tea.KeyLeft))
	require.Equal(t, 0, m.Cursor(), "held at the first")
	m, _ = m.Update(press(tea.KeyEnd))
	require.Equal(t, 4, m.Cursor())
	_, r = m.Update(press('x'))
	require.False(t, r.Handled)
}

func TestANarrowRowGivesWayFromTheMiddle(t *testing.T) {
	m := New(path).SetWidth(20)
	require.Equal(t, "kv › … › prod › db", text(m))

	m = m.SetCursor(2)
	require.Equal(t, "kv › … › billing › … › db", text(m.SetWidth(25)), "the cursor's part stays")
	require.Equal(t, "kv › apps › billing", text(m.SetWidth(24).SetParts(path[:3])))
	require.Equal(t, "kv › … › db", text(New(path).SetWidth(11)))
	require.Equal(t, "kv › … › …", text(New(path).SetWidth(10)), "cut when even that is too wide")
}

func TestAClickMovesThenChooses(t *testing.T) {
	m := New(path)
	m, r := m.Update(click(6)) // "apps" starts after "kv › "
	require.True(t, r.Handled)
	require.Equal(t, 1, m.Cursor())
	_, r = m.Update(click(7))
	require.Equal(t, Chosen{Path: []string{"kv", "apps"}}, r.Event)
	_, r = m.Update(click(3))
	require.False(t, r.Handled, "a separator is no part")

	m = New(path).SetWidth(20)
	m, _ = m.Update(click(6))
	require.Equal(t, 4, m.Cursor(), "the ellipsis is no part")
	m, _ = m.Update(click(10))
	require.Equal(t, 3, m.Cursor(), "prod, after the ellipsis")
}

func TestABlurredBreadcrumbHidesItsCursorAndLeavesTheKeys(t *testing.T) {
	st := DefaultStyles(th)
	m := New(path).SetCursor(1)
	require.Contains(t, m.View(th), st.Cursor.Render("apps"))
	m = m.Blur()
	require.False(t, m.Focused())
	require.Contains(t, m.View(th), st.Part.Render("apps"))
	_, r := m.Update(press(tea.KeyLeft))
	require.False(t, r.Handled)
	require.True(t, m.Focus().Focused())
}

func TestAnEmptyPathIsQuiet(t *testing.T) {
	m := New(nil)
	_, ok := m.Selected()
	require.False(t, ok)
	require.Nil(t, m.Path())
	require.Empty(t, text(m))
	_, r := m.Update(press(tea.KeyEnter))
	require.False(t, r.Handled)
}

func TestVimKeysAddTheLetters(t *testing.T) {
	m, r := New(path).WithKeys(VimKeys()).Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	require.True(t, r.Handled)
	require.Equal(t, 3, m.Cursor())
	require.Len(t, New(path).Legend(), 3)
}

func TestSetPartsCopiesThePath(t *testing.T) {
	parts := []string{"a", "b"}
	m := New(parts)
	parts[0] = "x"
	require.Equal(t, []string{"a", "b"}, m.Parts())
}

func FuzzBreadcrumbStaysInItsWidth(f *testing.F) {
	f.Add(20, 2)
	f.Add(3, 0)
	f.Fuzz(func(t *testing.T, w, cursor int) {
		w = w%60 + 1
		if w < 1 {
			w = 1
		}
		m := New(path).SetWidth(w).SetCursor(cursor)
		require.LessOrEqual(t, ansi.StringWidth(m.View(th)), w)
	})
}
