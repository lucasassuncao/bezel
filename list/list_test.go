package list

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.ThemePlain, true)

func rows() []Row {
	return []Row{
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
	require.Nil(t, New([]Row{{Section: true, Label: "x"}}, 5).Selected())
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
	m = m.SetRows([]Row{{Label: "cache"}, {Label: "logging"}})
	require.Equal(t, "logging", m.Selected().Label)
	m = m.SetRows([]Row{{Label: "other"}})
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

func twins() []Row {
	return []Row{{Label: "dup", Value: 1}, {Label: "dup", Value: 2}}
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
