// Shell: two tabs, two panes, a status that expires, a busy spinner and a
// confirm. The mouse works too: click a tab, a pane, a row (twice to open
// it) or a dialog's button.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/list"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

type tab struct {
	name  string
	items []string
}

func (t tab) Name() string                { return t.name }
func (t tab) Status(shell.Context) string { return fmt.Sprintf("%d items", len(t.items)) }
func (t tab) Actions(shell.Context) []shell.Action {
	return []shell.Action{
		shell.Custom("s", "save", shell.Send(saveMsg{})),
		shell.Custom("b", "index", shell.Send(indexMsg{})),
		shell.Custom("d", "delete", shell.Send(askMsg{})),
	}
}

type (
	saveMsg    struct{}
	indexMsg   struct{}
	indexedMsg struct{}
	askMsg     struct{}
	deletedMsg struct{ item string }
)

type model struct {
	sh    shell.Shell
	th    theme.Resolved
	tabs  []tab
	lists []list.Model[int] // one a tab, so each keeps its cursor
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	tabs := []tab{
		{name: "files", items: []string{"main.go", "go.mod", "README.md", "Makefile"}},
		{name: "notes", items: []string{"ideas", "todo", "meeting"}},
	}
	shTabs := make([]shell.Tab, len(tabs))
	lists := make([]list.Model[int], len(tabs))
	for i, t := range tabs {
		shTabs[i] = t
		rows := make([]list.Row[int], len(t.items))
		for j, it := range t.items {
			rows[j] = list.Row[int]{Label: it, Value: j}
		}
		lists[i] = list.New(rows)
	}
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fixed("list", layout.Ratio(1, 3), layout.Min(24)), layout.Fill("detail")),
		Tabs:   shTabs,
		Theme:  th,
		Title:  "shell",
		Mouse:  true, // Screen turns it on; tabs, panes and dialogs then take clicks
		Actions: []shell.Action{
			shell.Help(), shell.ChangeTab(), shell.Move(), shell.Quit(),
		},
	})
	return model{sh: sh, th: th, tabs: tabs, lists: lists}
}

func (m model) Init() tea.Cmd { return nil }

// selected is the item under the active tab's cursor.
func (m model) selected() string {
	if r, ok := m.lists[m.sh.ActiveTab()].Selected(); ok {
		return r.Label
	}
	return ""
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	at := m.sh.ActiveTab()
	m.lists[at] = m.lists[at].SetHeight(draw.InnerRect(m.sh.Rect("list")).H)
	switch msg := msg.(type) {
	case saveMsg:
		m.sh, cmd = m.sh.SetStatus("saved "+m.selected(), shell.OK, 2*time.Second)
		return m, cmd
	case indexMsg:
		m.sh, cmd = m.sh.Busy("indexing " + m.tabs[at].name)
		return m, tea.Batch(cmd, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return indexedMsg{} }))
	case indexedMsg:
		m.sh = m.sh.Idle()
	case askMsg:
		item := m.selected()
		yes := func() tea.Msg { return deletedMsg{item} }
		return m, overlay.Open(overlay.NewConfirm("Delete", item, yes, m.th).WithKind(overlay.Danger))
	case deletedMsg:
		m.sh, cmd = m.sh.SetStatus("deleted "+msg.item, shell.OK, 2*time.Second)
		return m, cmd
	case shell.ClickMsg:
		if msg.Pane != "list" {
			return m, nil
		}
		var r bezel.Result
		click := bezel.ClickMsg{Pane: msg.Pane, MouseClickMsg: msg.MouseClickMsg}
		if m.lists[at], r = m.lists[at].Update(click); r.Event != nil {
			m.sh, cmd = m.sh.SetStatus("opened "+m.selected(), shell.OK, 2*time.Second)
			return m, cmd
		}
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.lists[at], _ = m.lists[at].Update(msg)
	}
	return m, nil
}

func (m model) View() tea.View {
	at := m.sh.ActiveTab()
	l := m.lists[at].SetHeight(draw.InnerRect(m.sh.Rect("list")).H)
	return m.sh.Screen(map[string]shell.Pane{
		"list": {Title: m.tabs[at].name, Body: func(layout.Rect) string { return l.View(m.th) }},
		"detail": {Title: "detail", Body: func(r layout.Rect) string {
			return fmt.Sprintf("selected: %s\nfocus: %s\npane: %dx%d", m.selected(), m.sh.Focused(), r.W, r.H)
		}},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
