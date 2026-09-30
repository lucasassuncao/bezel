// Package browser is the pick-from-a-list screen: labels on the left, the
// selected item's detail on the right, tab to move focus between them. It
// owns only the browsing state; what a choice means is the caller's.
package browser

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

// Item is one row: what the list prints and what the preview shows for it.
// Detail is called on render, so an expensive body is only built when shown.
type Item struct {
	Label  string
	Detail func() string
}

// Action is what a key did.
type Action int

const (
	None      Action = iota
	Dismissed        // esc on the list: close without choosing
	Chosen           // enter on the list: the caller reads Selected
)

// Keys are the bindings the browser answers to.
type Keys struct {
	Up, Down, PageUp, PageDown, Enter, Esc, Tab key.Binding
}

func DefaultKeys() Keys {
	return Keys{
		Up:       key.NewBinding(key.WithKeys("up")),
		Down:     key.NewBinding(key.WithKeys("down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup")),
		PageDown: key.NewBinding(key.WithKeys("pgdown")),
		Enter:    key.NewBinding(key.WithKeys("enter")),
		Esc:      key.NewBinding(key.WithKeys("esc")),
		Tab:      key.NewBinding(key.WithKeys("tab")),
	}
}

// Model is a value: Update returns the new one.
type Model struct {
	items  []Item
	cursor int
	keys   Keys

	PreviewFocus  bool // the right pane has the keys
	previewScroll int  // first detail line shown
	previewHeight int  // rows the preview pane has, for paging
}

// New builds a browser over items with the cursor on the one labelled
// current, or the first when none is.
func New(items []Item, current string) Model {
	m := Model{items: slices.Clone(items), keys: DefaultKeys()}
	for i, it := range items {
		if it.Label == current {
			m.cursor = i
			break
		}
	}
	return m
}

func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// SetItems replaces the rows, keeping the cursor in range and the preview at
// the top: a new list is a new thing to read.
func (m Model) SetItems(items []Item) Model {
	m.items = slices.Clone(items)
	m.cursor = min(max(m.cursor, 0), max(len(items)-1, 0))
	m.previewScroll = 0
	return m
}

// SetPreviewHeight tells the browser how tall its preview pane is, which is
// what a page is. Call it from the owner's relayout.
func (m Model) SetPreviewHeight(h int) Model {
	m.previewHeight = h
	return m
}

func (m Model) Items() []Item { return slices.Clone(m.items) }
func (m Model) Cursor() int   { return m.cursor }
func (m Model) Len() int      { return len(m.items) }

// Selected is the item under the cursor, or the zero Item when there is none.
func (m Model) Selected() Item {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return Item{}
	}
	return m.items[m.cursor]
}

// Update handles one key. Keys the browser does not know return None and
// leave it unchanged, so the caller can layer its own on top.
func (m Model) Update(msg tea.KeyPressMsg) (Model, Action) {
	switch {
	case key.Matches(msg, m.keys.Esc):
		if m.PreviewFocus {
			m.PreviewFocus = false
			return m, None
		}
		return m, Dismissed
	case key.Matches(msg, m.keys.Tab):
		m.PreviewFocus = !m.PreviewFocus
	case key.Matches(msg, m.keys.Enter):
		if !m.PreviewFocus && len(m.items) > 0 {
			return m, Chosen
		}
	case key.Matches(msg, m.keys.Up):
		return m.step(-1, 1), None
	case key.Matches(msg, m.keys.Down):
		return m.step(1, 1), None
	case key.Matches(msg, m.keys.PageUp):
		return m.step(-1, max(1, m.previewHeight/2)), None
	case key.Matches(msg, m.keys.PageDown):
		return m.step(1, max(1, m.previewHeight/2)), None
	}
	return m, None
}

// step moves the cursor by dir when the list has focus, or scrolls the
// preview by dir*n when it does.
func (m Model) step(dir, n int) Model {
	if m.PreviewFocus {
		m.previewScroll = min(max(0, m.previewScroll+dir*n), m.maxPreviewScroll())
		return m
	}
	next := m.cursor + dir
	if next >= 0 && next < len(m.items) {
		m.cursor = next
		m.previewScroll = 0
	}
	return m
}

// maxPreviewScroll is the scroll that shows the detail's last full page.
func (m Model) maxPreviewScroll() int {
	it := m.Selected()
	if it.Detail == nil {
		return 0
	}
	lines := strings.Count(it.Detail(), "\n") + 1
	return max(lines-max(m.previewHeight, 1), 0)
}

// ListView renders the labels, height rows of them, scrolled to keep the
// cursor visible and centred where the list allows.
func (m Model) ListView(th theme.Resolved, height int) string {
	scroll := layout.ScrollStart(m.cursor, len(m.items), height)
	end := len(m.items)
	if height > 0 {
		end = min(end, scroll+height)
	}
	var sb strings.Builder
	for i := scroll; i < end; i++ {
		if i > scroll {
			sb.WriteByte('\n')
		}
		if i == m.cursor {
			sb.WriteString(th.Cursor.Render("▶  " + m.items[i].Label))
		} else {
			sb.WriteString(th.Dim.Render("   " + m.items[i].Label))
		}
	}
	return sb.String()
}

// PreviewView renders the selected item's detail, height rows of it from the
// current scroll, clamped so the last page is full.
func (m Model) PreviewView(height int) string {
	it := m.Selected()
	if it.Detail == nil {
		return ""
	}
	full := it.Detail()
	if full == "" {
		return ""
	}
	lines := strings.Split(full, "\n")
	height = max(height, 1)
	scroll := min(m.previewScroll, max(len(lines)-height, 0))
	return strings.Join(lines[scroll:min(scroll+height, len(lines))], "\n")
}
