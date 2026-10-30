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

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/paginator"
	"github.com/lucasassuncao/bezel/draw"
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

// Keys are the bindings the list answers to.
type Keys struct {
	Up, Down, PageUp, PageDown, Top, Bottom, Enter, Filter, Esc key.Binding
	Erase                                                       key.Binding // the filter's last character
}

// DefaultKeys are the arrows, pgup/pgdown, home/end, enter, / and esc.
func DefaultKeys() Keys {
	return Keys{
		Up:       key.New("up").WithHelp("↑/↓", "move"),
		Down:     key.New("down"),
		PageUp:   key.New("pgup").WithHelp("pgup/pgdn", "page"),
		PageDown: key.New("pgdown"),
		Top:      key.New("home").WithHelp("home/end", "ends"),
		Bottom:   key.New("end"),
		Enter:    key.New("enter").WithHelp("enter", "choose"),
		Filter:   key.New("/").WithHelp("/", "filter"),
		Esc:      key.New("esc").WithHelp("esc", "stop filtering"),
		Erase:    key.New("backspace", "ctrl+h"),
	}
}

// VimKeys are DefaultKeys with vim's letters added: k/j to move, g/G to the
// ends and ctrl+b/f to page. The filter prompt still takes them as text.
func VimKeys() Keys {
	k := DefaultKeys()
	k.Up, k.Down = key.New("up", "k").WithHelp("↑/↓ k/j", "move"), key.New("down", "j")
	k.Top, k.Bottom = key.New("home", "g").WithHelp("g/G", "ends"), key.New("end", "G")
	k.PageUp, k.PageDown = key.New("pgup", "ctrl+b").WithHelp("^b/^f", "page"), key.New("pgdown", "ctrl+f")
	return k
}

var _ bezel.Model[Model[int]] = Model[int]{}

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
	edit   func(Styles) Styles
}

// Styles are how a list paints its parts.
type Styles struct {
	Title       lipgloss.Style
	Section     lipgloss.Style // a heading row
	Row         lipgloss.Style // a row off the cursor without a Style of its own
	Cursor      lipgloss.Style // the row under the cursor, and the filter prompt
	Description lipgloss.Style // the second line of a two-line row
	More        lipgloss.Style // the "↓ N more" row
	Status      lipgloss.Style
	Arrive      lipgloss.Style // its foreground is where a new row fades from
	Gone        lipgloss.Style // its foreground is where a removed row fades to
}

// DefaultStyles draw the cursor row in the cursor role over dim rows, the
// title in bold accent and the new and removed rows fading through the
// accent and the border.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{
		Title: th.Accent.Bold(true), Section: th.Section, Row: th.Dim, Cursor: th.Cursor,
		Description: th.Muted, More: th.Dim, Status: th.Muted,
		Arrive: lipgloss.NewStyle().Foreground(lipgloss.Color(th.Colors.Accent)),
		Gone:   lipgloss.NewStyle().Foreground(lipgloss.Color(th.Colors.Border)),
	}
}

// Contains is the filter's default match: the label holds the query, in any case.
func Contains(query, label string) bool {
	return strings.Contains(strings.ToLower(label), strings.ToLower(query))
}

// Prefix is a match for the label starting with the query, in any case.
func Prefix(query, label string) bool {
	return strings.HasPrefix(strings.ToLower(label), strings.ToLower(query))
}

// filterMode is the / prompt while it is open: the text typed, the rows it
// matches, found once per change, and a window over them one row shorter
// than the list, for the prompt.
type filterMode struct {
	on   bool
	text string
	idx  []int
	win  layout.Window
}

// New builds a list with the cursor on the first selectable row; SetHeight
// gives it the rows its pane has.
func New[T any](rows []Row[T]) Model[T] {
	m := Model[T]{rows: slices.Clone(rows), keys: DefaultKeys()}
	m.win = layout.Window{Cursor: firstSelectable(rows)}
	return m.fit()
}

// WithMatch filters with match, which gets the typed text and each label:
// Prefix, a match of the app's own. nil is Contains again. It ends WithFuzzy.
func (m Model[T]) WithMatch(match func(query, label string) bool) Model[T] {
	m.match, m.fuzzy = match, false
	return m.refilter()
}

// WithFuzzy filters by the typed letters in order, anywhere in the label,
// best match first, and marks the letters that matched.
func (m Model[T]) WithFuzzy(on bool) Model[T] {
	m.fuzzy = on
	return m.refilter()
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
	m.win.Height = m.height
	if m.height > 0 {
		m.win.Height = m.body() / m.itemLines()
	}
	m.find.win.Height = m.filterRows()
	return m.clampScroll().clampFilter()
}

