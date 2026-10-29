// Scene: every scene the animation/scene package has, one at a time, filling
// the screen with a title over it. n plays the next; each keeps its own ticks.
package main

import (
	"fmt"
	"image/color"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// black is the gallery's dark background, what the scenes fade towards.
var black = color.RGBA{0, 0, 0, 255}

type model struct {
	sh shell.Shell
	th theme.Resolved
	sc scenes
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fill("main")),
		Theme:  th,
		Title:  "scene",
		Actions: []shell.Action{
			shell.Custom("n", "next scene", shell.Send(sceneMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, sc: newScenes(th)}
}

// Init sets the first scene going.
func (m model) Init() tea.Cmd { return m.sc.scene.Tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case sceneMsg:
		m.sc, cmd = m.sc.next(m.th)
		return m, cmd
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	m.sc, cmd = m.sc.update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"main": {Title: "every scene", Right: still(), Body: func(r layout.Rect) string {
			return m.sc.rows(m.th, m.label, r.W, r.H)
		}},
	}))
	v.AltScreen = true
	v.BackgroundColor, v.ForegroundColor = m.th.Background, m.th.Foreground
	return v
}

// still names the switch when motion is off, so a frozen scene says why.
func still() string {
	if animation.Still() {
		return "motion off"
	}
	return ""
}

// label is a row's kind and variant, in two fixed columns.
func (m model) label(kind, variant string) string {
	return m.th.Accent.Render(fmt.Sprintf("%-11s", kind)) + m.th.Muted.Render(fmt.Sprintf("%-14s", variant))
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
