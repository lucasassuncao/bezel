package list

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain, true)

func rows() []Row[int] {
	return []Row[int]{
		{Section: true, Label: "ADDED"},
		{Label: "server", Mark: "●", Value: 1},
		{Label: "logging", Mark: "●", Value: 2},
		{Section: true},
		{Section: true, Label: "AVAILABLE"},
		{Label: "cache", Mark: "+", Value: 3},
		{Label: "metrics", Mark: "+", Value: 4},
	}
}

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }
func typed(s string) tea.KeyPressMsg  { return tea.KeyPressMsg{Code: []rune(s)[0], Text: s} }

func TestCursorStartsOnAndSkipsOverHeadings(t *testing.T) {
	m := New(rows(), 10)
	require.Equal(t, "server", m.Selected().Label)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown)) // hops the spacer and the heading
	require.Equal(t, "cache", m.Selected().Label)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown)) // clamped at the end
	require.Equal(t, "metrics", m.Selected().Label)
	require.Equal(t, 4, m.Selected().Value)
}

func TestEnterChoosesASelectableRow(t *testing.T) {
	m := New(rows(), 10)
	_, act := m.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)
	require.Nil(t, New([]Row[int]{{Section: true, Label: "x"}}, 5).Selected())
}

func TestFilterNarrowsAndEnterPicksTheMatch(t *testing.T) {
	m := New(rows(), 10)
	m, _ = m.Update(typed("/"))
	require.True(t, m.IsFiltering())
	m, _ = m.Update(typed("m"))
	m, _ = m.Update(typed("e"))
	require.Equal(t, "me", m.Filter())
	require.Equal(t, "metrics", m.Selected().Label)
	require.Contains(t, ansi.Strip(m.View(th)), "/me▋")

	m, act := m.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)
	require.False(t, m.IsFiltering())
	require.Equal(t, "metrics", m.Selected().Label, "the pick becomes the cursor")

	m, _ = m.Update(typed("/"))
	m, _ = m.Update(typed("z"))
	require.Nil(t, m.Selected())
	m, _ = m.Update(press(tea.KeyBackspace))
	require.Equal(t, "", m.Filter())
	m, _ = m.Update(press(tea.KeyEscape))
	require.False(t, m.IsFiltering())
}

func TestSetRowsKeepsTheCursorByLabel(t *testing.T) {
	m := New(rows(), 10)
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "logging", m.Selected().Label)
	m = m.SetRows([]Row[int]{{Label: "cache"}, {Label: "logging"}})
	require.Equal(t, "logging", m.Selected().Label)
	m = m.SetRows([]Row[int]{{Label: "other"}})
	require.Equal(t, "other", m.Selected().Label)
}

func TestViewShowsHeadingsMarksAndTheOverflowIndicator(t *testing.T) {
	m := New(rows(), 4)
	got := ansi.Strip(m.View(th))
	lines := strings.Split(got, "\n")
	require.Len(t, lines, 4)
	require.Equal(t, " ADDED", lines[0])
	require.Equal(t, "▶ ●  server", lines[1])
	require.Equal(t, "  ●  logging", lines[2])
	require.Contains(t, lines[3], "↓ 4 more")

	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	require.Contains(t, ansi.Strip(m.View(th)), "▶ +  cache", "scrolled to keep the cursor visible")
}

func TestViewIsEmptyBeforeSized(t *testing.T) {
	require.Empty(t, New(rows(), 0).View(th))
}

func TestOneRowListShowsTheCursorNotTheIndicator(t *testing.T) {
	m := New(rows(), 1)
	m, _ = m.Update(press(tea.KeyDown))
	require.Contains(t, ansi.Strip(m.View(th)), "logging")
}

func twins() []Row[int] {
	return []Row[int]{{Label: "dup", Value: 1}, {Label: "dup", Value: 2}}
}

func TestFilterPicksTheRowItShowsAmongEqualLabels(t *testing.T) {
	m := New(twins(), 5)
	m, _ = m.Update(typed("/"))
	m, _ = m.Update(press(tea.KeyDown))
	m, act := m.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)
	require.Equal(t, 2, m.Selected().Value)
}

