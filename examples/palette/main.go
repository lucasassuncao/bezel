// Palette: ":" opens the command line; commands take arguments, complete
// with tab, and an unknown or refused one says so on the status row.
package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

var items = []string{"apple", "banana", "cherry", "grape", "lemon", "mango", "peach", "plum"}

type (
	gotoMsg  struct{ name string }
	themeMsg struct{ name string }
	clearMsg struct{}
)

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	cursor int
}

// actions is every key and command; delete needs a capability the session
// never grants, so the palette knows it but refuses it.
func actions() []shell.Action {
	return []shell.Action{
		shell.Commands(), shell.Move(), shell.Quit(),
		shell.Command("goto", "move to an item", func(ctx shell.ActionContext) tea.Cmd {
			return func() tea.Msg { return gotoMsg{ctx.Arg} }
		}, shell.Arg("item")),
		shell.Command("theme", "switch the theme", func(ctx shell.ActionContext) tea.Cmd {
			return func() tea.Msg { return themeMsg{ctx.Arg} }
		}, shell.Arg("name")),
		shell.Command("clear", "move back to the top", shell.Send(clearMsg{})),
		shell.Command("delete", "delete the item", shell.Send(clearMsg{}), shell.Needs("write")),
	}
}

func newModel() model {
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout:  layout.Columns(layout.Fill("items")),
		Theme:   th,
		Title:   "palette",
		Can:     func(c shell.Capability) bool { return c != "write" },
		Actions: actions(),
	})
	return model{sh: sh, th: th}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case gotoMsg:
		i := slices.Index(items, strings.TrimSpace(msg.name))
		if i < 0 {
			m.sh, cmd = m.sh.SetStatus("no item "+msg.name, shell.Error, 0)
			return m, cmd
		}
		m.cursor = i
	case themeMsg:
		t, err := theme.Lookup(strings.TrimSpace(msg.name))
		if err != nil {
			m.sh, cmd = m.sh.SetStatus(err.Error(), shell.Error, 0)
			return m, cmd
		}
		m.th = theme.Resolve(t, true)
		m.sh = m.sh.SetTheme(m.th)
	case clearMsg:
		m.cursor = 0
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			m.cursor = max(0, m.cursor-1)
		case "down":
			m.cursor = min(len(items)-1, m.cursor+1)
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"items": {Title: "items", Body: func(layout.Rect) string {
			var b strings.Builder
			for i, it := range items {
				if i == m.cursor {
					b.WriteString(m.th.Cursor.Render("> "+it) + "\n")
					continue
				}
				b.WriteString("  " + it + "\n")
			}
			return b.String()
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
