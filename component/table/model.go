package table

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/spinner"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Item is one row of a Model: its cells in column order, and a payload of
// the caller's type T. A cell may carry its own colour. A Busy row shows a
// turning spinner where the cursor's marker goes.
type Item[T any] struct {
	Cells   []string
	Value   T
	Busy    bool
	Checked bool // ticked for a batch, while the table is WithMulti
}

// Keys are the bindings the table answers to.
type Keys struct {
	Up, Down, PageUp, PageDown, HalfPageUp, HalfPageDown, Top, Bottom, Enter, Filter, Esc key.Binding
	Erase                                                                                 key.Binding // the filter's last character
	Check, CheckAll                                                                       key.Binding // while WithMulti
}

// DefaultKeys are the arrows, pgup/pgdown, ctrl+u/ctrl+d for half a page,
// home/end, enter, / and esc; space and ctrl+a check rows while WithMulti.
func DefaultKeys() Keys {
	return Keys{
		Up:           key.New("up").WithHelp("↑/↓", "move"),
		Down:         key.New("down"),
		PageUp:       key.New("pgup").WithHelp("pgup/pgdn", "page"),
		PageDown:     key.New("pgdown"),
		HalfPageUp:   key.New("ctrl+u").WithHelp("ctrl+u/d", "half page"),
		HalfPageDown: key.New("ctrl+d"),
		Top:          key.New("home").WithHelp("home/end", "ends"),
		Bottom:       key.New("end"),
		Enter:        key.New("enter").WithHelp("enter", "choose"),
		Filter:       key.New("/").WithHelp("/", "filter"),
		Esc:          key.New("esc").WithHelp("esc", "stop filtering"),
		Erase:        key.New("backspace", "ctrl+h"),
		Check:        key.New("space").WithHelp("space", "check"),
		CheckAll:     key.New("ctrl+a").WithHelp("ctrl+a", "check all"),
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

// marker is what the cursor row starts with. Every row keeps its width, so
// the columns stay put when the cursor lands.
const marker = "▶ "

var _ bezel.Model[Model[int]] = Model[int]{}

// Model is a table with a cursor: the header stays put and the rows under it
// scroll. A value: every method returns the new table.
type Model[T any] struct {
	cols   []Column
	items  []Item[T]
	widths []int // natural column widths over items, measured when they change
	height int
	width  int
	win    layout.Window // over the rows, one shorter than height for the header
	keys   Keys
	find   filterMode
	match  func(query, text string) bool // nil is Contains
	fuzzy  bool
	spin   spinner.Model // turns while a row is Busy
	blur   bool          // drawn without the cursor's colour, and deaf to keys
	multi  bool
	edit   func(Styles) Styles
}

// Styles are how a table paints its parts.
type Styles struct {
	Header  lipgloss.Style // the column titles
	Cursor  lipgloss.Style // the cursor row's marker and its plain cells
	Blurred lipgloss.Style // the cursor row's marker while the table is blurred
	More    lipgloss.Style // the "↓ N more" row
	Busy    lipgloss.Style // the spinner on a Busy row
	Box     lipgloss.Style // an unchecked row's box while WithMulti
	Checked lipgloss.Style // a checked row's box
}

// DefaultStyles take the header from the theme's table role, paint the
// cursor row in the cursor role, a busy row's spinner and a checked box in
// the accent and an empty box muted.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{
		Header: th.TableHeader, Cursor: th.Cursor, Blurred: th.Dim, More: th.Dim, Busy: th.Accent,
		Box: th.Muted, Checked: th.Accent,
	}
}

