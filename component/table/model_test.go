package table

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain, true)

var pkgCols = []Column{{Title: "NAME"}, {Title: "VERSION"}}

// pkgs are n rows: pkg-0 at v0.0, pkg-1 at v0.1 and so on.
func pkgs(n int) []Item[int] {
	items := make([]Item[int], n)
	for i := range items {
		items[i] = Item[int]{Cells: []string{fmt.Sprintf("pkg-%d", i), fmt.Sprintf("v0.%d", i)}, Value: i}
	}
	return items
}

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func viewLines(m Model[int]) []string {
	return strings.Split(ansi.Strip(m.View(th, 40)), "\n")
}

func TestModelViewIsTheHeaderTheRowsAndTheMoreRow(t *testing.T) {
	lines := viewLines(New(pkgCols, pkgs(10), 5))
	require.Len(t, lines, 5)
	require.Equal(t, "  NAME   VERSION", lines[0])
	require.Equal(t, "▶ pkg-0  v0.0", lines[1])
	require.Equal(t, "  pkg-1  v0.1", lines[2], "every row keeps the marker's width")
	require.Equal(t, "  ↓ 7 more", lines[4])
}

func TestModelMovesAndScrollsUnderTheHeader(t *testing.T) {
	m := New(pkgCols, pkgs(10), 5)
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, 1, m.Cursor())

	m, _ = m.Update(press(tea.KeyEnd))
	require.Equal(t, 9, m.Selected().Value)
	lines := viewLines(m)
	require.Equal(t, "  NAME   VERSION", lines[0], "the header stays put")
	require.Equal(t, "▶ pkg-9  v0.9", lines[4])
	require.Equal(t, layout.Scroll{Offset: 6, Visible: 4, Total: 10}, m.Scroll())

	m, _ = m.Update(press(tea.KeyPgUp))
	require.Equal(t, 6, m.Cursor(), "a page is the rows shown, less the more row")
	m, _ = m.Update(press(tea.KeyHome))
	require.Equal(t, 0, m.Cursor())
	m, _ = m.Update(press(tea.KeyPgDown))
	require.Equal(t, 3, m.Cursor())
}

func TestModelEnterChoosesARow(t *testing.T) {
	_, act := New(pkgCols, pkgs(3), 5).Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)
	_, act = New(pkgCols, pkgs(0), 5).Update(press(tea.KeyEnter))
	require.Equal(t, None, act, "nothing to choose")
	_, act = New(pkgCols, pkgs(3), 5).Update(press('x'))
	require.Equal(t, None, act)
}

func TestModelSetItemsKeepsTheCursorByTheFirstCell(t *testing.T) {
	m := New(pkgCols, pkgs(5), 10)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "pkg-2", m.Selected().Cells[0])

	m = m.SetItems(pkgs(5)[1:])
	require.Equal(t, "pkg-2", m.Selected().Cells[0], "followed its row up")

	m = m.SetItems(append(pkgs(5)[:1], pkgs(5)[3:]...))
	require.Equal(t, "pkg-3", m.Selected().Cells[0], "its row is gone: the one that took its place")

	m = m.SetItems(nil)
	require.Nil(t, m.Selected())
	require.Equal(t, 0, m.Cursor())
}

// A value far below the window still sizes its column: scrolling to it
// moves nothing.
func TestModelSizesColumnsOverEveryRow(t *testing.T) {
	items := pkgs(10)
	items[9].Cells[0] = "a-much-longer-name"
	m := New(pkgCols, items, 3)
	before := viewLines(m)[0]
	m, _ = m.Update(press(tea.KeyEnd))
	require.Equal(t, before, viewLines(m)[0])
	require.Equal(t, "  NAME                VERSION", before)
}

