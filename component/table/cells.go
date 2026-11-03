package table

import (
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
)

// WithCells puts the cursor on a cell, not a whole row: Left and Right move
// it between the columns, Cell reads it and Chosen says its column.
func (m Model[T]) WithCells(on bool) Model[T] {
	m.cells = on
	return m.clampColumn()
}

// Column is the column the cursor is on, 0 unless the table is WithCells.
func (m Model[T]) Column() int { return m.col }

// Cell is the text of the cell under the cursor, false with no row there.
func (m Model[T]) Cell() (string, bool) {
	it, ok := m.Selected()
	if !ok {
		return "", false
	}
	return cell(it, m.col), true
}

// moveColumn steps the cursor delta columns, held at the ends.
func (m Model[T]) moveColumn(delta int) Model[T] {
	m.col += delta
	return m.clampColumn()
}

// clampColumn keeps the cursor's column on one the table has.
func (m Model[T]) clampColumn() Model[T] {
	if !m.cells {
		m.col = 0
		return m
	}
	m.col = min(max(m.col, 0), max(len(m.cols)-1, 0))
	return m
}

// markCell paints the cursor's cell in st, padded to its column's width so
// the block covers the whole cell; the last column keeps its own end.
func (m Model[T]) markCell(cells []string, widths []int, st lipgloss.Style) []string {
	if !m.cells || m.col >= len(widths) {
		return cells
	}
	out := make([]string, max(len(cells), m.col+1))
	copy(out, cells)
	text := draw.Truncate(out[m.col], widths[m.col])
	if m.col < len(widths)-1 {
		text = draw.Fit(text, widths[m.col])
	}
	out[m.col] = st.Render(text)
	return out
}
