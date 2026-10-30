package table

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain(), true)

var pkgCols = []Column{{Title: "NAME"}, {Title: "VERSION"}}

// pkgs are n rows: pkg-0 at v0.0, pkg-1 at v0.1 and so on.
// sel is the item under the cursor, the zero Item when there is none.
func sel[T any](m Model[T]) Item[T] {
	it, _ := m.Selected()
	return it
}

// picked reports whether an item is under the cursor.
func picked[T any](m Model[T]) bool {
	_, ok := m.Selected()
	return ok
}

func pkgs(n int) []Item[int] {
	items := make([]Item[int], n)
	for i := range items {
		items[i] = Item[int]{Cells: []string{fmt.Sprintf("pkg-%d", i), fmt.Sprintf("v0.%d", i)}, Value: i}
	}
	return items
}

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func viewLines(m Model[int]) []string {
	return strings.Split(ansi.Strip(m.SetWidth(40).View(th)), "\n")
}

func TestModelViewIsTheHeaderTheRowsAndTheMoreRow(t *testing.T) {
	lines := viewLines(New(pkgCols, pkgs(10)).SetHeight(5))
	require.Len(t, lines, 5)
	require.Equal(t, "  NAME   VERSION", lines[0])
	require.Equal(t, "▶ pkg-0  v0.0", lines[1])
	require.Equal(t, "  pkg-1  v0.1", lines[2], "every row keeps the marker's width")
	require.Equal(t, "  ↓ 7 more", lines[4])
}

func TestModelMovesAndScrollsUnderTheHeader(t *testing.T) {
	m := New(pkgCols, pkgs(10)).SetHeight(5)
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, 1, m.Cursor())

	m, _ = m.Update(press(tea.KeyEnd))
	require.Equal(t, 9, sel(m).Value)
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

func TestEnterReportsTheChosenItem(t *testing.T) {
	m := New(pkgCols, pkgs(3)).SetHeight(5)
	m, _ = m.Update(press(tea.KeyDown))
	_, r := m.Update(press(tea.KeyEnter))
	require.True(t, r.Handled)
	require.Equal(t, Chosen[int]{Item: sel(m)}, r.Event)
	require.Equal(t, 1, r.Event.(Chosen[int]).Item.Value)
	_, r = New(pkgCols, pkgs(0)).SetHeight(5).Update(press(tea.KeyEnter))
	require.True(t, r.Handled, "enter is the table's even with nothing to choose")
	require.Nil(t, r.Event)
}

func TestForeignKeyIsNotHandled(t *testing.T) {
	m := New(pkgCols, pkgs(3)).SetHeight(5)
	_, r := m.Update(press('x'))
	require.Equal(t, bezel.Result{}, r)
	_, r = m.Update(tea.WindowSizeMsg{Width: 1, Height: 1})
	require.Equal(t, bezel.Result{}, r)
}

func TestAKeyAtTheEdgeIsStillTheTables(t *testing.T) {
	_, r := New(pkgCols, pkgs(3)).SetHeight(5).Update(press(tea.KeyUp))
	require.True(t, r.Handled)
	require.Nil(t, r.Event)
}

func TestModelSetItemsKeepsTheCursorByTheFirstCell(t *testing.T) {
	m := New(pkgCols, pkgs(5)).SetHeight(10)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "pkg-2", sel(m).Cells[0])

	m = m.SetItems(pkgs(5)[1:])
	require.Equal(t, "pkg-2", sel(m).Cells[0], "followed its row up")

	m = m.SetItems(append(pkgs(5)[:1], pkgs(5)[3:]...))
	require.Equal(t, "pkg-3", sel(m).Cells[0], "its row is gone: the one that took its place")

	m = m.SetItems(nil)
	require.False(t, picked(m))
	require.Equal(t, 0, m.Cursor())
}

// A value far below the window still sizes its column: scrolling to it
// moves nothing.
func TestModelSizesColumnsOverEveryRow(t *testing.T) {
	items := pkgs(10)
	items[9].Cells[0] = "a-much-longer-name"
	m := New(pkgCols, items).SetHeight(3)
	before := viewLines(m)[0]
	m, _ = m.Update(press(tea.KeyEnd))
	require.Equal(t, before, viewLines(m)[0])
	require.Equal(t, "  NAME                VERSION", before)
}

