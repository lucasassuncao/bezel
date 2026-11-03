package table

import (
	"slices"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
)

// order is how the rows are sorted: by col, the last first when desc; a
// col below zero leaves them as SetItems gave them.
type order struct {
	col  int
	desc bool
}

// Natural is the default order of a column: the cells' text, colours
// stripped and in any case, with each run of digits compared as a number,
// so v1.4 comes before v1.11 and 9 before 10.
func Natural(a, b string) bool {
	return compareText(strings.ToLower(ansi.Strip(a)), strings.ToLower(ansi.Strip(b))) < 0
}

// compareText compares a and b a run of digits or of other runes at a time.
func compareText(a, b string) int {
	for a != "" && b != "" {
		ra, rest := run(a)
		rb, restB := run(b)
		if c := compareRun(ra, rb); c != 0 {
			return c
		}
		a, b = rest, restB
	}
	return len(a) - len(b)
}

// run is the leading run of s, all digits or none, and what follows it.
func run(s string) (string, string) {
	digit := unicode.IsDigit(rune(s[0]))
	i := 1
	for i < len(s) && unicode.IsDigit(rune(s[i])) == digit {
		i++
	}
	return s[:i], s[i:]
}

// compareRun orders two runs: digits by value, else by text.
func compareRun(a, b string) int {
	if unicode.IsDigit(rune(a[0])) && unicode.IsDigit(rune(b[0])) {
		a, b = strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")
		if len(a) != len(b) {
			return len(a) - len(b)
		}
	}
	return strings.Compare(a, b)
}

// WithLess sorts column col by less instead of Natural; a col out of range
// is ignored.
func (m Model[T]) WithLess(col int, less func(a, b string) bool) Model[T] {
	if col < 0 || col >= len(m.cols) {
		return m
	}
	m.less = slices.Clone(m.less)
	if len(m.less) < len(m.cols) {
		m.less = append(m.less, make([]func(a, b string) bool, len(m.cols)-len(m.less))...)
	}
	m.less[col] = less
	return m.sorted().clamp().refilter()
}

// SetSort sorts the rows by column col, the last first when desc, and keeps
// them sorted through SetItems; a col out of range leaves them as given.
func (m Model[T]) SetSort(col int, desc bool) Model[T] {
	if col < 0 || col >= len(m.cols) {
		col = -1
	}
	m.sort = order{col: col, desc: desc}
	return m.sorted().clamp().refilter()
}

// SortedBy is the column the rows are sorted by and whether the last comes
// first; col is -1 while they are as SetItems gave them.
func (m Model[T]) SortedBy() (col int, desc bool) { return m.sort.col, m.sort.desc }

// nextSort is the order Sort steps to: this column the other way, then the
// next column ascending, round to the first.
func (m Model[T]) nextSort() order {
	switch {
	case len(m.cols) == 0:
		return order{col: -1}
	case m.sort.col < 0:
		return order{col: 0}
	case !m.sort.desc:
		return order{col: m.sort.col, desc: true}
	}
	return order{col: (m.sort.col + 1) % len(m.cols)}
}

// sortOn sorts by col, the other way when it already sorts by col.
func (m Model[T]) sortOn(col int) order {
	if m.sort.col == col {
		return order{col: col, desc: !m.sort.desc}
	}
	return order{col: col}
}

// sorted puts the rows in the table's order, stable, the cursor kept on
// its row, and measures the columns again for the header's arrow.
func (m Model[T]) sorted() Model[T] {
	if c := m.sort.col; c >= 0 && c < len(m.cols) {
		var less func(a, b string) bool
		if c < len(m.less) {
			less = m.less[c]
		}
		if less == nil {
			less = Natural
		}
		idx := make([]int, len(m.items))
		for i := range idx {
			idx[i] = i
		}
		slices.SortStableFunc(idx, func(a, b int) int {
			x, y := cell(m.items[a], c), cell(m.items[b], c)
			if m.sort.desc {
				x, y = y, x
			}
			switch {
			case less(x, y):
				return -1
			case less(y, x):
				return 1
			}
			return 0
		})
		items := make([]Item[T], len(idx))
		for k, i := range idx {
			items[k] = m.items[i]
		}
		m.win.Cursor = max(slices.Index(idx, m.win.Cursor), 0)
		m.items = items
	}
	m.widths = natural(m.titled(), cellsOf(m.items))
	return m
}

// cell is it's cell in column c, "" when the row is short.
func cell[T any](it Item[T], c int) string {
	if c < len(it.Cells) {
		return it.Cells[c]
	}
	return ""
}

// cellsOf is every row's cells, for measuring.
func cellsOf[T any](items []Item[T]) [][]string {
	rows := make([][]string, len(items))
	for i, it := range items {
		rows[i] = it.Cells
	}
	return rows
}

// titled is the columns with an arrow after the sorted one's title.
func (m Model[T]) titled() []Column {
	cols := slices.Clone(m.cols)
	if c := m.sort.col; c >= 0 && c < len(cols) {
		arrow := " ▲"
		if m.sort.desc {
			arrow = " ▼"
		}
		cols[c].Title += arrow
	}
	return cols
}

// columnAt is the column whose cells cover x of a row of View, -1 for the
// marker, a gap or past the last.
func (m Model[T]) columnAt(x int) int {
	x -= m.lead()
	for i, w := range shrink(m.cols, slices.Clone(m.widths), m.width-m.lead()) {
		if x >= 0 && x < w {
			return i
		}
		x -= w + Gap
	}
	return -1
}

// lead is the cells before the first column: the marker, and the box while
// the table is multi.
func (m Model[T]) lead() int {
	n := draw.Width(marker)
	if m.multi {
		n += draw.Width("[ ] ")
	}
	return n
}
