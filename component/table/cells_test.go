package table

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestWithCellsMovesBetweenColumnsAndReadsTheCell(t *testing.T) {
	m := New(pkgCols, pkgs(3)).WithCells(true).SetHeight(4)
	v, ok := m.Cell()
	require.True(t, ok)
	require.Equal(t, "pkg-0", v)

	m, r := m.Update(press(tea.KeyRight))
	require.True(t, r.Handled)
	m, _ = m.Update(press(tea.KeyRight))
	require.Equal(t, 1, m.Column(), "held at the last column")
	v, _ = m.Cell()
	require.Equal(t, "v0.0", v)

	m, r = m.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen[int]{Item: sel(m)}, r.Event)
	require.Equal(t, 1, m.Column(), "the chosen cell is read from the model")
}

func TestTheCursorsCellWearsTheCellStyle(t *testing.T) {
	m := New(pkgCols, pkgs(2)).WithCells(true).SetHeight(3).SetWidth(40)
	m, _ = m.Update(press(tea.KeyRight))
	require.Contains(t, m.View(th), DefaultStyles(th).Cell.Render("v0.0"))
	require.Contains(t, viewLines(m)[1], "▶ pkg-0  v0.0")
}

func TestWithoutCellsTheArrowsAreNotTheTables(t *testing.T) {
	m := New(pkgCols, pkgs(2)).SetHeight(3)
	_, r := m.Update(press(tea.KeyRight))
	require.False(t, r.Handled)
	require.Zero(t, m.Column())
	require.Len(t, m.WithCells(true).Legend(), len(m.Legend())+1)
}

func TestAClickPutsTheCursorOnTheCell(t *testing.T) {
	m := New(pkgCols, pkgs(3)).WithCells(true).SetHeight(4).SetWidth(40)
	c := click(2)
	c.X = 9
	m, _ = m.Update(c)
	require.Equal(t, 1, m.Cursor())
	require.Equal(t, 1, m.Column())
}
