package table

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func space() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "} }

func names(items []Item[int]) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = first(it)
	}
	return out
}

func TestMultiChecksTheRowUnderTheCursor(t *testing.T) {
	m := New(pkgCols, pkgs(3)).WithMulti(true).SetHeight(4)
	m, r := m.Update(space())
	require.True(t, r.Handled)
	require.Equal(t, []string{"pkg-0"}, names(r.Event.(Toggled[int]).Items))
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(space())
	require.Equal(t, []string{"pkg-0", "pkg-2"}, names(m.Checked()))

	lines := viewLines(m)
	require.Equal(t, "      NAME   VERSION", lines[0], "the header makes room for the box")
	require.Equal(t, "  [x] pkg-0  v0.0", lines[1])
	require.Equal(t, "  [ ] pkg-1  v0.1", lines[2])
	require.Equal(t, "▶ [x] pkg-2  v0.2", lines[3])
}

func TestCheckAllChecksEveryRowThenClearsThem(t *testing.T) {
	m := New(pkgCols, pkgs(3)).WithMulti(true).SetHeight(4)
	m, _ = m.Update(space())
	m, r := m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	require.Len(t, r.Event.(Toggled[int]).Items, 2, "only the rows that changed")
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	require.Empty(t, m.Checked())
}

func TestChecksSurviveAReloadByFirstCell(t *testing.T) {
	m := New(pkgCols, pkgs(3)).WithMulti(true).SetHeight(4)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(space())
	m = m.SetItems(pkgs(3)[1:])
	require.Equal(t, []string{"pkg-1"}, names(m.Checked()))
	require.Empty(t, m.ClearChecks().Checked())
}

func TestWithoutMultiSpaceIsNotTheTables(t *testing.T) {
	m := New(pkgCols, pkgs(3)).SetHeight(4)
	_, r := m.Update(space())
	require.False(t, r.Handled)
	require.Len(t, m.WithMulti(true).Legend(), len(m.Legend())+2)
}
