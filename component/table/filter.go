package table

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/fuzzy"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
)

// Contains is the filter's default match: the row's text holds the query,
// in any case.
func Contains(query, text string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(query))
}

// filterMode is the / prompt while it is open: the text typed, the rows it
// matches, found once per change, and a window over them one row shorter
// than the table's, for the prompt.
type filterMode struct {
	on   bool
	text string
	idx  []int
	win  layout.Window
}

// WithMatch filters with match, which gets the typed text and each row's
// cells joined by spaces, colours stripped. nil is Contains again. It ends
// WithFuzzy.
func (m Model[T]) WithMatch(match func(query, text string) bool) Model[T] {
	m.match, m.fuzzy = match, false
	return m.refilter()
}

// WithFuzzy filters by the typed letters in order, anywhere in the row,
// best match first.
func (m Model[T]) WithFuzzy(on bool) Model[T] {
	m.fuzzy = on
	return m.refilter()
}

// Filtering reports whether the / prompt is open and taking the keys.
func (m Model[T]) Filtering() bool { return m.find.on }

// Filter is the text typed at the / prompt.
func (m Model[T]) Filter() string { return m.find.text }

// updateFilter is the filter prompt: text edits the filter, arrows walk the
// matches, enter picks one and leaves, esc leaves with nothing.
func (m Model[T]) updateFilter(msg tea.KeyPressMsg) (Model[T], bool) {
	switch {
	case key.Matches(msg, m.keys.Esc):
		m.find = filterMode{}
	case key.Matches(msg, m.keys.Enter):
		return m.pickMatch()
	case key.Matches(msg, m.keys.Erase):
		if text := m.find.text; text != "" {
			_, size := utf8.DecodeLastRuneInString(text)
			m.find = m.filterFor(text[:len(text)-size])
		}
	case utf8.RuneCountInString(msg.Key().Text) == 1:
		// Text before the moves, so vim's j and k stay typeable while filtering.
		m.find = m.filterFor(m.find.text + msg.Key().Text)
	case key.Matches(msg, m.keys.Up):
		m.find.win = m.find.win.Move(-1, len(m.find.idx), nil)
	case key.Matches(msg, m.keys.Down):
		m.find.win = m.find.win.Move(1, len(m.find.idx), nil)
	}
	return m, false
}

// pickMatch closes the filter on the match under its cursor, chosen, or on
// nothing when nothing matched.
func (m Model[T]) pickMatch() (Model[T], bool) {
	idx, at := m.find.idx, m.find.win.Cursor
	m.find = filterMode{}
	if at >= len(idx) {
		return m, false
	}
	m.win.Cursor = idx[at]
	return m.clamp(), true
}

// filterFor is the prompt open on text, its matches found here once.
func (m Model[T]) filterFor(text string) filterMode {
	return filterMode{on: true, text: text, idx: m.filtered(text), win: layout.Window{Height: m.filterRows()}}
}

// filterRows is the rows shown under the header and over the prompt.
func (m Model[T]) filterRows() int { return max(m.height-2, 0) }

// filtered is the indices of the rows that match query, so a pick maps
// back to its own row even when rows repeat.
func (m Model[T]) filtered(query string) []int {
	texts := make([]string, len(m.items))
	for i, it := range m.items {
		texts[i] = ansi.Strip(strings.Join(it.Cells, " "))
	}
	if m.fuzzy && query != "" {
		return fuzzy.Ranked(query, texts)
	}
	match := m.match
	if match == nil {
		q := strings.ToLower(query)
		match = func(_, text string) bool { return strings.Contains(strings.ToLower(text), q) }
	}
	var out []int
	for i, t := range texts {
		if query == "" || match(query, t) {
			out = append(out, i)
		}
	}
	return out
}

// refilter matches the open filter again after the rows or the match changed.
func (m Model[T]) refilter() Model[T] {
	if m.find.on {
		m.find.idx = m.filtered(m.find.text)
		m.find.win.Height = m.filterRows()
		m.find.win = m.find.win.Clamp(len(m.find.idx))
	}
	return m
}
