package table

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
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
	Up, Down, PageUp, PageDown, Top, Bottom, Enter key.Binding
}

// DefaultKeys are the arrows, pgup/pgdown, home/end and enter.
func DefaultKeys() Keys {
	return Keys{
		Up:       key.NewBinding(key.WithKeys("up")),
		Down:     key.NewBinding(key.WithKeys("down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup")),
		PageDown: key.NewBinding(key.WithKeys("pgdown")),
		Top:      key.NewBinding(key.WithKeys("home")),
		Bottom:   key.NewBinding(key.WithKeys("end")),
		Enter:    key.NewBinding(key.WithKeys("enter")),
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
}

// New builds a table height rows tall, header included, with the cursor on
// the first row.
func New[T any](cols []Column, items []Item[T], height int) Model[T] {
	m := Model[T]{cols: slices.Clone(cols), items: slices.Clone(items), keys: DefaultKeys()}
	return m.SetHeight(height)
}

// WithKeys rebinds the table.
func (m Model[T]) WithKeys(k Keys) Model[T] {
	m.keys = k
	return m
}

// Items is a copy of every row.
func (m Model[T]) Items() []Item[T] { return slices.Clone(m.items) }

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
	m.items = slices.Clone(items)
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
	switch {
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

// page is how far pgup and pgdown move: the rows shown, less the "more" row.
func (m Model[T]) page() int { return max(1, m.win.Height-1) }

func (m Model[T]) clamp() Model[T] {
	m.win = m.win.Clamp(len(m.items))
	return m
}

// View is the header and the rows in its window, width cells wide. Columns
// are sized over every row, not the ones shown, so scrolling never moves
// them. The cursor row drops its cells' own colours to wear the cursor's.
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
		if i == w.Cursor {
			lines = append(lines, th.Cursor.Render(marker+Row(widths, plain(cells[i]))))
		} else {
			lines = append(lines, pad+Row(widths, cells[i]))
		}
	}
	if below > 0 {
		lines = append(lines, th.Dim.Render(fmt.Sprintf("%s↓ %d more", pad, below)))
	}
	return strings.Join(lines, "\n")
}

// plain is cells without their escapes, which would end the cursor's style.
func plain(cells []string) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = ansi.Strip(c)
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
