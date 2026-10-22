package viewport

import (
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/edit"
)

// span is the cells [start, end) one match covers on one line.
type span struct{ start, end, match int }

// highlights are the matches kept by line for View, and the line each one
// starts on, in order.
type highlights struct {
	byLine map[int][]span
	first  []int
}

// toSpans turns byte ranges into plain into each match's cells on every line
// it touches, walking plain's graphemes once. A range out of order is skipped.
func toSpans(plain string, matches [][]int) highlights {
	h := highlights{byLine: map[int][]span{}}
	gs := edit.Graphemes(plain)
	gi, b, line, col := 0, 0, 0, 0
	step := func() {
		if gs[gi] == "\n" {
			line, col = line+1, 0
		} else {
			col += edit.Cells(gs[gi])
		}
		b += len(gs[gi])
		gi++
	}
	for _, mt := range matches {
		if len(mt) < 2 || mt[0] < b || mt[1] < mt[0] {
			continue
		}
		for gi < len(gs) && b < mt[0] {
			step()
		}
		k := len(h.first)
		h.first = append(h.first, line)
		start := col
		for gi < len(gs) && b < mt[1] {
			nl := gs[gi] == "\n"
			if nl && col > start {
				h.byLine[line] = append(h.byLine[line], span{start, col, k})
			}
			step()
			if nl {
				start = 0
			}
		}
		if col > start {
			h.byLine[line] = append(h.byLine[line], span{start, col, k})
		}
	}
	return h
}

// SetHighlights paints matches: byte ranges into the content's plain text,
// escape sequences stripped, in order and not overlapping, as
// regexp.FindAllStringIndex returns them. It shows the first at or below the top.
func (m Model) SetHighlights(matches [][]int) Model {
	m.hl, m.cur = toSpans(ansi.Strip(m.Content()), matches), -1
	if len(m.hl.first) == 0 {
		return m
	}
	m.cur = max(m.nearest(), 0)
	return m.show()
}

// ClearHighlights removes every match.
func (m Model) ClearHighlights() Model {
	m.hl, m.cur = highlights{}, -1
	return m
}

// NextHighlight makes the next match current, the first after the last,
// and brings it into view.
func (m Model) NextHighlight() Model {
	if n := len(m.hl.first); n > 0 {
		m.cur = (m.cur + 1) % n
		return m.show()
	}
	return m
}

// PrevHighlight makes the previous match current, the last before the
// first, and brings it into view.
func (m Model) PrevHighlight() Model {
	if n := len(m.hl.first); n > 0 {
		m.cur = (max(m.cur, 0) - 1 + n) % n
		return m.show()
	}
	return m
}

// Highlight is the current match, from 0; -1 for none.
func (m Model) Highlight() int {
	if len(m.hl.first) == 0 {
		return -1
	}
	return m.cur
}

// Highlights is how many matches there are.
func (m Model) Highlights() int { return len(m.hl.first) }

// nearest is the first match starting at or below the top line, -1 for none.
func (m Model) nearest() int {
	top, _ := m.idx.at(m.YOffset())
	for i, line := range m.hl.first {
		if line >= top {
			return i
		}
	}
	return -1
}

// renearest moves the current match after a scroll, without scrolling.
func (m Model) renearest() Model {
	if len(m.hl.first) > 0 {
		m.cur = m.nearest()
	}
	return m
}

// show brings the current match's first cells into view.
func (m Model) show() Model {
	line := m.hl.first[m.cur]
	for _, s := range m.hl.byLine[line] {
		if s.match == m.cur {
			return m.EnsureVisible(line, s.start, s.end)
		}
	}
	return m.EnsureVisible(line, 0, 0)
}