// filterRows is the rows the body holds beside the filter prompt's line.
func (m Model[T]) filterRows() int {
	if m.height <= 0 {
		return 0
	}
	return max(m.body()-1, 0) / m.itemLines()
}

// WithKeys rebinds the list.
func (m Model[T]) WithKeys(k Keys) Model[T] {
	m.keys = k
	return m
}

// WithStyles changes the styles derived from the theme through edit, under
// whatever theme View gets. It replaces an earlier edit.
func (m Model[T]) WithStyles(edit func(Styles) Styles) Model[T] {
	m.edit = edit
	return m
}

// styles is DefaultStyles of th with the app's edit applied.
func (m Model[T]) styles(th theme.Resolved) Styles {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	return st
}

// Rows is a copy of every row, headings included.
func (m Model[T]) Rows() []Row[T] { return slices.Clone(m.rows) }

// Len is how many rows there are, headings included.
func (m Model[T]) Len() int { return len(m.rows) }

// Filtering reports whether the / prompt is open and taking the keys.
func (m Model[T]) Filtering() bool { return m.find.on }

// Filter is the text typed at the / prompt.
func (m Model[T]) Filter() string { return m.find.text }

// Scroll is where the list shows its rows from, for a pane's scrollbar; while
// filtering, where the matches show from.
func (m Model[T]) Scroll() layout.Scroll {
	if m.find.on {
		n := len(m.find.idx)
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
	if r, ok := m.Selected(); ok {
		prev = r.Label
	}
	m.rows = slices.Clone(rows)
	m = m.fit()
	if at < len(m.rows) && !m.rows[at].Section && m.rows[at].Label == prev {
		return m.clampScroll().refilter()
	}
	m.win.Cursor = firstSelectable(m.rows)
	for i, r := range m.rows {
		if !r.Section && r.Label == prev {
			m.win.Cursor = i
			break
		}
	}
	return m.clampScroll().refilter()
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

// Selected is a copy of the row under the cursor, false on a heading or an
// empty list. While filtering it is the row under the filter cursor.
func (m Model[T]) Selected() (Row[T], bool) {
	if m.find.on {
		idx := m.find.idx
		if m.find.win.Cursor >= len(idx) {
			return Row[T]{}, false
		}
		return m.rows[idx[m.find.win.Cursor]], true
	}
	if m.win.Cursor >= len(m.rows) || m.rows[m.win.Cursor].Section {
		return Row[T]{}, false
	}
	return m.rows[m.win.Cursor], true
}

// picked reports whether a row is under the cursor.
func (m Model[T]) picked() bool {
	_, ok := m.Selected()
	return ok
}

// Update takes the list's keys, a bezel.ClickMsg on a row and its own
// frames. Choosing a row reports a [Chosen]; a move with motion returns the
// list's Tick. Anything else comes back unhandled and unchanged.
func (m Model[T]) Update(msg tea.Msg) (Model[T], bezel.Result) {
	var chosen, handled bool
	before := m.mo.run.ID
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		handled = m.handles(msg)
		m, chosen = m.press(msg)
	case bezel.ClickMsg:
		m, chosen, handled = m.click(msg.Y)
	default:
		run, p, ok, cmd := m.mo.run.Update(msg)
		if !ok {
			return m, bezel.Result{}
		}
		m.mo.run, m.mo.p = run, p
		if !m.Active() {
			m.mo = motion{dur: m.mo.dur, p: 1}
		}
		return m, bezel.Result{Handled: true, Cmd: cmd}
	}
	if !handled {
		return m, bezel.Result{}
	}
	r := bezel.Result{Handled: true}
	if m.mo.run.ID != before {
		r.Cmd = m.Tick() // a new motion began; the old frames are foreign now
	}
	if row, ok := m.Selected(); chosen && ok {
		r.Event = Chosen[T]{Row: row}
	}
	return m, r
}

// handles reports whether k is one of the list's bindings, or text for an
// open filter.
func (m Model[T]) handles(k tea.KeyPressMsg) bool {
	ks := m.keys
	if m.find.on {
		return key.Matches(k, ks.Esc, ks.Enter, ks.Erase, ks.Up, ks.Down) ||
			utf8.RuneCountInString(k.Key().Text) == 1
	}
	return key.Matches(k, ks.Filter, ks.Up, ks.Down, ks.PageUp, ks.PageDown, ks.Top, ks.Bottom, ks.Enter)
}

// press is one key: the list after it and whether it chose a row.
func (m Model[T]) press(msg tea.KeyPressMsg) (Model[T], bool) {
	if m.find.on {
		return m.updateFilter(msg)
	}
	switch {
	case key.Matches(msg, m.keys.Filter):
		m.find = m.filterFor("")
	case key.Matches(msg, m.keys.Up):
		return m.move(-1), false
	case key.Matches(msg, m.keys.Down):
		return m.move(1), false
	case key.Matches(msg, m.keys.PageUp):
		return m.jump(-max(m.win.Height, 1)), false
	case key.Matches(msg, m.keys.PageDown):
		return m.jump(max(m.win.Height, 1)), false
	case key.Matches(msg, m.keys.Top):
		return m.jump(-len(m.rows)), false
	case key.Matches(msg, m.keys.Bottom):
		return m.jump(len(m.rows)), false
	case key.Matches(msg, m.keys.Enter):
		_, ok := m.Selected()
		return m, ok
	}
	return m, false
}

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
	case key.Matches(msg, m.keys.Up, m.keys.Down):
		d, n := 1, len(m.find.idx)
		if key.Matches(msg, m.keys.Up) {
			d = -1
		}
		m.find.win = m.find.win.Move(m.wrapped(d, m.find.win.Cursor, n), n, nil)
	}
	return m, false
}

