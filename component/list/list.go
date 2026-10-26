// Package list is a cursor list with section headings, a scroll indicator
// or pages, a / filter that can rank fuzzily, and an optional title and
// status line. Rows carry a caller payload; what a choice means is the caller's.
package list

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/component/paginator"
	"github.com/lucasassuncao/bezel/internal/fuzzy"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Row is one line, carrying a payload of the caller's type T. A Section row
// is a heading (or a blank spacer when its Label is empty) and cannot be
// selected. Mark is a glyph before the label; Style colours an unselected
// row, falling back to the theme's Dim. A Description shows muted under the
// label, and once one row has one every row takes two lines.
type Row[T any] struct {
	Label       string
	Description string
	Section     bool
	Mark        string
	Style       *lipgloss.Style
	Value       T
}

// Action is what a key did.
type Action int

const (
	None   Action = iota
	Chosen        // enter on a selectable row, filtering or not
)

// Keys are the bindings the list answers to.
type Keys struct {
	Up, Down, PageUp, PageDown, Top, Bottom, Enter, Filter, Esc key.Binding
}

// DefaultKeys are the arrows, pgup/pgdown, home/end, enter, / and esc.
func DefaultKeys() Keys {
	return Keys{
		Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("↑/↓", "move")),
		Down:     key.NewBinding(key.WithKeys("down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup/pgdn", "page")),
		PageDown: key.NewBinding(key.WithKeys("pgdown")),
		Top:      key.NewBinding(key.WithKeys("home"), key.WithHelp("home/end", "ends")),
		Bottom:   key.NewBinding(key.WithKeys("end")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "choose")),
		Filter:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Esc:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "stop filtering")),
	}
}

// Model is a value: every method returns the new list.
type Model[T any] struct {
	rows   []Row[T]
	win    layout.Window
	keys   Keys
	find   filterMode
	match  func(query, label string) bool // nil is Contains
	fuzzy  bool
	mo     motion
	height int // the lines the list takes, title and status included
	title  string
	status bool
	pages  bool
	wrap   bool
}

// Contains is the filter's default match: the label holds the query, in any case.
func Contains(query, label string) bool {
	return strings.Contains(strings.ToLower(label), strings.ToLower(query))
}

