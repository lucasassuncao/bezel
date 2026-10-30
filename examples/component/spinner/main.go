// Spinner: the twelve presets turning side by side, each at its own pace.
// space stops them all, and starts them again.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/component/spinner"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// toggleMsg stops every spinner, or starts them when they are stopped.
type toggleMsg struct{}

var presets = []struct {
	name string
	s    spinner.Spinner
}{
	{"Line", spinner.Line()}, {"Dot", spinner.Dot()}, {"MiniDot", spinner.MiniDot()},
	{"Jump", spinner.Jump()}, {"Pulse", spinner.Pulse()}, {"Points", spinner.Points()},
	{"Globe", spinner.Globe()}, {"Moon", spinner.Moon()}, {"Monkey", spinner.Monkey()},
	{"Meter", spinner.Meter()}, {"Hamburger", spinner.Hamburger()}, {"Ellipsis", spinner.Ellipsis()},
}

type model struct {
	sh       shell.Shell
	th       theme.Resolved
	spinners [12]spinner.Model
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("spinner")),
		Theme:  th,
		Title:  "spinner",
		Actions: []shell.Action{
			shell.Custom("space", "stop/start", shell.Send(toggleMsg{})),
			shell.Quit(),
		},
	})
	m := model{sh: sh, th: th}
	for i, p := range presets {
		m.spinners[i] = spinner.New(p.s)
	}
	return m
}

// Init starts the spinners the way space does.
func (m model) Init() tea.Cmd { return func() tea.Msg { return toggleMsg{} } }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case animation.FrameMsg:
		for i := range m.spinners {
			if m.spinners[i].Owns(msg) {
				var r bezel.Result
				m.spinners[i], r = m.spinners[i].Update(msg)
				return m, r.Cmd
			}
		}
	case toggleMsg:
		var cmds []tea.Cmd
		running := m.spinners[0].Running()
		for i := range m.spinners {
			if running {
				m.spinners[i] = m.spinners[i].Stop()
				continue
			}
			m.spinners[i], cmd = m.spinners[i].Start()
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) body(r layout.Rect) string {
	out := []string{draw.Heading("Presets", m.th.Section, r.W), ""}
	for i, p := range presets {
		out = append(out, "  "+fmt.Sprintf("%-11s", p.name)+m.spinners[i].View(m.th))
	}
	return strings.Join(out, "\n")
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"spinner": {Title: "spinner", Body: m.body},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
