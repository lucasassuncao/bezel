// Breadcrumb: a store browsed by path. enter on a folder goes into it; tab
// moves to the path, where the arrows pick a part and enter (or a second
// click) goes back up to it. A narrow window gives the path's middle way.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/breadcrumb"
	"github.com/lucasassuncao/bezel/component/list"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// store is each folder's entries by its path; an entry with no path of its
// own is a secret.
var store = map[string][]string{
	"kv":                                  {"apps", "infra"},
	"kv/apps":                             {"billing", "search"},
	"kv/apps/billing":                     {"production", "staging"},
	"kv/apps/billing/production":          {"database", "stripe-key"},
	"kv/apps/billing/production/database": {"password", "username"},
	"kv/apps/billing/staging":             {"database"},
	"kv/apps/search":                      {"index-token"},
	"kv/infra":                            {"dns-token", "tls-cert"},
}

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	crumbs breadcrumb.Model
	items  list.Model[string]
	opened string // the last secret enter opened
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fixed("path", layout.Lines(3)), layout.Fill("entries")),
		Theme:  th,
		Title:  "breadcrumb",
		Mouse:  true,
		Actions: []shell.Action{
			shell.ChangePane(),
			shell.Move(),
			shell.Custom("←/→", "part", nil, shell.DisplayOnly()),
			shell.Custom("enter", "open", nil, shell.DisplayOnly()),
			shell.Quit(),
		},
	})
	sh = sh.SetFocus("entries")
	m := model{sh: sh, th: th, crumbs: breadcrumb.New(nil).Blur(), items: list.New[string](nil)}
	return m.enter([]string{"kv"})
}

// enter shows the folder at path: the crumbs end there and the list holds
// its entries.
func (m model) enter(path []string) model {
	m.crumbs = m.crumbs.SetParts(path)
	at := strings.Join(path, "/")
	var rows []list.Row[string]
	for _, name := range store[at] {
		label := name
		if _, folder := store[at+"/"+name]; folder {
			label += "/"
		}
		rows = append(rows, list.Row[string]{Label: label, Mark: "•", Value: name})
	}
	m.items = m.items.SetRows(rows)
	m.opened = ""
	return m.resize()
}

// resize fits the crumbs and the list to their panes.
func (m model) resize() model {
	m.crumbs = m.crumbs.SetWidth(draw.InnerRect(m.sh.Rect("path")).W)
	m.items = m.items.SetHeight(draw.InnerRect(m.sh.Rect("entries")).H)
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	m.sh, r = m.sh.Update(msg)
	if m = m.focus(); r.Handled {
		return m, r.Cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.resize(), nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.sh.Focused() == "path" {
			return m.toCrumbs(msg), nil
		}
	case shell.ClickMsg:
		if msg.Pane == "path" {
			return m.toCrumbs(msg), nil
		}
	}
	return m.toItems(msg), nil
}

// focus shows the crumbs' cursor only while their pane has the keys.
func (m model) focus() model {
	if m.sh.Focused() == "path" {
		m.crumbs = m.crumbs.Focus()
	} else {
		m.crumbs = m.crumbs.Blur()
	}
	return m
}

// toCrumbs hands msg to the path; a part chosen goes back up to it.
func (m model) toCrumbs(msg tea.Msg) model {
	var r bezel.Result
	m.crumbs, r = m.crumbs.Update(msg)
	if c, ok := r.Event.(breadcrumb.Chosen); ok {
		m = m.enter(c.Path)
		m.sh = m.sh.SetFocus("entries")
		m = m.focus()
	}
	return m
}

// toItems hands msg to the list; a folder chosen opens, a secret is shown.
func (m model) toItems(msg tea.Msg) model {
	var r bezel.Result
	m.items, r = m.items.Update(msg)
	if c, ok := r.Event.(list.Chosen[string]); ok {
		path := append(m.crumbs.Parts(), c.Row.Value)
		if _, folder := store[strings.Join(path, "/")]; folder {
			return m.enter(path)
		}
		m.opened = strings.Join(path, "/")
	}
	return m
}

// entries is the list, and the secret enter opened under it.
func (m model) entries(r layout.Rect) string {
	out := m.items.View(m.th)
	if m.opened != "" {
		out += "\n\n" + m.th.Dim.Render("opened  ") + m.th.Accent.Render(m.opened)
	}
	return draw.FitBlock(out, r.W, r.H)
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"path":    {Title: "path", Body: func(layout.Rect) string { return m.crumbs.View(m.th) }},
		"entries": {Title: "entries", Body: m.entries},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