// Prefix is a match for the label starting with the query, in any case.
func Prefix(query, label string) bool {
	return strings.HasPrefix(strings.ToLower(label), strings.ToLower(query))
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
func New[T any](rows []Row[T], height int) Model[T] {
	m := Model[T]{rows: slices.Clone(rows), keys: DefaultKeys(), height: height}
	m.win = layout.Window{Cursor: firstSelectable(rows)}
	return m.fit()
}

// WithMatch filters with match, which gets the typed text and each label:
// Prefix, a match of the app's own. nil is Contains again. It ends WithFuzzy.
func (m Model[T]) WithMatch(match func(query, label string) bool) Model[T] {
	m.match, m.fuzzy = match, false
	return m.clampFilter()
}

// WithFuzzy filters by the typed letters in order, anywhere in the label,
// best match first, and marks the letters that matched.
func (m Model[T]) WithFuzzy() Model[T] {
	m.fuzzy = true
	return m.clampFilter()
}

// WithTitle draws title on the list's first line.
func (m Model[T]) WithTitle(title string) Model[T] {
	m.title = title
	return m.fit()
}

// WithStatusBar draws a line under the rows saying how many there are, or
// while filtering how many match.
func (m Model[T]) WithStatusBar(on bool) Model[T] {
	m.status = on
	return m.fit()
}

// WithPages shows the rows a page at a time with the page marked as dots
// under them, instead of scrolling.
func (m Model[T]) WithPages(on bool) Model[T] {
	m.pages = on
	return m.fit()
}

// WithWrap moves the cursor from the last row to the first, and back.
func (m Model[T]) WithWrap(on bool) Model[T] {
	m.wrap = on
	return m
}

// Legend is the keys the list answers to now: the filter's while it is open.
func (m Model[T]) Legend() []legend.Entry {
	k := m.keys
	if m.find.on {
		return legend.FromBindings([]key.Binding{k.Up, k.Down, k.Enter, k.Esc})
	}
	return legend.FromBindings([]key.Binding{k.Up, k.Down, k.PageUp, k.PageDown, k.Top, k.Bottom, k.Enter, k.Filter})
}

// itemLines is the lines a row takes: two once a row has a description.
func (m Model[T]) itemLines() int {
	for _, r := range m.rows {
		if r.Description != "" {
			return 2
		}
	}
	return 1
}

// body is the lines left for the rows under the title and over the status
// line and the pages.
func (m Model[T]) body() int {
	n := m.height
	if m.title != "" {
		n--
	}
	if m.status {
		n--
	}
	if m.pages {
		n--
	}
	return max(n, 0)
}

// fit sizes the windows to the rows that fit in the body.
func (m Model[T]) fit() Model[T] {
	if m.height <= 0 {
		m.win.Height, m.find.win.Height = m.height, 0
		return m.clampScroll().clampFilter()
	}
	m.win.Height = m.body() / m.itemLines()
	m.find.win.Height = max(m.body()-1, 0) / m.itemLines()
	return m.clampScroll().clampFilter()
}

// WithKeys rebinds the list.
func (m Model[T]) WithKeys(k Keys) Model[T] {
	m.keys = k
	return m
}

// Rows is a copy of every row, headings included.
func (m Model[T]) Rows() []Row[T] { return slices.Clone(m.rows) }

// Len is how many rows there are, headings included.
func (m Model[T]) Len() int { return len(m.rows) }

// IsFiltering reports whether the / prompt is open and taking the keys.
func (m Model[T]) IsFiltering() bool { return m.find.on }

// Filter is the text typed at the / prompt.
func (m Model[T]) Filter() string { return m.find.text }

// Scroll is where the list shows its rows from, for a pane's scrollbar; while
// filtering, where the matches show from.
func (m Model[T]) Scroll() layout.Scroll {
	if m.find.on {
		n := len(m.filtered())
		return m.find.win.Clamp(n).Scroll(n)
	}
	return m.win.Clamp(len(m.rows)).Scroll(len(m.rows))
}

// SetRows replaces the rows, keeping the cursor on the row with the same
// label when it survives (at the same index first), else on the first
// selectable one.
func (m Model[T]) SetRows(rows []Row[T]) Model[T] {
	before, cursor := m.rows, m.win.Cursor
	return m.setRows(rows).changed(before, cursor)
}

// setRows is SetRows without the motion.
func (m Model[T]) setRows(rows []Row[T]) Model[T] {
	prev, at := "", m.win.Cursor
	if r := m.Selected(); r != nil {
		prev = r.Label
	}
	m.rows = slices.Clone(rows)
	m = m.fit()
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

// SetHeight updates the lines the list takes and re-clamps the scroll.
func (m Model[T]) SetHeight(h int) Model[T] {
	m.height = h
	return m.fit()
}

// InsertRow puts r before row i, i kept within the rows, the cursor staying
// on the row it was on.
func (m Model[T]) InsertRow(i int, r Row[T]) Model[T] {
	i = min(max(i, 0), len(m.rows))
	return m.SetRows(slices.Insert(slices.Clone(m.rows), i, r))
}

// RemoveRow takes row i out; an index out of range changes nothing.
func (m Model[T]) RemoveRow(i int) Model[T] {
	if i < 0 || i >= len(m.rows) {
		return m
	}
	return m.SetRows(slices.Delete(slices.Clone(m.rows), i, i+1))
}

// SetRow replaces row i; an index out of range changes nothing.
func (m Model[T]) SetRow(i int, r Row[T]) Model[T] {
	if i < 0 || i >= len(m.rows) {
		return m
	}
	rows := slices.Clone(m.rows)
	rows[i] = r
	return m.SetRows(rows)
}

// Selected is the row under the cursor, or nil on a heading or an empty
// list. While filtering it is the row under the filter cursor.
func (m Model[T]) Selected() *Row[T] {
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
func (m Model[T]) Update(msg tea.KeyPressMsg) (Model[T], Action) {
	if m.find.on {
		return m.updateFilter(msg)
	}
	switch {
	case key.Matches(msg, m.keys.Filter):
		m.find = m.newFilter()
	case key.Matches(msg, m.keys.Up):
		return m.move(-1), None
	case key.Matches(msg, m.keys.Down):
		return m.move(1), None
	case key.Matches(msg, m.keys.PageUp):
		return m.jump(-max(m.win.Height, 1)), None
	case key.Matches(msg, m.keys.PageDown):
		return m.jump(max(m.win.Height, 1)), None
	case key.Matches(msg, m.keys.Top):
		return m.jump(-len(m.rows)), None
	case key.Matches(msg, m.keys.Bottom):
		return m.jump(len(m.rows)), None
	case key.Matches(msg, m.keys.Enter):
		if m.Selected() != nil {
			return m, Chosen
		}
	}
	return m, None
}

// updateFilter is the filter prompt: text edits the filter, arrows walk the
// matches, enter picks one and leaves, esc leaves with nothing.
func (m Model[T]) updateFilter(msg tea.KeyPressMsg) (Model[T], Action) {
	switch {
	case key.Matches(msg, m.keys.Esc):
		m.find = filterMode{}
	case key.Matches(msg, m.keys.Enter):
		return m.pickMatch()
	case msg.String() == "backspace" || msg.String() == "ctrl+h":
		if text := m.find.text; text != "" {
			_, size := utf8.DecodeLastRuneInString(text)
			m.find = m.newFilter()
			m.find.text = text[:len(text)-size]
		}
	case key.Matches(msg, m.keys.Up):
		m.find.win = m.find.win.Move(m.wrapped(-1, m.find.win.Cursor, len(m.filtered())), len(m.filtered()), nil)
	case key.Matches(msg, m.keys.Down):
		m.find.win = m.find.win.Move(m.wrapped(1, m.find.win.Cursor, len(m.filtered())), len(m.filtered()), nil)
	default:
		// Only arrows navigate while filtering, so letters stay typeable.
		if typed := msg.Key().Text; utf8.RuneCountInString(typed) == 1 {
			text := m.find.text + typed
			m.find = m.newFilter()
			m.find.text = text
		}
	}
	return m, None
}

// Click is a left click on line y of View: on a row it moves the cursor
// there, and on the row already under the cursor it chooses it, as enter
// would. A heading, the title or a line past the rows does nothing.
func (m Model[T]) Click(y int) (Model[T], Action) {
	if m.title != "" {
		y--
	}
	if m.height <= 0 || y < 0 {
		return m, None
	}
	at := y / m.itemLines()
	if m.find.on {
		idx := m.filtered()
		w := m.find.win.Clamp(len(idx))
		start, end, _ := w.Bounds(len(idx))
		if at >= end-start {
			return m, None
		}
		if start+at == w.Cursor {
			return m.pickMatch()
		}
		m.find.win.Cursor = start + at
		return m, None
	}
	start, end, _ := m.win.Bounds(len(m.rows))
	if m.pages && m.win.Height > 0 {
		start = m.win.Cursor / m.win.Height * m.win.Height
		end = min(start+m.win.Height, len(m.rows))
	}
	i := start + at
	switch {
	case i >= end || m.rows[i].Section:
		return m, None
	case i == m.win.Cursor:
		return m, Chosen
	}
	from := m.win.Cursor
	m.win.Cursor = i
	return m.start(float64(from)).clampScroll(), None
}

// pickMatch closes the filter on the match under its cursor, Chosen, or on
// nothing when nothing matched.
func (m Model[T]) pickMatch() (Model[T], Action) {
	idx, at := m.filtered(), m.find.win.Cursor
	m.find = filterMode{}
	if at >= len(idx) {
		return m, None
	}
	from := m.win.Cursor
	m.win.Cursor = idx[at]
	return m.start(float64(from)).clampScroll(), Chosen
}

// newFilter is the prompt opened afresh, its window the rows the body holds
// beside the prompt's line.
func (m Model[T]) newFilter() filterMode {
	f := newFilter(0)
	f.win.Height = max(m.body()-1, 0) / m.itemLines()
	if m.height <= 0 {
		f.win.Height = 0
	}
	return f
}

// wrapped is delta, or the step that wraps round, for a cursor at at over n
// rows when the list wraps and a step of one would go past an end.
func (m Model[T]) wrapped(delta, at, n int) int {
	switch {
	case !m.wrap || n == 0:
		return delta
	case delta > 0 && at >= n-1:
		return -n
	case delta < 0 && at <= 0:
		return n
	}
	return delta
}

// filtered is the indices into rows that match the filter, so a pick maps
// back to its own row even when labels repeat.
func (m Model[T]) filtered() []int {
	if m.fuzzy && m.find.text != "" {
		return m.ranked()
	}
	query, match := m.find.text, m.match
	if match == nil {
		// Contains, with the query lowered once rather than once per row.
		q := strings.ToLower(query)
		match = func(_, label string) bool { return strings.Contains(strings.ToLower(label), q) }
	}
	var out []int
	for i, r := range m.rows {
		if !r.Section && (query == "" || match(query, r.Label)) {
			out = append(out, i)
		}
	}
	return out
}

// ranked is the rows the fuzzy filter keeps, best match first.
func (m Model[T]) ranked() []int {
	labels := make([]string, len(m.rows))
	for i, r := range m.rows {
		if !r.Section { // a heading's empty label matches no query
			labels[i] = r.Label
		}
	}
	return fuzzy.Ranked(m.find.text, labels)
}

// move steps the cursor over headings, clamped at the ends, or round them
// when the list wraps.
func (m Model[T]) move(delta int) Model[T] {
	moved := m.jump(delta)
	if !m.wrap || moved.win.Cursor != m.win.Cursor || m.Selected() == nil {
		return moved
	}
	if delta > 0 {
		return m.jump(-len(m.rows))
	}
	return m.jump(len(m.rows))
}

// jump moves the cursor delta rows over headings, clamped at the ends.
func (m Model[T]) jump(delta int) Model[T] {
	from := m.win.Cursor
	m.win = m.win.Move(delta, len(m.rows), func(i int) bool { return m.rows[i].Section })
	if m.win.Cursor != from {
		m = m.start(float64(from))
	}
	return m
}

func (m Model[T]) clampScroll() Model[T] {
	m.win = m.win.Clamp(len(m.rows))
	return m
}

// clampFilter keeps the filter's window on its matches after the rows changed.
func (m Model[T]) clampFilter() Model[T] {
	if m.find.on {
		m.find.win = m.find.win.Clamp(len(m.filtered()))
	}
	return m
}

func firstSelectable[T any](rows []Row[T]) int {
	for i, r := range rows {
		if !r.Section {
			return i
		}
	}
	return 0
}

// View renders the title, the visible rows or the filter prompt with its
// matches, then the pages and the status line.
func (m Model[T]) View(th theme.Resolved) string {
	if m.height <= 0 {
		return ""
	}
	var body string
	if m.find.on {
		body = m.viewFilter(&th)
	} else {
		body = m.viewRows(&th)
	}
	if m.title == "" && !m.status && !m.pages {
		return body
	}
	var lines []string
	if m.title != "" {
		lines = append(lines, th.Accent.Bold(true).Render(m.title))
	}
	rows := strings.Split(body, "\n")
	for len(rows) < m.body() {
		rows = append(rows, "")
	}
	lines = append(lines, rows...)
	if m.pages {
		lines = append(lines, m.pager(th))
	}
	if m.status {
		lines = append(lines, th.Muted.Render(m.statusText()))
	}
	return strings.Join(lines, "\n")
}

// pager is the dots under a paged list, blank while filtering.
func (m Model[T]) pager(th theme.Resolved) string {
	if m.find.on || m.win.Height <= 0 {
		return ""
	}
	p := paginator.New(m.win.Height).SetItems(len(m.rows)).SetPage(m.win.Cursor / m.win.Height)
	return "  " + p.View(th)
}

// statusText counts the rows, or while filtering the ones that match.
func (m Model[T]) statusText() string {
	n := 0
	for _, r := range m.rows {
		if !r.Section {
			n++
		}
	}
	unit := "items"
	if n == 1 {
		unit = "item"
	}
	switch {
	case !m.find.on || m.find.text == "":
		return fmt.Sprintf("  %d %s", n, unit)
	case len(m.filtered()) == 0:
		return "  nothing matched"
	}
	return fmt.Sprintf("  %d of %d", len(m.filtered()), n)
}

// viewRows is the rows in the window, or on the cursor's page.
func (m Model[T]) viewRows(th *theme.Resolved) string {
	if m.win.Height <= 0 {
		return ""
	}
	start, end, below := m.win.Bounds(len(m.rows))
	if m.pages {
		start = m.win.Cursor / m.win.Height * m.win.Height
		end, below = min(start+m.win.Height, len(m.rows)), 0
	}
	cursor := m.cursorRow()

	two := m.itemLines() == 2
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		if g, ok := m.ghostLine(i, th); ok {
			lines = append(lines, g)
		}
		r := m.rows[i]
		switch {
		case r.Section:
			heading := th.Section.Render(r.Label)
			if two {
				heading += "\n"
			}
			lines = append(lines, heading)
		case i != cursor && m.arriving(r, th) != nil:
			r.Style = m.arriving(r, th)
			lines = append(lines, renderRow(r, false, th, nil, two))
		default:
			lines = append(lines, renderRow(r, i == cursor, th, nil, two))
		}
	}
	if g, ok := m.ghostLine(end, th); ok && below == 0 {
		lines = append(lines, g)
	}
	room := max(end-start, m.win.Height-boolInt(below > 0)) // a ghost never pushes the list past its height
	lines = lines[:min(len(lines), room)]
	if below > 0 {
		lines = append(lines, moreRow(below, th))
	}
	return strings.Join(lines, "\n")
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// viewFilter is the matches in the filter's window, padded to it, then the prompt.
func (m Model[T]) viewFilter(th *theme.Resolved) string {
	idx := m.filtered()
	w := m.find.win.Clamp(len(idx))
	start, end, below := w.Bounds(len(idx))

	lines := make([]string, 0, m.win.Height)
	two := m.itemLines() == 2
	for i := start; i < end; i++ {
		r := m.rows[idx[i]]
		var hits []int
		if m.fuzzy {
			_, hits, _ = fuzzy.Match(m.find.text, r.Label)
		}
		lines = append(lines, renderRow(r, i == w.Cursor, th, hits, two))
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
func moreRow(n int, th *theme.Resolved) string { return th.Dim.Render(fmt.Sprintf("  ↓ %d more", n)) }

// renderRow is a row in its style, the runes of its label at hits
// underlined, and with two its description, or a blank, under it.
func renderRow[T any](r Row[T], selected bool, th *theme.Resolved, hits []int, two bool) string {
	style, lead := th.Dim, "  "
	switch {
	case selected:
		style, lead = th.Cursor, "▶ "
	case r.Style != nil:
		style = *r.Style
	}
	if r.Mark != "" {
		lead += r.Mark + "  "
	}
	row := style.Render(lead) + underlined(r.Label, hits, style)
	if two {
		indent := strings.Repeat(" ", utf8.RuneCountInString(lead))
		row += "\n" + th.Muted.Render(indent+r.Description)
	}
	return row
}

// underlined is label in style, the runes at hits underlined and bold.
func underlined(label string, hits []int, style lipgloss.Style) string {
	if len(hits) == 0 {
		return style.Render(label)
	}
	hit := style.Underline(true).Bold(true)
	var b strings.Builder
	runes, k, run := []rune(label), 0, 0
	for i := range runes {
		on := k < len(hits) && hits[k] == i
		if on {
			k++
		}
		next := k < len(hits) && hits[k] == i+1
		if i+1 == len(runes) || on != next {
			st := style
			if on {
				st = hit
			}
			b.WriteString(st.Render(string(runes[run : i+1])))
			run = i + 1
		}
	}
	return b.String()
}
