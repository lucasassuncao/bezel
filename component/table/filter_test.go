package table

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

// filtering is m with the / prompt open on text.
func filtering(m Model[int], text string) Model[int] {
	m, _ = m.Update(press('/'))
	for _, r := range text {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func TestFilterNarrowsTheRowsAndEnterPicksTheMatch(t *testing.T) {
	m := filtering(New(pkgCols, pkgs(12)).SetHeight(6), "v0.1")
	require.True(t, m.Filtering())
	require.Equal(t, "v0.1", m.Filter())
	lines := viewLines(m)
	require.Len(t, lines, 6)
	require.Equal(t, "  NAME    VERSION", lines[0], "the header stays")
	require.Equal(t, "▶ pkg-1   v0.1", lines[1])
	require.Equal(t, "  pkg-10  v0.10", lines[2])
	require.Equal(t, "/v0.1▋", lines[5], "the prompt is the last line")

	m, _ = m.Update(press(tea.KeyDown))
	m, r := m.Update(press(tea.KeyEnter))
	require.False(t, m.Filtering())
	require.Equal(t, Chosen[int]{Item: sel(m)}, r.Event)
	require.Equal(t, 10, sel(m).Value)
	require.Equal(t, 10, m.Cursor(), "the table's cursor lands on the pick")
}

func TestFilterEscLeavesWithNothingAndEraseTakesALetter(t *testing.T) {
	m := filtering(New(pkgCols, pkgs(3)).SetHeight(6), "zz")
	require.False(t, picked(m), "nothing matched")
	_, r := m.Update(press(tea.KeyEnter))
	require.True(t, r.Handled)
	require.Nil(t, r.Event)

	m, _ = m.Update(press(tea.KeyBackspace))
	require.Equal(t, "z", m.Filter())
	m, r = m.Update(press(tea.KeyEscape))
	require.True(t, r.Handled)
	require.False(t, m.Filtering())
	require.Equal(t, 0, m.Cursor())
}

func TestFilterLegendFollowsThePrompt(t *testing.T) {
	var keys []string
	for _, e := range filtering(New(pkgCols, pkgs(3)).SetHeight(5), "").Legend() {
		keys = append(keys, e.Help.Key)
	}
	require.Equal(t, []string{"↑/↓", "enter", "esc"}, keys)
}

func TestFuzzyFilterRanksTheBestRowFirst(t *testing.T) {
	items := []Item[int]{{Cells: []string{"carousel", "1"}}, {Cells: []string{"cherry", "2"}, Value: 1}}
	m := filtering(New(pkgCols, items).SetHeight(6).WithFuzzy(true), "chy")
	require.Equal(t, 1, sel(m).Value)
	m = filtering(New(pkgCols, items).SetHeight(6).WithFuzzy(true).WithMatch(nil), "chy")
	require.False(t, picked(m), "WithMatch ends it")
}

func TestWithMatchDecidesWhatTheFilterKeeps(t *testing.T) {
	prefix := func(q, text string) bool { return len(text) >= len(q) && text[:len(q)] == q }
	m := filtering(New(pkgCols, pkgs(3)).SetHeight(6).WithMatch(prefix), "v0")
	require.False(t, picked(m), "rows start with pkg")
	require.True(t, Contains("V0", "pkg-0 v0.0"))
}

func TestAClickPicksAMatchWhileFiltering(t *testing.T) {
	m := filtering(New(pkgCols, pkgs(12)).SetHeight(6), "v0.1")
	m, r := m.Update(click(2))
	require.True(t, r.Handled)
	require.Nil(t, r.Event)
	m, r = m.Update(click(2))
	require.Equal(t, 10, r.Event.(Chosen[int]).Item.Value)
	require.False(t, m.Filtering())
	_, r = filtering(New(pkgCols, pkgs(3)).SetHeight(6), "").Update(click(5))
	require.False(t, r.Handled, "the prompt is not a row")
}

func TestSetItemsRematchesTheOpenFilter(t *testing.T) {
	m := filtering(New(pkgCols, pkgs(3)).SetHeight(6), "pkg-1")
	m = m.SetItems(pkgs(12))
	require.True(t, m.Filtering())
	require.Equal(t, 3, m.Scroll().Total, "pkg-1, pkg-10, pkg-11")
}

func TestVimKeysAreTextAtThePrompt(t *testing.T) {
	m := filtering(New(pkgCols, pkgs(3)).SetHeight(6).WithKeys(VimKeys()), "j")
	require.Equal(t, "j", m.Filter())
}