func TestModelCursorRowPaintsThePlainCells(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	mint := theme.Resolve(theme.Mint, true)
	items := []Item[int]{
		{Cells: []string{"api", red.Render("Failed")}},
		{Cells: []string{"web", red.Render("Failed")}},
	}
	lines := strings.Split(New(pkgCols, items, 5).View(mint, 40), "\n")
	require.Contains(t, lines[1], mint.Cursor.Render("▶ ")+mint.Cursor.Render("api"))
	require.Contains(t, lines[1], red.Render("Failed"), "a coloured cell keeps its colour under the cursor")
	require.Equal(t, "  web   Failed", ansi.Strip(lines[2]))
	require.NotContains(t, lines[2], mint.Cursor.Render("web"), "off the cursor, plain stays plain")
}

func TestModelViewIsEmptyBeforeSized(t *testing.T) {
	require.Empty(t, New(pkgCols, pkgs(3), 0).View(th, 40))
	require.Equal(t, []string{"  NAME   VERSION"}, viewLines(New(pkgCols, pkgs(3), 1)), "a one-row table is its header")
}

func TestModelCopiesTheSlices(t *testing.T) {
	items := pkgs(2)
	m := New(pkgCols, items, 5)
	items[0].Value = 99
	require.Equal(t, 0, m.Items()[0].Value)
	got := m.Items()
	got[1].Value = 99
	require.Equal(t, 1, m.Items()[1].Value)
}

// The cells are the table's own: writing to the slice handed in, or to one
// handed out, changes nothing it shows.
func TestModelCopiesTheCells(t *testing.T) {
	items := pkgs(2)
	m := New(pkgCols, items, 5)
	items[0].Cells[0] = "changed"
	require.Equal(t, "pkg-0", m.Items()[0].Cells[0])

	m.Items()[0].Cells[0] = "changed"
	m.Selected().Cells[0] = "changed"
	require.Equal(t, "pkg-0", m.Selected().Cells[0])

	more := pkgs(3)
	m = m.SetItems(more)
	more[2].Cells[0] = "changed"
	require.Equal(t, "pkg-2", m.Items()[2].Cells[0])
}

func TestHalfPageKeysMoveHalfThePage(t *testing.T) {
	m := New(pkgCols, pkgs(30), 11) // ten rows, nine a page beside the more row
	m, _ = m.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	require.Equal(t, 4, m.Cursor())
	m, _ = m.Update(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	require.Equal(t, 0, m.Cursor())
}

func TestABlurredTableDimsItsCursorAndTakesNoKeys(t *testing.T) {
	m := New(pkgCols, pkgs(3), 5).Blur()
	require.False(t, m.Focused())
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, 0, m.Cursor())
	require.Contains(t, m.View(th, 40), th.Dim.Render(marker))
	require.NotContains(t, m.View(th, 40), th.Cursor.Render("pkg-0"))
	m, _ = m.Focus().Update(press(tea.KeyDown))
	require.Equal(t, 1, m.Cursor())
}

func TestFromValuesSplitsLinesIntoCells(t *testing.T) {
	items := FromValues[int]("a,1\r\n\nb,2\n", ",")
	require.Len(t, items, 2)
	require.Equal(t, []string{"b", "2"}, items[1].Cells)
}

func TestTheLegendNamesTheKeys(t *testing.T) {
	var keys []string
	for _, e := range New(pkgCols, pkgs(3), 5).Legend() {
		keys = append(keys, e.Help().Key)
	}
	require.Equal(t, []string{"↑/↓", "pgup/pgdn", "ctrl+u/d", "home/end", "enter"}, keys)
}

func TestAClickMovesTheCursorAndASecondChooses(t *testing.T) {
	m := New(pkgCols, pkgs(10), 5) // header, three rows, the more row
	m, act := m.Click(3)
	require.Equal(t, None, act)
	require.Equal(t, 2, m.Cursor())
	_, act = m.Click(3)
	require.Equal(t, Chosen, act)
	for _, y := range []int{0, 4, -1} {
		got, act := m.Click(y)
		require.Equal(t, None, act, "line %d", y)
		require.Equal(t, 2, got.Cursor())
	}
	blurred, _ := m.Blur().Click(1)
	require.Equal(t, 0, blurred.Cursor(), "a click reaches a blurred table")
}
