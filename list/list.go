// Package list is a cursor list with section headings, a scroll indicator
// and a / filter. Rows carry a caller payload; what a choice means is the
// caller's.
package list

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

// Row is one line. A Section row is a heading (or a blank spacer when its
// Label is empty) and cannot be selected. Mark is a glyph before the label;
// Style colours an unselected row, falling back to the theme's Dim.
type Row struct {
	Label   string
	Section bool
	Mark    string
	Style   *lipgloss.Style
	Value   any
}

// Action is what a key did.
type Action int

const (
	None   Action = iota
	Chosen        // enter on a selectable row, filtering or not
)

// Keys are the bindings the list answers to.
type Keys struct {
	Up, Down, Enter, Filter, Esc key.Binding
}

// DefaultKeys are the arrows, enter, / and esc.
func DefaultKeys() Keys {
	return Keys{
		Up:     key.NewBinding(key.WithKeys("up")),
		Down:   key.NewBinding(key.WithKeys("down")),
		Enter:  key.NewBinding(key.WithKeys("enter")),
		Filter: key.NewBinding(key.WithKeys("/")),
		Esc:    key.NewBinding(key.WithKeys("esc")),
	}
}

// Model is a value: every method returns the new list.
type Model struct {
	rows []Row
	win  layout.Window
	keys Keys
	find filterMode
}

// filterMode is the / prompt while it is open: the text typed, and a window
// over the rows it matches one row shorter than the list, for the prompt.
type filterMode struct {
	on   bool
	text string
	win  layout.Window
}

// newFilter is the prompt opened afresh under a list height rows tall.
func newFilter(height int) filterMode {
	return filterMode{on: true, win: layout.Window{Height: max(height-1, 0)}}
}

// New builds a list height rows tall with the cursor on the first
// selectable row.
func New(rows []Row, height int) Model {
	m := Model{rows: slices.Clone(rows), keys: DefaultKeys()}
	m.win = layout.Window{Cursor: firstSelectable(rows), Height: height}
	return m
}

// WithKeys rebinds the list.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// Rows is a copy of every row, headings included.
func (m Model) Rows() []Row { return slices.Clone(m.rows) }

// Len is how many rows there are, headings included.
func (m Model) Len() int { return len(m.rows) }

// IsFiltering reports whether the / prompt is open and taking the keys.
func (m Model) IsFiltering() bool { return m.find.on }

// Filter is the text typed at the / prompt.
func (m Model) Filter() string { return m.find.text }

// SetRows replaces the rows, keeping the cursor on the row with the same
// label when it survives (at the same index first), else on the first
// selectable one.
func (m Model) SetRows(rows []Row) Model {
	prev, at := "", m.win.Cursor
	if r := m.Selected(); r != nil {
		prev = r.Label
	}
	m.rows = slices.Clone(rows)
	if at < len(m.rows) && !m.rows[at].Section && m.rows[at].Label == prev {
		return m.clampScroll().clampFilter()
	}
	m.win.Cursor = firstSelectable(m.rows)
	for i, r := range m.rows {
		if !r.Section && r.Label == prev {
			m.win.Cursor = i
			break
		}
	}
	return m.clampScroll().clampFilter()
}

// SetHeight updates the visible row count and re-clamps the scroll.
func (m Model) SetHeight(h int) Model {
	m.win.Height = h
	m.find.win.Height = max(h-1, 0)
	return m.clampScroll().clampFilter()
}

// Selected is the row under the cursor, or nil on a heading or an empty
// list. While filtering it is the row under the filter cursor.
func (m Model) Selected() *Row {
	if m.find.on {
		idx := m.filtered()
		if m.find.win.Cursor >= len(idx) {
			return nil
		}
		r := m.rows[idx[m.find.win.Cursor]]
		return &r
	}
	if m.win.Cursor >= len(m.rows) || m.rows[m.win.Cursor].Section {
		return nil
	}
	r := m.rows[m.win.Cursor]
	return &r
}

// Update handles one key. Keys the list does not know return None and
// leave it unchanged, so the caller can layer its own on top.
func (m Model) Update(msg tea.KeyPressMsg) (Model, Action) {
	if m.find.on {
		return m.updateFilter(msg)
	}
	switch {
	case key.Matches(msg, m.keys.Filter):
		m.find = newFilter(m.win.Height)
	case key.Matches(msg, m.keys.Up):
		return m.move(-1), None
	case key.Matches(msg, m.keys.Down):
		return m.move(1), None
	case key.Matches(msg, m.keys.Enter):
		if m.Selected() != nil {
			return m, Chosen
		}
	}
	return m, None
}

