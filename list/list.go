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
	rows   []Row
	cursor int
	height int
	offset int
	keys   Keys

	filter    string
	filtering bool
	fCursor   int
	fOffset   int
}

// New builds a list height rows tall with the cursor on the first
// selectable row.
func New(rows []Row, height int) Model {
	m := Model{rows: slices.Clone(rows), height: height, keys: DefaultKeys()}
	m.cursor = firstSelectable(rows)
	return m
}

func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

func (m Model) Rows() []Row       { return slices.Clone(m.rows) }
func (m Model) Len() int          { return len(m.rows) }
func (m Model) IsFiltering() bool { return m.filtering }
func (m Model) Filter() string    { return m.filter }

// SetRows replaces the rows, keeping the cursor on the row with the same
// label when it survives (at the same index first), else on the first
// selectable one.
func (m Model) SetRows(rows []Row) Model {
	prev, at := "", m.cursor
	if r := m.Selected(); r != nil {
		prev = r.Label
	}
	m.rows = slices.Clone(rows)
	if at < len(m.rows) && !m.rows[at].Section && m.rows[at].Label == prev {
		return m.clampScroll().clampFilterCursor()
	}
	m.cursor = firstSelectable(m.rows)
	for i, r := range m.rows {
		if !r.Section && r.Label == prev {
			m.cursor = i
			break
		}
	}
	return m.clampScroll().clampFilterCursor()
}

// SetHeight updates the visible row count and re-clamps the scroll.
func (m Model) SetHeight(h int) Model {
	m.height = h
	return m.clampScroll()
}

// Selected is the row under the cursor, or nil on a heading or an empty
// list. While filtering it is the row under the filter cursor.
func (m Model) Selected() *Row {
	if m.filtering {
		idx := m.filtered()
		if m.fCursor >= len(idx) {
			return nil
		}
		r := m.rows[idx[m.fCursor]]
		return &r
	}
	if m.cursor >= len(m.rows) || m.rows[m.cursor].Section {
		return nil
	}
	r := m.rows[m.cursor]
	return &r
}

// Update handles one key. Keys the list does not know return None and
// leave it unchanged, so the caller can layer its own on top.
func (m Model) Update(msg tea.KeyPressMsg) (Model, Action) {
	if m.filtering {
		return m.updateFilter(msg)
	}
	switch {
	case key.Matches(msg, m.keys.Filter):
		m.filtering, m.filter, m.fCursor, m.fOffset = true, "", 0, 0
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
		m.filtering, m.filter, m.fCursor, m.fOffset = false, "", 0, 0
	case key.Matches(msg, m.keys.Enter):
		idx := m.filtered()
		m.filtering = false
		if m.fCursor >= len(idx) {
			return m, None
		}
		m.cursor = idx[m.fCursor]
		return m.clampScroll(), Chosen
	case msg.String() == "backspace" || msg.String() == "ctrl+h":
		if m.filter != "" {
			_, size := utf8.DecodeLastRuneInString(m.filter)
			m.filter = m.filter[:len(m.filter)-size]
			m.fCursor, m.fOffset = 0, 0
		}
	case key.Matches(msg, m.keys.Up):
		return m.moveFilter(-1), None
	case key.Matches(msg, m.keys.Down):
		return m.moveFilter(1), None
	default:
		// Only arrows navigate while filtering, so letters stay typeable.
		if text := msg.Key().Text; utf8.RuneCountInString(text) == 1 {
			m.filter += text
			m.fCursor, m.fOffset = 0, 0
		}
	}
	return m, None
}

// filtered is the indices into rows that match the filter, so a pick maps
// back to its own row even when labels repeat.
func (m Model) filtered() []int {
	f := strings.ToLower(m.filter)
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
	for i := m.cursor + delta; i >= 0 && i < len(m.rows); i += delta {
		if !m.rows[i].Section {
			m.cursor = i
			break
		}
	}
	return m.clampScroll()
}

func (m Model) moveFilter(delta int) Model {
	next := m.fCursor + delta
	if next < 0 || next >= len(m.filtered()) {
		return m
	}
	m.fCursor = next
	return m.clampFilterScroll()
}

func (m Model) clampScroll() Model {
	if m.height <= 0 {
		return m
	}
	m.offset = layout.ClampScroll(m.cursor, m.offset, m.height)
	// The last row becomes the "↓ N more" indicator when rows overflow below,
	// so a cursor landing there needs one more line of offset to stay visible.
	if m.hasMore() && m.cursor >= m.offset+m.height-1 {
		m.offset = m.cursor - m.height + 2
	}
	return m
}

func (m Model) clampFilterScroll() Model {
	if visible := m.height - 1; visible > 0 {
		m.fOffset = layout.ClampScroll(m.fCursor, m.fOffset, visible)
	}
	return m
}

// clampFilterCursor keeps the filter cursor in range after the rows changed.
func (m Model) clampFilterCursor() Model {
	if !m.filtering {
		return m
	}
	n := len(m.filtered())
	if m.fCursor >= n {
		m.fCursor = max(n-1, 0)
	}
	m.fOffset = 0
	return m.clampFilterScroll()
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
	if m.filtering {
		return m.viewFilter(th)
	}
	if m.height <= 0 {
		return ""
	}
	visible := m.height
	more := m.hasMore()
	if more {
		visible--
	}
	end := min(m.offset+visible, len(m.rows))

	var sb strings.Builder
	for i := m.offset; i < end; i++ {
		if i > m.offset {
			sb.WriteByte('\n')
		}
		r := m.rows[i]
		if r.Section {
			sb.WriteString(th.Section.Render(r.Label))
		} else {
			sb.WriteString(renderRow(r, i == m.cursor, th))
		}
	}
	if more {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(th.Dim.Render(fmt.Sprintf("  ↓ %d more", len(m.rows)-end)))
	}
	return sb.String()
}

// hasMore reports whether rows overflow below the window. A one-row window
// never trades its only row for the indicator.
func (m Model) hasMore() bool {
	return m.height > 1 && m.offset+m.height < len(m.rows)
}

func (m Model) viewFilter(th theme.Resolved) string {
	idx := m.filtered()
	visible := max(m.height-1, 0)
	end := min(m.fOffset+visible, len(idx))

	lines := make([]string, 0, m.height)
	for i := m.fOffset; i < end; i++ {
		lines = append(lines, renderRow(m.rows[idx[i]], i == m.fCursor, th))
	}
	for len(lines) < visible {
		lines = append(lines, "")
	}
	lines = append(lines, th.Cursor.Render("/"+m.filter+"▋"))
	return strings.Join(lines, "\n")
}

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
