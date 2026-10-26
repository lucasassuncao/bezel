package table

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Item is one row of a Model: its cells in column order, and a payload of
// the caller's type T. A cell may carry its own colour.
type Item[T any] struct {
	Cells []string
	Value T
}

// Action is what a key did.
type Action int

const (
	None   Action = iota
	Chosen        // enter on a row
)

// Keys are the bindings the table answers to.
type Keys struct {
	Up, Down, PageUp, PageDown, HalfPageUp, HalfPageDown, Top, Bottom, Enter key.Binding
}

// DefaultKeys are the arrows, pgup/pgdown, ctrl+u/ctrl+d for half a page,
// home/end and enter.
func DefaultKeys() Keys {
	return Keys{
		Up:           key.NewBinding(key.WithKeys("up"), key.WithHelp("↑/↓", "move")),
		Down:         key.NewBinding(key.WithKeys("down")),
		PageUp:       key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup/pgdn", "page")),
		PageDown:     key.NewBinding(key.WithKeys("pgdown")),
		HalfPageUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u/d", "half page")),
		HalfPageDown: key.NewBinding(key.WithKeys("ctrl+d")),
		Top:          key.NewBinding(key.WithKeys("home"), key.WithHelp("home/end", "ends")),
		Bottom:       key.NewBinding(key.WithKeys("end")),
		Enter:        key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "choose")),
	}
}

// marker is what the cursor row starts with. Every row keeps its width, so
// the columns stay put when the cursor lands.
const marker = "▶ "

// Model is a table with a cursor: the header stays put and the rows under it
// scroll. A value: every method returns the new table.
type Model[T any] struct {
	cols   []Column
	items  []Item[T]
	height int
	win    layout.Window // over the rows, one shorter than height for the header
	keys   Keys
	blur   bool // drawn without the cursor's colour, and deaf to keys
}

// New builds a table height rows tall, header included, with the cursor on
// the first row.
func New[T any](cols []Column, items []Item[T], height int) Model[T] {
	m := Model[T]{cols: slices.Clone(cols), items: cloneItems(items), keys: DefaultKeys()}
	return m.SetHeight(height)
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

// Legend lists the keys the table answers to.
func (m Model[T]) Legend() []legend.Entry {
	k := m.keys
	return legend.FromBindings([]key.Binding{k.Up, k.Down, k.PageUp, k.PageDown, k.HalfPageUp, k.HalfPageDown, k.Top, k.Bottom, k.Enter})
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

// Cursor is the index of the row under the cursor.
func (m Model[T]) Cursor() int { return m.win.Cursor }

// Selected is the row under the cursor, or nil when there are none.
func (m Model[T]) Selected() *Item[T] {
	if m.win.Cursor >= len(m.items) {
		return nil
	}
	it := m.items[m.win.Cursor]
	it.Cells = slices.Clone(it.Cells)
	return &it
}

// Scroll is where the table shows its rows from, for a pane's scrollbar.
func (m Model[T]) Scroll() layout.Scroll {
	return m.win.Clamp(len(m.items)).Scroll(len(m.items))
}

// SetItems replaces the rows, keeping the cursor on the row with the same
// first cell when it survives (at the same index first), else at the same
// index: a refresh that drops a row lands on the one that took its place.
func (m Model[T]) SetItems(items []Item[T]) Model[T] {
	prev, at := "", m.win.Cursor
	if it := m.Selected(); it != nil {
		prev = first(*it)
	}
	m.items = cloneItems(items)
	if at < len(m.items) && first(m.items[at]) == prev {
		return m.clamp()
	}
	for i, it := range m.items {
		if first(it) == prev {
			m.win.Cursor = i
			break
		}
	}
	return m.clamp()
}

// SetHeight updates the rows the table takes, header included.
func (m Model[T]) SetHeight(h int) Model[T] {
	m.height = h
	m.win.Height = max(h-1, 0)
	return m.clamp()
}

// Update handles one key. Keys the table does not know return None and
// leave it unchanged, so the caller can layer its own on top.
func (m Model[T]) Update(msg tea.KeyPressMsg) (Model[T], Action) {
	n := len(m.items)
	if m.blur {
		return m, None
	}
	switch {
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
		if n > 0 {
			return m, Chosen
		}
	}
	return m, None
}

// Click is a left click on line y of View: on a row it moves the cursor
// there, and on the row already under the cursor it chooses it. The header
// and the "more" row do nothing. A blurred table takes it too.
func (m Model[T]) Click(y int) (Model[T], Action) {
	w := m.win.Clamp(len(m.items))
	start, end, _ := w.Bounds(len(m.items))
	i := start + y - 1 // the header takes line 0
	switch {
	case y < 1 || i >= end:
		return m, None
	case i == w.Cursor:
		return m, Chosen
	}
	m.win.Cursor = i
	return m.clamp(), None
}

// page is how far pgup and pgdown move: the rows shown, less the "more" row.
func (m Model[T]) page() int { return max(1, m.win.Height-1) }

func (m Model[T]) clamp() Model[T] {
	m.win = m.win.Clamp(len(m.items))
	return m
}

// View is the header and the rows in its window, width cells wide. Columns
// are sized over every row, not the ones shown, so scrolling never moves
// them. On the cursor row a cell with no colour of its own wears the cursor's.
func (m Model[T]) View(th theme.Resolved, width int) string {
	if m.height <= 0 {
		return ""
	}
	cells := make([][]string, len(m.items))
	for i, it := range m.items {
		cells[i] = it.Cells
	}
	pad := strings.Repeat(" ", draw.Width(marker))
	widths := Fit(m.cols, cells, width-len(pad))
	lines := []string{th.TableHeader.Render(pad + Titles(m.cols, widths))}

	w := m.win.Clamp(len(m.items))
	start, end, below := w.Bounds(len(m.items))
	for i := start; i < end; i++ {
		switch {
		case i == w.Cursor && m.blur:
			lines = append(lines, th.Dim.Render(marker)+Row(widths, cells[i]))
		case i == w.Cursor:
			lines = append(lines, th.Cursor.Render(marker)+Row(widths, cursorCells(cells[i], &th)))
		default:
			lines = append(lines, pad+Row(widths, cells[i]))
		}
	}
	if below > 0 {
		lines = append(lines, th.Dim.Render(fmt.Sprintf("%s↓ %d more", pad, below)))
	}
	return strings.Join(lines, "\n")
}

// cursorCells paints each plain cell in the cursor's style, one by one, so a
// coloured cell keeps its colour and its reset ends nothing but itself.
func cursorCells(cells []string, th *theme.Resolved) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = c
		if c != "" && ansi.Strip(c) == c {
			out[i] = th.Cursor.Render(c)
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
		out[i] = Item[T]{Cells: slices.Clone(it.Cells), Value: it.Value}
	}
	return out
}