func TestSetRowsKeepsTheSameRowAmongEqualLabels(t *testing.T) {
	m := New(twins(), 5)
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, 2, m.SetRows(twins()).Selected().Value)
}

// The list is a value: the caller's slice must not reach into it.
func TestNewAndSetRowsCopyTheSlice(t *testing.T) {
	xs := rows()
	m := New(xs, 10)
	xs[1].Label = "changed"
	require.NotContains(t, ansi.Strip(m.View(th)), "changed")

	ys := rows()
	m = m.SetRows(ys)
	ys[1].Label = "changed"
	require.NotContains(t, ansi.Strip(m.View(th)), "changed")
}

func numbered(n int) []Row[int] {
	rs := make([]Row[int], n)
	for i := range rs {
		rs[i] = Row[int]{Label: fmt.Sprintf("item-%02d", i), Value: i}
	}
	return rs
}

// The matches scroll the way the list does: a window that keeps the cursor
// in view and counts the rows under it, the prompt on the last row.
func TestFilterScrollsItsMatchesAndCountsTheRest(t *testing.T) {
	m := New(numbered(20), 6)
	m, _ = m.Update(typed("/"))
	lines := strings.Split(ansi.Strip(m.View(th)), "\n")
	require.Len(t, lines, 6)
	require.Contains(t, lines[4], "↓ 16 more")
	require.Equal(t, "/▋", lines[5])

	for range 10 {
		m, _ = m.Update(press(tea.KeyDown))
	}
	require.Equal(t, "item-10", m.Selected().Label)
	require.Contains(t, ansi.Strip(m.View(th)), "▶ item-10")
}

// The payload keeps its type: nothing to assert between the row and the caller.
func TestSelectedValueKeepsItsType(t *testing.T) {
	m := New([]Row[string]{{Label: "a", Value: "payload"}}, 5)
	// strings.ToUpper compiles only while Value is a string, not an any.
	require.Equal(t, "PAYLOAD", strings.ToUpper(m.Selected().Value))
}

func TestWithMatchDecidesWhatTheFilterKeeps(t *testing.T) {
	filter := func(m Model[int], text string) Model[int] {
		m, _ = m.Update(typed("/"))
		for _, r := range text {
			m, _ = m.Update(typed(string(r)))
		}
		return m
	}
	require.Equal(t, "server", filter(New(rows(), 10), "e").Selected().Label, "the default keeps any label holding the text")

	prefix := New(rows(), 10).WithMatch(Prefix)
	require.Nil(t, filter(prefix, "e").Selected(), "no label starts with e")
	require.Equal(t, "cache", filter(prefix, "c").Selected().Label)
	require.Equal(t, "server", filter(prefix.WithMatch(nil), "e").Selected().Label, "nil is the default again")
}

func TestTheMatchersIgnoreCase(t *testing.T) {
	require.True(t, Contains("RV", "server"))
	require.True(t, Prefix("SER", "server"))
	require.False(t, Prefix("rv", "server"))
}

func bigList(n int) Model[int] {
	rows := make([]Row[int], n)
	for i := range rows {
		rows[i] = Row[int]{Label: fmt.Sprintf("package-%05d", i), Value: i}
	}
	return New(rows, 40)
}

func BenchmarkUpdateDown(b *testing.B) {
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	m := bigList(100_000)
	for b.Loop() {
		m, _ = m.Update(down)
	}
}

// The filter re-matches every row on each keystroke and each frame.
func BenchmarkFilter(b *testing.B) {
	th := theme.Resolve(theme.Plain, true)
	m := bigList(100_000)
	m, _ = m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '7', Text: "7"})
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	b.Run("down", func(b *testing.B) {
		for b.Loop() {
			_, _ = m.Update(down)
		}
	})
	b.Run("view", func(b *testing.B) {
		for b.Loop() {
			_ = m.View(th)
		}
	})
}
