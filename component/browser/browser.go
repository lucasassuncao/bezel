// Package browser is the pick-from-a-list screen: labels on the left, the
// selected item's detail on the right, tab to move focus between them. It
// owns only the browsing state; what a choice means is the caller's.
package browser

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Item is one row: what the list prints and what the preview shows for it.
// Detail is called on render, so an expensive body is only built when shown.
type Item struct {
	Label  string
	Detail func() string
}

// Keys are the bindings the browser answers to.
type Keys struct {
	Up, Down, PageUp, PageDown, Enter, Esc, Tab key.Binding
}

// DefaultKeys are the arrows, pgup/pgdown, enter, esc and tab.
func DefaultKeys() Keys {
	return Keys{
		Up:       key.New("up").WithHelp("↑/↓", "move"),
		Down:     key.New("down"),
		PageUp:   key.New("pgup").WithHelp("pgup/pgdn", "page"),
		PageDown: key.New("pgdown"),
		Enter:    key.New("enter").WithHelp("enter", "choose"),
		Esc:      key.New("esc").WithHelp("esc", "back"),
		Tab:      key.New("tab").WithHelp("tab", "preview"),
	}
}

// VimKeys are DefaultKeys with k/j added to move.
func VimKeys() Keys {
	k := DefaultKeys()
	k.Up, k.Down = key.New("up", "k").WithHelp("↑/↓ k/j", "move"), key.New("down", "j")
	return k
}

var _ bezel.Model[Model] = Model{}

// Model is a value: Update returns the new one.
type Model struct {
	items  []Item
	cursor int
	keys   Keys

	previewFocus  bool // the right pane has the keys
	previewScroll int  // first detail line shown
	previewHeight int  // rows the preview pane has, for paging
	listHeight    int  // rows the list pane has; 0 shows every label
	edit          func(Styles) Styles
}

// Styles are how the list pane paints its labels; the preview is the
// item's own text.
type Styles struct {
	Cursor lipgloss.Style // the label under the cursor
	Label  lipgloss.Style // every other label
}

// DefaultStyles paint the cursor in the cursor role over dim labels.
func DefaultStyles(th theme.Resolved) Styles { return Styles{Cursor: th.Cursor, Label: th.Dim} }

// New builds a browser over items with the cursor on the first.
func New(items []Item) Model {
	return Model{items: slices.Clone(items), keys: DefaultKeys()}
}

// WithCurrent puts the cursor on the item labelled label; an unknown label
// leaves it where it was.
func (m Model) WithCurrent(label string) Model {
	for i, it := range m.items {
		if it.Label == label {
			m.cursor = i
			break
		}
	}
	return m
}

// WithStyles changes the styles derived from the theme through edit, under
// whatever theme View gets. It replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

// styles is DefaultStyles of th with the app's edit applied.
func (m Model) styles(th theme.Resolved) Styles {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	return st
}

// PreviewFocused reports whether the preview pane has the keys.
func (m Model) PreviewFocused() bool { return m.previewFocus }

// SetPreviewFocus hands the keys to the preview pane, or back to the list.
func (m Model) SetPreviewFocus(on bool) Model {
	m.previewFocus = on
	return m
}

// WithKeys rebinds the browser.
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

// SetListHeight tells the browser how tall its list pane is; 0 shows every
// label.
func (m Model) SetListHeight(h int) Model {
	m.listHeight = h
	return m
}

// SetPreviewHeight tells the browser how tall its preview pane is, which is
// what a page is. Call it from the owner's relayout.
func (m Model) SetPreviewHeight(h int) Model {
	m.previewHeight = h
	return m
}

// Items is a copy of the rows.
func (m Model) Items() []Item { return slices.Clone(m.items) }

// Cursor is the index of the row under the cursor.
func (m Model) Cursor() int { return m.cursor }

// Len is how many rows there are.
func (m Model) Len() int { return len(m.items) }

// Selected is the item under the cursor, false when there is none.
func (m Model) Selected() (Item, bool) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return Item{}, false
	}
	return m.items[m.cursor], true
}

// Legend is the keys the browser answers to: enter only while the list has
// the keys.
func (m Model) Legend() []legend.Entry {
	bs := []key.Binding{m.keys.Up, m.keys.PageUp, m.keys.Tab}
	if !m.previewFocus {
		bs = append(bs, m.keys.Enter)
	}
	return legend.FromBindings(append(bs, m.keys.Esc))
}

// Update takes the browser's keys. Enter on the list reports a [Chosen] and
// esc on the list a [Dismissed]; any other message comes back unhandled and
// unchanged.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok || !key.Matches(k, m.keys.Esc, m.keys.Tab, m.keys.Enter, m.keys.Up, m.keys.Down, m.keys.PageUp, m.keys.PageDown) {
		return m, bezel.Result{}
	}
	r := bezel.Result{Handled: true}
	switch {
	case key.Matches(k, m.keys.Esc):
		if m.previewFocus {
			m.previewFocus = false
		} else {
			r.Event = Dismissed{}
		}
	case key.Matches(k, m.keys.Tab):
		m.previewFocus = !m.previewFocus
	case key.Matches(k, m.keys.Enter):
		if it, ok := m.Selected(); ok && !m.previewFocus {
			r.Event = Chosen{Item: it}
		}
	case key.Matches(k, m.keys.Up):
		m = m.step(-1, 1)
	case key.Matches(k, m.keys.Down):
		m = m.step(1, 1)
	case key.Matches(k, m.keys.PageUp):
		m = m.step(-1, max(1, m.previewHeight/2))
	case key.Matches(k, m.keys.PageDown):
		m = m.step(1, max(1, m.previewHeight/2))
	}
	return m, r
}

// step moves the cursor by dir when the list has focus, or scrolls the
// preview by dir*n when it does.
func (m Model) step(dir, n int) Model {
	if m.previewFocus {
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
	it, _ := m.Selected()
	if it.Detail == nil {
		return 0
	}
	lines := strings.Count(it.Detail(), "\n") + 1
	return max(lines-max(m.previewHeight, 1), 0)
}

// ListView renders the labels, SetListHeight rows of them, scrolled to keep
// the cursor visible and centred where the list allows.
func (m Model) ListView(th theme.Resolved) string {
	height := m.listHeight
	st := m.styles(th)
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
			sb.WriteString(st.Cursor.Render("▶  " + oneLine(m.items[i].Label)))
		} else {
			sb.WriteString(st.Label.Render("   " + oneLine(m.items[i].Label)))
		}
	}
	return sb.String()
}

// PreviewView renders the selected item's detail, SetPreviewHeight rows of it
// from the current scroll, clamped so the last page is full.
func (m Model) PreviewView() string {
	height := m.previewHeight
	it, _ := m.Selected()
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

// oneLine is s with its line breaks as spaces, so a label keeps one row.
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, s)
}