func TestModelCursorRowPaintsThePlainCells(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	mint := theme.Resolve(theme.Mint(), true)
	items := []Item[int]{
		{Cells: []string{"api", red.Render("Failed")}},
		{Cells: []string{"web", red.Render("Failed")}},
	}
	lines := strings.Split(New(pkgCols, items).SetHeight(5).SetWidth(40).View(mint), "\n")
	require.Contains(t, lines[1], mint.Cursor.Render("▶ ")+mint.Cursor.Render("api"))
	require.Contains(t, lines[1], red.Render("Failed"), "a coloured cell keeps its colour under the cursor")
	require.Equal(t, "  web   Failed", ansi.Strip(lines[2]))
	require.NotContains(t, lines[2], mint.Cursor.Render("web"), "off the cursor, plain stays plain")
}

func TestModelViewIsEmptyBeforeSized(t *testing.T) {
	require.Empty(t, New(pkgCols, pkgs(3)).SetWidth(40).View(th))
	require.Equal(t, []string{"  NAME   VERSION"}, viewLines(New(pkgCols, pkgs(3)).SetHeight(1)), "a one-row table is its header")
}

func TestModelCopiesTheSlices(t *testing.T) {
	items := pkgs(2)
	m := New(pkgCols, items).SetHeight(5)
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
	m := New(pkgCols, items).SetHeight(5)
	items[0].Cells[0] = "changed"
	require.Equal(t, "pkg-0", m.Items()[0].Cells[0])

	m.Items()[0].Cells[0] = "changed"
	sel(m).Cells[0] = "changed"
	require.Equal(t, "pkg-0", sel(m).Cells[0])

	more := pkgs(3)
	m = m.SetItems(more)
	more[2].Cells[0] = "changed"
	require.Equal(t, "pkg-2", m.Items()[2].Cells[0])
}

func TestHalfPageKeysMoveHalfThePage(t *testing.T) {
	m := New(pkgCols, pkgs(30)).SetHeight(11) // ten rows, nine a page beside the more row
	m, _ = m.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	require.Equal(t, 4, m.Cursor())
	m, _ = m.Update(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	require.Equal(t, 0, m.Cursor())
}

func TestABlurredTableDimsItsCursorAndTakesNoKeys(t *testing.T) {
	m := New(pkgCols, pkgs(3)).SetHeight(5).Blur()
	require.False(t, m.Focused())
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, 0, m.Cursor())
	require.Contains(t, m.SetWidth(40).View(th), th.Dim.Render(marker))
	require.NotContains(t, m.SetWidth(40).View(th), th.Cursor.Render("pkg-0"))
	_, r := m.Update(press(tea.KeyDown))
	require.Equal(t, bezel.Result{}, r)
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
	for _, e := range New(pkgCols, pkgs(3)).SetHeight(5).Legend() {
		keys = append(keys, e.Help.Key)
	}
	require.Equal(t, []string{"↑/↓", "pgup/pgdn", "ctrl+u/d", "home/end", "enter"}, keys)
}

func click(y int) bezel.ClickMsg {
	return bezel.ClickMsg{Pane: "p", MouseClickMsg: tea.MouseClickMsg{Y: y, Button: tea.MouseLeft}}
}

func TestSecondClickOnARowChoosesIt(t *testing.T) {
	m := New(pkgCols, pkgs(10)).SetHeight(5) // header, three rows, the more row
	m, r := m.Update(click(3))
	require.True(t, r.Handled)
	require.Nil(t, r.Event)
	require.Equal(t, 2, m.Cursor())
	_, r = m.Update(click(3))
	require.True(t, r.Handled)
	require.Equal(t, Chosen[int]{Item: sel(m)}, r.Event)
	blurred, r := m.Blur().Update(click(1))
	require.True(t, r.Handled)
	require.Equal(t, 0, blurred.Cursor(), "a click reaches a blurred table")
}

func TestClickPastTheRowsIsNotHandled(t *testing.T) {
	m := New(pkgCols, pkgs(10)).SetHeight(5)
	m, _ = m.Update(click(3))
	for _, y := range []int{0, 4, -1, 99} { // the header, the more row, off the pane
		got, r := m.Update(click(y))
		require.Equal(t, bezel.Result{}, r, "line %d", y)
		require.Equal(t, 2, got.Cursor())
	}
}

func TestWithStylesRecoloursTheHeader(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New(pkgCols, pkgs(2)).SetHeight(5).WithStyles(func(s Styles) Styles { s.Header = red; return s })
	head := strings.Split(m.SetWidth(40).View(theme.Resolve(theme.Mint(), true)), "\n")[0]
	require.Equal(t, red.Render(ansi.Strip(head)), head, "the edit holds under another theme")
}
