package table

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
)

func versions() []Item[int] {
	return []Item[int]{
		{Cells: []string{"b", "v1.11.0"}, Value: 0},
		{Cells: []string{"a", "v1.4.2"}, Value: 1},
		{Cells: []string{"C", "v1.4.10"}, Value: 2},
	}
}

func ctrlS() tea.KeyPressMsg { return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl} }

func TestNaturalComparesDigitsAsNumbers(t *testing.T) {
	require.True(t, Natural("v1.4.2", "v1.4.10"))
	require.True(t, Natural("v1.4.10", "v1.11.0"))
	require.True(t, Natural("9", "10"))
	require.True(t, Natural("a", "B"), "in any case")
	require.False(t, Natural("x", "x"))
}

func TestSortStepsThroughTheColumnsAndKeepsTheCursorOnItsRow(t *testing.T) {
	m := New(pkgCols, versions()).SetHeight(5)
	m, r := m.Update(ctrlS())
	require.Equal(t, Sorted{Column: 0}, r.Event)
	require.Equal(t, []string{"a", "b", "C"}, names(m.Items()))
	require.Equal(t, 0, sel(m).Value, "the cursor stays on b")

	m, _ = m.Update(ctrlS())
	require.Equal(t, []string{"C", "b", "a"}, names(m.Items()))
	m, r = m.Update(ctrlS())
	require.Equal(t, Sorted{Column: 1}, r.Event)
	require.Equal(t, []int{1, 2, 0}, values(m.Items()), "versions by number")
	col, desc := m.SortedBy()
	require.Equal(t, 1, col)
	require.False(t, desc)
}

func TestWithLessSortsAColumnItsOwnWay(t *testing.T) {
	byLength := func(a, b string) bool { return len(a) < len(b) }
	m := New(pkgCols, versions()).WithLess(1, byLength).SetSort(1, false)
	require.Equal(t, []int{1, 0, 2}, values(m.Items()), "shortest version first, stable")
	require.Equal(t, m, m.WithLess(5, byLength), "out of range: ignored")

	before := New(pkgCols, versions())
	_ = before.WithLess(0, byLength)
	require.Nil(t, before.less, "the copy it came from keeps Natural")
}

func TestTheSortedTitleWearsAnArrowAndSetItemsStaysSorted(t *testing.T) {
	m := New(pkgCols, versions()).SetHeight(5).SetSort(1, true)
	require.Equal(t, "  NAME  VERSION ▼", viewLines(m)[0])
	m = m.SetItems(append(versions(), Item[int]{Cells: []string{"d", "v2"}, Value: 3}))
	require.Equal(t, 3, m.Items()[0].Value)
	require.Equal(t, -1, m.SetSort(9, false).sort.col, "out of range: as given")
}

func TestAClickOnATitleSortsByItThenTheOtherWay(t *testing.T) {
	m := New(pkgCols, versions()).SetHeight(5).SetWidth(40)
	at := func(x int) bezel.ClickMsg {
		c := click(0)
		c.X = x
		return c
	}
	m, r := m.Update(at(8))
	require.Equal(t, Sorted{Column: 1}, r.Event)
	m, r = m.Update(at(8))
	require.Equal(t, Sorted{Column: 1, Desc: true}, r.Event)
	_, r = m.Update(at(0))
	require.Nil(t, r.Event, "the marker is no column")
}

func values(items []Item[int]) []int {
	out := make([]int, len(items))
	for i, it := range items {
		out[i] = it.Value
	}
	return out
}
