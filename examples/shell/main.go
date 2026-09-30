// Shell: two tabs, two panes, a status that expires and a busy spinner.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

type tab struct {
	name  string
	items []string
}

func (t *tab) Name() string                { return t.name }
func (t *tab) Status(shell.Context) string { return fmt.Sprintf("%d items", len(t.items)) }
func (t *tab) Actions(shell.Context) []shell.Action {
	return []shell.Action{
		shell.Custom("s", "save", shell.Send(saveMsg{})),
		shell.Custom("b", "index", shell.Send(indexMsg{})),
	}
}

type (
	saveMsg    struct{}
	indexMsg   struct{}
	indexedMsg struct{}
)

type model struct {
	sh     shell.Shell
	tabs   []*tab
	cursor int
}

func newModel() model {
	tabs := []*tab{
		{name: "files", items: []string{"main.go", "go.mod", "README.md", "Makefile"}},
		{name: "notes", items: []string{"ideas", "todo", "meeting"}},
	}
	shTabs := make([]shell.Tab, len(tabs))
	for i, t := range tabs {
		shTabs[i] = t
	}
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fixed("list", layout.Ratio(1, 3), layout.Min(24)), layout.Fill("detail")),
		Tabs:   shTabs,
		Theme:  theme.Resolve(theme.ThemeMint, true),
		Title:  "shell",
		Actions: []shell.Action{
			shell.Help(), shell.ChangeTab(), shell.ChangePane(), shell.Move(), shell.Quit(),
		},
	})
	return model{sh: sh, tabs: tabs}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	t := m.tabs[m.sh.ActiveTab()]
	switch msg := msg.(type) {
	case saveMsg:
		m.sh, cmd = m.sh.SetStatus("saved "+t.items[m.cursor], shell.OK, 2*time.Second)
		return m, cmd
	case indexMsg:
		m.sh, cmd = m.sh.Busy("indexing " + t.name)
		return m, tea.Batch(cmd, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return indexedMsg{} }))
	case indexedMsg:
		m.sh = m.sh.Idle()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			m.cursor = max(0, m.cursor-1)
		case "down":
			m.cursor = min(len(t.items)-1, m.cursor+1)
		}
		m.cursor = min(m.cursor, len(t.items)-1)
	}
	return m, nil
}

func (m model) View() tea.View {
	t := m.tabs[m.sh.ActiveTab()]
	cursor := min(m.cursor, len(t.items)-1)
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"list": {Title: t.name, Body: func(layout.Rect) string {
			var b strings.Builder
			for i, it := range t.items {
				mark := "  "
				if i == cursor {
					mark = "> "
				}
				b.WriteString(mark + it + "\n")
			}
			return b.String()
		}},
		"detail": {Title: "detail", Body: func(r layout.Rect) string {
			return fmt.Sprintf("selected: %s\nfocus: %s\npane: %dx%d", t.items[cursor], m.sh.Focus(), r.W, r.H)
		}},
	}))
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