// click is a left click on line y of View: it moves the cursor to a row, or
// chooses the row already under it, as enter would. It reports whether it chose
// and whether y fell on a row; headings, the title and the empty tail are not.
func (m Model[T]) click(y int) (Model[T], bool, bool) {
	if m.title != "" {
		y--
	}
	if m.height <= 0 || y < 0 {
		return m, false, false
	}
	if m.find.on {
		at := y / m.itemLines()
		w := m.find.win.Clamp(len(m.find.idx))
		start, end, _ := w.Bounds(len(m.find.idx))
		if at >= end-start {
			return m, false, false
		}
		if start+at == w.Cursor {
			m, chosen := m.pickMatch()
			return m, chosen, true
		}
		m.find.win.Cursor = start + at
		return m, false, true
	}
	i := -1
	lines, of := m.rowLines(new(Styles)) // what View draws, ghosts included
	for k, l := range lines {
		if y -= strings.Count(l, "\n") + 1; y < 0 {
			i = of[k]
			break
		}
	}
	switch {
	case i < 0 || m.rows[i].Section:
		return m, false, false
	case i == m.win.Cursor:
		return m, true, true
	}
	from := m.win.Cursor
	m.win.Cursor = i
	return m.start(float64(from)).clampScroll(), false, true
}

// pickMatch closes the filter on the match under its cursor, chosen, or on
// nothing when nothing matched.
func (m Model[T]) pickMatch() (Model[T], bool) {
	idx, at := m.find.idx, m.find.win.Cursor
	m.find = filterMode{}
	if at >= len(idx) {
		return m, false
	}
	from := m.win.Cursor
	m.win.Cursor = idx[at]
	return m.start(float64(from)).clampScroll(), true
}

