// Demo: two tabs, list + detail, a confirm overlay and a status banner.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

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

func (t *tab) Name() string { return t.name }

func (t *tab) Status(shell.Context) string { return fmt.Sprintf("%d items", len(t.items)) }

// Actions are this tab's own keys; help, tab, the arrows and quit are global.
func (t *tab) Actions(ctx shell.Context) []shell.Action {
	if ctx.Focus == "list" {
		return []shell.Action{
			shell.Custom("enter", "open", shell.Send(openMsg{})),
			shell.Custom("d", "delete", shell.Send(askDeleteMsg{}), shell.Needs("write")),
		}
	}
	return []shell.Action{shell.Custom("esc", "back", shell.Send(backMsg{}))}
}

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	tabs   []*tab
	cursor int
}

type (
	openMsg      struct{}
	backMsg      struct{}
	askDeleteMsg struct{}
	deletedMsg   struct{}
)

func newModel() model {
	tabs := []*tab{
		{name: "fruits", items: []string{"banana", "mint", "strawberry", "blueberry", "mango"}},
		{name: "tools", items: []string{"hammer", "wrench", "screwdriver"}},
	}
	shTabs := make([]shell.Tab, len(tabs))
	for i, t := range tabs {
		shTabs[i] = t
	}
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout:  layout.Columns(layout.Fixed("list", layout.Ratio(1, 3), layout.Min(26), layout.Max(52)), layout.Fill("detail")).Collapse(72, "detail"),
		Tabs:    shTabs,
		Theme:   th,
		Title:   "bezel demo",
		Version: "v0.1.0",
		Can:     func(c shell.Capability) bool { return c != "write" || os.Getenv("DEMO_RW") != "" },
		Actions: []shell.Action{shell.Help(), shell.ChangeTab(), shell.Move(), shell.Quit()},
	})
	return model{sh: sh, th: th, tabs: tabs}
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
	case openMsg:
		m.sh = m.sh.SetFocus("detail")
	case backMsg:
		m.sh = m.sh.SetFocus("list")
	case askDeleteMsg:
		if len(t.items) > 0 {
			m.sh = m.sh.Push(overlay.NewConfirm("Delete "+t.items[m.cursor]+"?", "This only affects the demo.",
				func() tea.Msg { return deletedMsg{} }, m.th.Modal, m.th.Legend))
		}
	case deletedMsg:
		t.items = append(t.items[:m.cursor], t.items[m.cursor+1:]...)
		m.cursor = max(0, min(m.cursor, len(t.items)-1))
		m.sh, cmd = m.sh.SetStatus("deleted", shell.OK, 2*time.Second)
		return m, cmd
	case tea.KeyPressMsg:
		// Move is display only: the arrows are the list's, here the model's.
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			m.cursor = max(0, m.cursor-1)
		case "down":
			m.cursor = min(len(t.items)-1, m.cursor+1)
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	t := m.tabs[m.sh.ActiveTab()]
	panes := map[string]shell.Pane{
		"list": {Title: t.name, Body: func(r layout.Rect) string {
			var b strings.Builder
			for i, it := range t.items {
				mark := "  "
				if i == m.cursor {
					mark = "> "
				}
				b.WriteString(mark + it + "\n")
			}
			return b.String()
		}},
		"detail": {Title: "detail", Body: func(r layout.Rect) string {
			if len(t.items) == 0 {
				return draw.EmptyState(r, "nothing here", "switch tab", m.th.Chrome)
			}
			return fmt.Sprintf("selected: %s\n\npane is %dx%d", t.items[m.cursor], r.W, r.H)
		}},
	}
	v := tea.NewView(m.sh.View(panes))
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
