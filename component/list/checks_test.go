package list

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func fruit() []Row[int] {
	return []Row[int]{{Label: "Fruit", Section: true}, {Label: "apple", Value: 1}, {Label: "pear", Value: 2}, {Label: "plum", Value: 3}}
}

func space() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "} }

func TestMultiChecksTheRowUnderTheCursor(t *testing.T) {
	m := New(fruit()).WithMulti(true).SetHeight(4)
	m, r := m.Update(space())
	require.True(t, r.Handled)
	require.Equal(t, Toggled[int]{Rows: []Row[int]{{Label: "apple", Value: 1, Checked: true}}}, r.Event)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(space())
	require.Equal(t, []string{"apple", "plum"}, labels(m.Checked()))

	lines := plainLines(m)
	require.Equal(t, "  [x] apple", lines[1])
	require.Equal(t, "  [ ] pear", lines[2])
	require.Equal(t, "▶ [x] plum", lines[3])
	require.Equal(t, "Fruit", lines[0][1:], "a heading has no box")
}

func TestCheckAllChecksEveryRowThenClearsThem(t *testing.T) {
	m := New(fruit()).WithMulti(true).SetHeight(4)
	m, _ = m.Update(space())
	m, r := m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	require.Len(t, r.Event.(Toggled[int]).Rows, 2, "only the rows that changed")
	require.Len(t, m.Checked(), 3)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	require.Empty(t, m.Checked(), "all checked: it unchecks")
}

func TestChecksSurviveAReloadByLabel(t *testing.T) {
	m := New(fruit()).WithMulti(true).SetHeight(4)
	m, _ = m.Update(space())
	m = m.SetRows(fruit()[1:])
	require.Equal(t, []string{"apple"}, labels(m.Checked()))
	require.Empty(t, m.ClearChecks().Checked())
}

func TestWithoutMultiSpaceIsNotTheLists(t *testing.T) {
	m := New(fruit()).SetHeight(4)
	_, r := m.Update(space())
	require.False(t, r.Handled)
	require.Len(t, m.Legend(), 5)
	require.Len(t, m.WithMulti(true).Legend(), 7)
}

func TestFilteringSpaceIsText(t *testing.T) {
	m := New(fruit()).WithMulti(true).SetHeight(4)
	m, _ = m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m, _ = m.Update(space())
	require.Equal(t, " ", m.Filter())
	require.Empty(t, m.Checked())
}

func labels(rows []Row[int]) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Label
	}
	return out
}