// New builds a table with the cursor on the first row; SetHeight gives it
// the rows its pane has, header included. With a Busy row, Tick turns the
// spinner.
func New[T any](cols []Column, items []Item[T]) Model[T] {
	m := Model[T]{cols: slices.Clone(cols), keys: DefaultKeys(), spin: spinner.New(spinner.MiniDot())}
	return m.withItems(items).SetHeight(0)
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

// WithKeys rebinds the table.
func (m Model[T]) WithKeys(k Keys) Model[T] {
	m.keys = k
	return m
}

// Focus lets the table take keys and paints its cursor row; a table is
// focused when built.
func (m Model[T]) Focus() Model[T] {
	m.blur = false
	return m
}

// Blur stops the table taking keys and draws its cursor row dimmed, for a
// pane that lost the focus.
func (m Model[T]) Blur() Model[T] {
	m.blur = true
	return m
}

// Focused reports whether the table takes keys.
func (m Model[T]) Focused() bool { return !m.blur }

// Legend lists the keys the table answers to now: the filter's while it is open.
func (m Model[T]) Legend() []legend.Entry {
	k := m.keys
	if m.find.on {
		return legend.FromBindings([]key.Binding{k.Up, k.Down, k.Enter, k.Esc})
	}
	keys := []key.Binding{k.Up, k.Down, k.PageUp, k.PageDown, k.HalfPageUp, k.HalfPageDown, k.Top, k.Bottom, k.Enter, k.Filter}
	if m.multi {
		keys = append(keys, k.Check, k.CheckAll)
	}
	return legend.FromBindings(keys)
}

// FromValues is the rows of delimited text, one per line, its cells split
// at sep; blank lines are skipped. The payload is left zero.
func FromValues[T any](text, sep string) []Item[T] {
	var items []Item[T]
	for line := range strings.SplitSeq(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		items = append(items, Item[T]{Cells: strings.Split(line, sep)})
	}
	return items
}

// Items is a copy of every row, cells included.
func (m Model[T]) Items() []Item[T] { return cloneItems(m.items) }

// Len is how many rows there are.
func (m Model[T]) Len() int { return len(m.items) }

// Cursor is the index of the row under the cursor, the filter's aside.
func (m Model[T]) Cursor() int { return m.win.Cursor }

// Selected is a copy of the row under the cursor, false when there are none.
// While filtering it is the row under the filter cursor.
func (m Model[T]) Selected() (Item[T], bool) {
	at := m.win.Cursor
	if m.find.on {
		if m.find.win.Cursor >= len(m.find.idx) {
			return Item[T]{}, false
		}
		at = m.find.idx[m.find.win.Cursor]
	}
	if at >= len(m.items) {
		return Item[T]{}, false
	}
	it := m.items[at]
	it.Cells = slices.Clone(it.Cells)
	return it, true
}

// Scroll is where the table shows its rows from, for a pane's scrollbar;
// while filtering, where the matches show from.
func (m Model[T]) Scroll() layout.Scroll {
	if m.find.on {
		n := len(m.find.idx)
		return m.find.win.Clamp(n).Scroll(n)
	}
	return m.win.Clamp(len(m.items)).Scroll(len(m.items))
}

// SetItems replaces the rows, keeping the cursor and the checks on the rows
// with the same first cell (the cursor at the same index first), else at the
// same index: a refresh that drops a row lands on the one that took its place.
func (m Model[T]) SetItems(items []Item[T]) Model[T] {
	prev, at := "", m.win.Cursor
	if at < len(m.items) {
		prev = first(m.items[at])
	}
	match, found := m.matched()
	before := m.items
	m = m.withItems(items)
	keepChecks(before, m.items)
	if at < len(m.items) && first(m.items[at]) == prev {
		return m.clamp().refilter().refind(match, found)
	}
	for i, it := range m.items {
		if first(it) == prev {
			m.win.Cursor = i
			break
		}
	}
	return m.clamp().refilter().refind(match, found)
}

// SetHeight updates the rows the table takes, header included.
func (m Model[T]) SetHeight(h int) Model[T] {
	m.height = h
	m.win.Height = max(h-1, 0)
	return m.clamp().refilter()
}

// Update takes the table's keys unless blurred, a bezel.ClickMsg on a row and
// the Busy spinner's frames. Choosing a row or a / match reports a [Chosen],
// checking one a [Toggled]; any other message comes back unhandled.
func (m Model[T]) Update(msg tea.Msg) (Model[T], bezel.Result) {
	var chosen, handled bool
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.blur || !m.handles(msg) {
			return m, bezel.Result{}
		}
		if m.multi && !m.find.on && key.Matches(msg, m.keys.Check, m.keys.CheckAll) {
			var items []Item[T]
			if key.Matches(msg, m.keys.Check) {
				m, items = m.check()
			} else {
				m, items = m.checkAll()
			}
			r := bezel.Result{Handled: true}
			if len(items) > 0 {
				r.Event = Toggled[T]{Items: items}
			}
			return m, r
		}
		if m.find.on {
			m, chosen = m.updateFilter(msg)
		} else {
			m, chosen = m.press(msg)
		}
		handled = true
	case bezel.ClickMsg:
		m, chosen, handled = m.click(msg.Y)
	default:
		if !m.spin.Owns(msg) {
			return m, bezel.Result{}
		}
		var r bezel.Result
		m.spin, r = m.spin.Update(msg)
		return m, r
	}
	if !handled {
		return m, bezel.Result{}
	}
	r := bezel.Result{Handled: true}
	if it, ok := m.Selected(); chosen && ok {
		r.Event = Chosen[T]{Item: it}
	}
	return m, r
}