// filterFor is the prompt open on text, its matches found here once.
func (m Model[T]) filterFor(text string) filterMode {
	return filterMode{on: true, text: text, idx: m.filtered(text), win: layout.Window{Height: m.filterRows()}}
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

// filtered is the indices into rows that match query, so a pick maps back
// to its own row even when labels repeat.
func (m Model[T]) filtered(query string) []int {
	if m.fuzzy && query != "" {
		return m.ranked(query)
	}
	match := m.match
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
func (m Model[T]) ranked(query string) []int {
	labels := make([]string, len(m.rows))
	for i, r := range m.rows {
		if !r.Section { // a heading's empty label matches no query
			labels[i] = r.Label
		}
	}
	return fuzzy.Ranked(query, labels)
}

// move steps the cursor over headings, clamped at the ends, or round them
// when the list wraps.
func (m Model[T]) move(delta int) Model[T] {
	moved := m.jump(delta)
	if !m.wrap || moved.win.Cursor != m.win.Cursor || !m.picked() {
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

// clampFilter keeps the filter's window on its matches.
func (m Model[T]) clampFilter() Model[T] {
	if m.find.on {
		m.find.win = m.find.win.Clamp(len(m.find.idx))
	}
	return m
}

// refilter matches the open filter again after the rows or the match changed.
func (m Model[T]) refilter() Model[T] {
	if m.find.on {
		m.find.idx = m.filtered(m.find.text)
	}
	return m.clampFilter()
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
	st := m.styles(th)
	var body string
	if m.find.on {
		body = m.viewFilter(&st)
	} else {
		body = m.viewRows(&st)
	}
	if m.title == "" && !m.status && !m.pages {
		return body
	}
	var lines []string
	if m.title != "" {
		lines = append(lines, st.Title.Render(m.title))
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
		lines = append(lines, st.Status.Render(m.statusText()))
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
	case len(m.find.idx) == 0:
		return "  nothing matched"
	}
	return fmt.Sprintf("  %d of %d", len(m.find.idx), n)
}

// viewRows is the rows in the window, or on the cursor's page.
func (m Model[T]) viewRows(st *Styles) string {
	lines, _ := m.rowLines(st)
	return strings.Join(lines, "\n")
}

// bounds is the rows the window shows and how many hide below it; a paged
// list shows the cursor's page and hides none.
func (m Model[T]) bounds() (start, end, below int) {
	if !m.pages {
		return m.win.Bounds(len(m.rows))
	}
	start = m.win.Cursor / m.win.Height * m.win.Height
	return start, min(start+m.win.Height, len(m.rows)), 0
}

// rowLines is what viewRows draws, a row or a ghost a piece, with the row
// each piece is: -1 for a ghost or the "more" row.
func (m Model[T]) rowLines(st *Styles) (lines []string, of []int) {
	if m.win.Height <= 0 {
		return nil, nil
	}
	start, end, below := m.bounds()
	cursor := m.cursorRow()

	two := m.itemLines() == 2
	lines, of = make([]string, 0, end-start), make([]int, 0, end-start)
	for i := start; i < end; i++ {
		if g, ok := m.ghostLine(i, st); ok {
			lines, of = append(lines, g), append(of, -1)
		}
		r, arriving := m.rows[i], m.arriving(m.rows[i], st)
		switch {
		case r.Section:
			heading := st.Section.Render(oneLine(r.Label))
			if two {
				heading += "\n"
			}
			lines = append(lines, heading)
		case i != cursor && arriving != nil:
			r.Style = arriving
			lines = append(lines, renderRow(r, false, st, nil, two))
		default:
			lines = append(lines, renderRow(r, i == cursor, st, nil, two))
		}
		of = append(of, i)
	}
	if g, ok := m.ghostLine(end, st); ok && below == 0 {
		lines, of = append(lines, g), append(of, -1)
	}
	room := max(end-start, m.win.Height-boolInt(below > 0)) // a ghost never pushes the list past its height
	lines, of = lines[:min(len(lines), room)], of[:min(len(of), room)]
	if below > 0 {
		lines, of = append(lines, moreRow(below, st)), append(of, -1)
	}
	return lines, of
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// viewFilter is the matches in the filter's window, padded to it, then the prompt.
func (m Model[T]) viewFilter(st *Styles) string {
	idx := m.find.idx
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
		lines = append(lines, renderRow(r, i == w.Cursor, st, hits, two))
	}
	if below > 0 {
		lines = append(lines, moreRow(below, st))
	}
	for len(lines) < w.Height {
		lines = append(lines, "")
	}
	lines = append(lines, st.Cursor.Render("/"+m.find.text+"▋"))
	return strings.Join(lines, "\n")
}

// moreRow stands in for the last row while n rows are hidden below it.
func moreRow(n int, st *Styles) string { return draw.More(n, 2, st.More) }

// renderRow is a row in its style, the runes of its label at hits
// underlined, and with two its description, or a blank, under it.
func renderRow[T any](r Row[T], selected bool, st *Styles, hits []int, two bool) string {
	style, lead := st.Row, "  "
	switch {
	case selected:
		style, lead = st.Cursor, "▶ "
	case r.Style != nil:
		style = *r.Style
	}
	if r.Mark != "" {
		lead += r.Mark + "  "
	}
	row := style.Render(lead) + underlined(oneLine(r.Label), hits, style)
	if two {
		indent := strings.Repeat(" ", utf8.RuneCountInString(lead))
		row += "\n" + st.Description.Render(indent+oneLine(r.Description))
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

// oneLine is s with its line breaks as spaces, rune for rune, so a row
// keeps the lines the list counted for it and its filter hits.
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, s)
}
