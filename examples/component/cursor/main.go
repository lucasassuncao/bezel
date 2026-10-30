// Cursor: three cursors, one per mode, each on the first letter of a word.
// tab moves the focus down the rows and m changes the focused cursor's mode:
// blinking, steady, hidden and around again.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/cursor"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

type (
	nextMsg struct{}
	modeMsg struct{}
)

type model struct {
	sh      shell.Shell
	th      theme.Resolved
	cursors [3]cursor.Model
	focus   int
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("cursor")),
		Theme:  th,
		Title:  "cursor",
		Actions: []shell.Action{
			shell.Custom("tab", "next cursor", shell.Send(nextMsg{})),
			shell.Custom("m", "change mode", shell.Send(modeMsg{})),
			shell.Quit(),
		},
	})
	m := model{sh: sh, th: th}
	for i, mode := range []cursor.Mode{cursor.Blinking, cursor.Steady, cursor.Hidden} {
		m.cursors[i], _ = cursor.New(th).SetChar("h").SetMode(mode)
	}
	m.cursors[0], _ = m.cursors[0].Focus()
	return m
}

// Init starts the blink of the cursor newModel focused.
func (m model) Init() tea.Cmd { return m.cursors[m.focus].Blink() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case cursor.BlinkMsg:
		var cmds []tea.Cmd
		for i := range m.cursors {
			var r bezel.Result
			m.cursors[i], r = m.cursors[i].Update(msg)
			cmds = append(cmds, r.Cmd)
		}
		return m, tea.Batch(cmds...)
	case nextMsg:
		m.cursors[m.focus] = m.cursors[m.focus].Blur()
		m.focus = (m.focus + 1) % len(m.cursors)
		m.cursors[m.focus], cmd = m.cursors[m.focus].Focus()
		return m, cmd
	case modeMsg:
		c := m.cursors[m.focus]
		m.cursors[m.focus], cmd = c.SetMode((c.Mode() + 1) % 3)
		return m, cmd
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) body(r layout.Rect) string {
	out := []string{draw.Heading("Modes", m.th.Section, r.W), ""}
	for i, c := range m.cursors {
		mark := "  "
		if i == m.focus {
			mark = m.th.Accent.Render("▸ ")
		}
		out = append(out, " "+mark+fmt.Sprintf("%-10s", c.Mode())+c.View()+"ello")
	}
	return strings.Join(out, "\n")
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"cursor": {Title: "cursor", Body: m.body},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
