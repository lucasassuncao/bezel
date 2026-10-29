// Accordion: a column of sections that open and close, sliding. The arrows
// move between headings and enter or space opens the one under the cursor.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/component/accordion"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// faq is the questions the accordion holds.
var faq = []accordion.Section{
	{Title: "What is bezel?", Body: "The frame around a Bubble Tea app: tabs, panes, dialogs, the status row and the key legend. The app draws what goes inside."},
	{Title: "Does it need a mouse?", Body: "No. Every action has a key, listed at the foot of the screen and in the help."},
	{Title: "Can I turn the motion off?", Body: "Set BEZEL_NO_MOTION, or TERM=dumb, and everything lands where it would end. BEZEL_MOTION sets the speed instead."},
	{Title: "Which themes are there?", Body: "Ninety-two presets, from mint to dracula, and any theme of your own built on one."},
}

type model struct {
	sh  shell.Shell
	th  theme.Resolved
	acc accordion.Model
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	acc := accordion.New(faq)
	sh := shell.New(shell.Config{
		Layout:  layout.Columns(layout.Fill("faq")),
		Theme:   th,
		Title:   "accordion",
		Actions: []shell.Action{shell.Custom("↑/↓", "move", nil, shell.DisplayOnly()), shell.Custom("enter", "open/close", nil, shell.DisplayOnly()), shell.Quit()},
	})
	return model{sh: sh, th: th, acc: acc}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "ctrl+c" {
		return m, tea.Quit
	}
	m.acc, cmd = m.acc.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"faq": {Title: "faq", Body: func(r layout.Rect) string { return m.acc.SetWidth(r.W).View(m.th) }},
	}))
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