// handles reports whether k is one of the table's bindings, or text for an
// open filter.
func (m Model[T]) handles(k tea.KeyPressMsg) bool {
	ks := m.keys
	if m.find.on {
		return key.Matches(k, ks.Esc, ks.Enter, ks.Erase, ks.Up, ks.Down) || edit.Typed(k) != ""
	}
	return key.Matches(k, ks.Filter, ks.HalfPageUp, ks.HalfPageDown, ks.Up, ks.Down,
		ks.PageUp, ks.PageDown, ks.Top, ks.Bottom, ks.Enter) ||
		m.multi && key.Matches(k, ks.Check, ks.CheckAll)
}

// press is one of the table's keys: the table after it and whether it chose.
func (m Model[T]) press(msg tea.KeyPressMsg) (Model[T], bool) {
	n := len(m.items)
	switch {
	case key.Matches(msg, m.keys.Filter):
		m.find = m.filterFor("")
	case key.Matches(msg, m.keys.HalfPageUp):
		m.win = m.win.Move(-max(m.page()/2, 1), n, nil)
	case key.Matches(msg, m.keys.HalfPageDown):
		m.win = m.win.Move(max(m.page()/2, 1), n, nil)
	case key.Matches(msg, m.keys.Up):
		m.win = m.win.Move(-1, n, nil)
	case key.Matches(msg, m.keys.Down):
		m.win = m.win.Move(1, n, nil)
	case key.Matches(msg, m.keys.PageUp):
		m.win = m.win.Move(-m.page(), n, nil)
	case key.Matches(msg, m.keys.PageDown):
		m.win = m.win.Move(m.page(), n, nil)
	case key.Matches(msg, m.keys.Top):
		m.win = m.win.Move(-n, n, nil)
	case key.Matches(msg, m.keys.Bottom):
		m.win = m.win.Move(n, n, nil)
	case key.Matches(msg, m.keys.Enter):
		return m, n > 0
	}
	return m, false
}

// click is a left click on line y of View: it moves the cursor to a row, or
// chooses the row already under it. It reports whether it chose and whether y
// fell on a row (not the header or "more"). A blurred table takes it too.
func (m Model[T]) click(y int) (Model[T], bool, bool) {
	if m.find.on {
		w := m.find.win.Clamp(len(m.find.idx))
		start, end, _ := w.Bounds(len(m.find.idx))
		i := start + y - 1
		switch {
		case y < 1 || i >= end:
			return m, false, false
		case i == w.Cursor:
			m, chosen := m.pickMatch()
			return m, chosen, true
		}
		m.find.win.Cursor = i
		return m, false, true
	}
	w := m.win.Clamp(len(m.items))
	start, end, _ := w.Bounds(len(m.items))
	i := start + y - 1 // the header takes line 0
	switch {
	case y < 1 || i >= end:
		return m, false, false
	case i == w.Cursor:
		return m, true, true
	}
	m.win.Cursor = i
	return m.clamp(), false, true
}