// updateFilter is the filter prompt: text edits the filter, arrows walk the
// matches, enter picks one and leaves, esc leaves with nothing.
func (m Model) updateFilter(msg tea.KeyPressMsg) (Model, Action) {
	switch {
	case key.Matches(msg, m.keys.Esc):
		m.find = filterMode{}
	case key.Matches(msg, m.keys.Enter):
		idx, at := m.filtered(), m.find.win.Cursor
		m.find = filterMode{}
		if at >= len(idx) {
			return m, None
		}
		m.win.Cursor = idx[at]
		return m.clampScroll(), Chosen
	case msg.String() == "backspace" || msg.String() == "ctrl+h":
		if text := m.find.text; text != "" {
			_, size := utf8.DecodeLastRuneInString(text)
			m.find = newFilter(m.win.Height)
			m.find.text = text[:len(text)-size]
		}
	case key.Matches(msg, m.keys.Up):
		m.find.win = m.find.win.Move(-1, len(m.filtered()), nil)
	case key.Matches(msg, m.keys.Down):
		m.find.win = m.find.win.Move(1, len(m.filtered()), nil)
	default:
		// Only arrows navigate while filtering, so letters stay typeable.
		if typed := msg.Key().Text; utf8.RuneCountInString(typed) == 1 {
			text := m.find.text + typed
			m.find = newFilter(m.win.Height)
			m.find.text = text
		}
	}
	return m, None
}

// filtered is the indices into rows that match the filter, so a pick maps
// back to its own row even when labels repeat.
func (m Model) filtered() []int {
	f := strings.ToLower(m.find.text)
	var out []int
	for i, r := range m.rows {
		if !r.Section && (f == "" || strings.Contains(strings.ToLower(r.Label), f)) {
			out = append(out, i)
		}
	}
	return out
}

// move steps the cursor over headings, clamped at the ends.
func (m Model) move(delta int) Model {
	m.win = m.win.Move(delta, len(m.rows), func(i int) bool { return m.rows[i].Section })
	return m
}

func (m Model) clampScroll() Model {
	m.win = m.win.Clamp(len(m.rows))
	return m
}

// clampFilter keeps the filter's window on its matches after the rows changed.
func (m Model) clampFilter() Model {
	if m.find.on {
		m.find.win = m.find.win.Clamp(len(m.filtered()))
	}
	return m
}

func firstSelectable(rows []Row) int {
	for i, r := range rows {
		if !r.Section {
			return i
		}
	}
	return 0
}

// View renders the visible rows, or the filter prompt with its matches.
func (m Model) View(th theme.Resolved) string {
	if m.find.on {
		return m.viewFilter(th)
	}
	if m.win.Height <= 0 {
		return ""
	}
	start, end, below := m.win.Bounds(len(m.rows))

	var sb strings.Builder
	for i := start; i < end; i++ {
		if i > start {
			sb.WriteByte('\n')
		}
		r := m.rows[i]
		if r.Section {
			sb.WriteString(th.Section.Render(r.Label))
		} else {
			sb.WriteString(renderRow(r, i == m.win.Cursor, th))
		}
	}
	if below > 0 {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(moreRow(below, th))
	}
	return sb.String()
}

// viewFilter is the matches in the filter's window, padded to it, then the prompt.
func (m Model) viewFilter(th theme.Resolved) string {
	idx := m.filtered()
	w := m.find.win.Clamp(len(idx))
	start, end, below := w.Bounds(len(idx))

	lines := make([]string, 0, m.win.Height)
	for i := start; i < end; i++ {
		lines = append(lines, renderRow(m.rows[idx[i]], i == w.Cursor, th))
	}
	if below > 0 {
		lines = append(lines, moreRow(below, th))
	}
	for len(lines) < w.Height {
		lines = append(lines, "")
	}
	lines = append(lines, th.Cursor.Render("/"+m.find.text+"▋"))
	return strings.Join(lines, "\n")
}

// moreRow stands in for the last row while n rows are hidden below it.
func moreRow(n int, th theme.Resolved) string { return th.Dim.Render(fmt.Sprintf("  ↓ %d more", n)) }

func renderRow(r Row, selected bool, th theme.Resolved) string {
	text := r.Label
	if r.Mark != "" {
		text = r.Mark + "  " + r.Label
	}
	if selected {
		return th.Cursor.Render("▶ " + text)
	}
	if r.Style != nil {
		return r.Style.Render("  " + text)
	}
	return th.Dim.Render("  " + text)
}
