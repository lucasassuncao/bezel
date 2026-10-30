// Toast: notices that slide into the top right corner, stack, and leave on
// their own. s, c, w and e push one of each kind.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/toast"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

type pushMsg struct {
	text string
	kind toast.Kind
}

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	toasts toast.Model
}

func push(text string, kind toast.Kind) func(shell.ActionContext) tea.Cmd {
	return shell.Send(pushMsg{text: text, kind: kind})
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("doc")),
		Theme:  th,
		Title:  "toast",
		Actions: []shell.Action{
			shell.Custom("s", "saved", push("Saved notes.md", toast.Success)),
			shell.Custom("c", "copied", push("Link copied to the clipboard", toast.Info)),
			shell.Custom("w", "warning", push("Disk is 90% full", toast.Warning)),
			shell.Custom("e", "error", push("Could not reach the server; retrying in 5s", toast.Error)),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, toasts: toast.New().WithWidth(34)}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case pushMsg:
		m.toasts, cmd = m.toasts.Push(msg.text, msg.kind)
		return m, cmd
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	m.toasts, r = m.toasts.Update(msg)
	return m, r.Cmd
}

func (m model) body(r layout.Rect) string {
	return strings.Join([]string{
		draw.Heading("notes.md", m.th.Section, r.W),
		"",
		"  A toast says what happened without taking the keys:",
		"  it slides in at the top right, stays three seconds",
		"  and slides away. Several stack, the newest on top.",
	}, "\n")
}

func (m model) View() tea.View {
	screen := m.sh.View(map[string]shell.Pane{"doc": {Title: "doc", Body: m.body}})
	v := tea.NewView(m.toasts.Over(screen, m.th))
	v.AltScreen = true
	v.BackgroundColor, v.ForegroundColor = m.th.Background, m.th.Foreground
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
