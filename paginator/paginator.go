// Package paginator is pages over a run of items: which page shows, the slice
// of items on it, and the page drawn as dots or as "2/5".
package paginator

import (
	"cmp"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/icon"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Kind is how View draws the page.
type Kind int

const (
	Numbers Kind = iota // "2/5", in the format WithFormat sets
	Dots                // a mark per page, the current one filled
)

// Action is what a key did.
type Action int

const (
	None  Action = iota
	Paged        // the page changed
)

// Keys are the bindings that turn the page. Next has no help of its own:
// Prev's names both.
type Keys struct {
	Prev, Next key.Binding
}

// DefaultKeys are pgup/pgdown, the arrows and h/l.
func DefaultKeys() Keys {
	return Keys{
		Prev: key.NewBinding(key.WithKeys("pgup", "left", "h"), key.WithHelp("←/→", "page")),
		Next: key.NewBinding(key.WithKeys("pgdown", "right", "l")),
	}
}

// Styles are how the page is painted.
type Styles struct {
	Active   lipgloss.Style // the current page's dot
	Inactive lipgloss.Style // every other dot
	Numbers  lipgloss.Style // "2/5"
}

// DefaultStyles fills the current dot in the accent and mutes the rest.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Active: th.Accent, Inactive: th.Muted, Numbers: th.Dim}
}

// Model is a value: every method returns the new paginator.
type Model struct {
	kind    Kind
	page    int
	perPage int // 0 in a Model never built with New, read as 1
	items   int
	format  string
	icons   icon.Set
	keys    Keys
	edit    func(Styles) Styles
}

// New is one page of perPage items, drawn as numbers, until SetItems says how
// many there are. A perPage below 1 is 1.
func New(perPage int) Model {
	return Model{perPage: max(perPage, 1), format: "%d/%d", icons: icon.FromEnv(), keys: DefaultKeys()}
}

// WithKind sets how the page is drawn.
func (m Model) WithKind(k Kind) Model {
	m.kind = k
	return m
}

// WithFormat writes Numbers with f, which gets the page from 1 and the count
// of pages, as "%d/%d" does.
func (m Model) WithFormat(f string) Model {
	m.format = f
	return m
}

// WithIcons draws Dots with set: Running for the current page, Pending for
// the rest.
func (m Model) WithIcons(set icon.Set) Model {
	m.icons = set
	return m
}

// WithKeys rebinds the page keys.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// WithStyles changes the styles derived from the theme through edit, under
// whatever theme View gets. It replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

// SetItems makes as many pages as n items fill, at least one, keeping the
// page within them.
func (m Model) SetItems(n int) Model {
	m.items = max(n, 0)
	return m.SetPage(m.page)
}

// SetPerPage makes a page hold n items, at least one, as after a resize; the
// page turns to the one holding the item that led the page before.
func (m Model) SetPerPage(n int) Model {
	first := m.page * m.PerPage()
	m.perPage = max(n, 1)
	return m.SetPage(first / m.perPage)
}

// SetPage moves to page p, from 0, kept within the pages.
func (m Model) SetPage(p int) Model {
	m.page = min(max(p, 0), m.Pages()-1)
	return m
}

// Prev moves back a page; on the first it stays.
func (m Model) Prev() Model { return m.SetPage(m.page - 1) }

// Next moves on a page; on the last it stays.
func (m Model) Next() Model { return m.SetPage(m.page + 1) }

// Page is the current page, from 0.
func (m Model) Page() int { return m.page }

// Pages is how many pages the items fill, at least one.
func (m Model) Pages() int {
	per := m.PerPage()
	return max(1, (m.items+per-1)/per)
}

// PerPage is how many items a page holds.
func (m Model) PerPage() int { return max(m.perPage, 1) }

// OnFirst reports whether the current page is the first.
func (m Model) OnFirst() bool { return m.page == 0 }

// OnLast reports whether the current page is the last.
func (m Model) OnLast() bool { return m.page == m.Pages()-1 }

// Bounds is the slice [start, end) of n items the current page shows, both
// within n, so a page left past the end after the items shrank is empty.
func (m Model) Bounds(n int) (start, end int) {
	n = max(n, 0)
	start = min(m.page*m.PerPage(), n)
	return start, min(start+m.PerPage(), n)
}

// ItemsOnPage is how many of n items the current page shows.
func (m Model) ItemsOnPage(n int) int {
	start, end := m.Bounds(n)
	return end - start
}

// Update turns the page on Prev and Next. Paged says the page changed.
func (m Model) Update(msg tea.KeyPressMsg) (Model, Action) {
	before := m.page
	switch {
	case key.Matches(msg, m.keys.Prev):
		m = m.Prev()
	case key.Matches(msg, m.keys.Next):
		m = m.Next()
	}
	if m.page != before {
		return m, Paged
	}
	return m, None
}

// Legend lists the page keys, for the shell's legend or a help overlay.
func (m Model) Legend() []legend.Entry {
	return legend.FromBindings([]key.Binding{m.keys.Prev, m.keys.Next})
}

// View is the page as dots or numbers, in the theme's styles.
func (m Model) View(th theme.Resolved) string {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	pages := m.Pages()
	if m.kind != Dots {
		return st.Numbers.Render(fmt.Sprintf(cmp.Or(m.format, "%d/%d"), m.page+1, pages))
	}
	var b strings.Builder
	if m.page > 0 {
		b.WriteString(st.Inactive.Render(strings.Repeat(m.icons.Pending, m.page)))
	}
	b.WriteString(st.Active.Render(m.icons.Running))
	if after := pages - m.page - 1; after > 0 {
		b.WriteString(st.Inactive.Render(strings.Repeat(m.icons.Pending, after)))
	}
	return b.String()
}