// page is how far pgup and pgdown move: the rows shown, less the "more" row.
func (m Model[T]) page() int { return max(1, m.win.Height-1) }

func (m Model[T]) clamp() Model[T] {
	m.win = m.win.Clamp(len(m.items))
	return m
}

// withItems takes a copy of items and measures their columns once, so View
// does not walk every row each frame.
func (m Model[T]) withItems(items []Item[T]) Model[T] {
	m.items = cloneItems(items)
	rows := make([][]string, len(m.items))
	for i, it := range m.items {
		rows[i] = it.Cells
	}
	m.widths = natural(m.cols, rows)
	return m.spun()
}

// SetWidth changes the cells the table takes; columns shrink to fit.
func (m Model[T]) SetWidth(w int) Model[T] {
	m.width = w
	return m
}

// View is the header and the rows in its window, SetWidth cells wide, and
// nothing until SetHeight and SetWidth size it. Columns are sized over every
// row, so scrolling never moves them; a plain cell wears the cursor's colour.
func (m Model[T]) View(th theme.Resolved) string {
	if m.height <= 0 || m.width <= 0 {
		return ""
	}
	width := m.width
	st := m.styles(th)
	pad := strings.Repeat(" ", draw.Width(marker))
	if m.multi {
		pad += strings.Repeat(" ", draw.Width("[ ] "))
	}
	widths := shrink(m.cols, slices.Clone(m.widths), width-len(pad))
	lines := []string{st.Header.Render(pad + Titles(m.cols, widths))}

	w, rows := m.win, m.items
	if m.find.on {
		w, rows = m.find.win, make([]Item[T], len(m.find.idx))
		for k, i := range m.find.idx {
			rows[k] = m.items[i]
		}
	}
	w = w.Clamp(len(rows))
	start, end, below := w.Bounds(len(rows))
	busy := m.busyMark(st)
	for i := start; i < end; i++ {
		cells := rows[i].Cells
		lead := strings.Repeat(" ", draw.Width(marker))
		switch {
		case rows[i].Busy:
			lead = busy
		case i == w.Cursor && m.blur:
			lead = st.Blurred.Render(marker)
		case i == w.Cursor:
			lead = st.Cursor.Render(marker)
		}
		if i == w.Cursor && !m.blur {
			cells = cursorCells(cells, st.Cursor)
		}
		lines = append(lines, lead+m.box(rows[i], st)+Row(widths, cells))
	}
	if below > 0 {
		lines = append(lines, draw.More(below, draw.Width(marker), st.More))
	}
	if m.find.on && m.height > 1 {
		for len(lines) < m.height-1 {
			lines = append(lines, "")
		}
		lines = append(lines, st.Cursor.Render(draw.Truncate("/"+m.find.text+"▋", width)))
	}
	return strings.Join(lines, "\n")
}

// cursorCells paints each plain cell in the cursor's style, one by one, so a
// coloured cell keeps its colour and its reset ends nothing but itself.
func cursorCells(cells []string, cursor lipgloss.Style) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = c
		if c != "" && ansi.Strip(c) == c {
			out[i] = cursor.Render(c)
		}
	}
	return out
}

// first is the cell SetItems knows a row by.
func first[T any](it Item[T]) string {
	if len(it.Cells) == 0 {
		return ""
	}
	return it.Cells[0]
}

// cloneItems copies items down to their cells, so neither side's later
// writes to a Cells slice reach the other. The payload is the caller's.
func cloneItems[T any](items []Item[T]) []Item[T] {
	out := make([]Item[T], len(items))
	for i, it := range items {
		out[i] = Item[T]{Cells: slices.Clone(it.Cells), Value: it.Value, Busy: it.Busy, Checked: it.Checked}
	}
	return out
}
