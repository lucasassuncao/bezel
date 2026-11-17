// Scene: every scene the animation/scene package has, one at a time, filling
// the screen with a title over it. n plays the next; each keeps its own ticks.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/examples/internal/look"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
)

// black is the gallery's dark background, what the scenes fade towards.
var black = color.RGBA{0, 0, 0, 255}

type model struct {
	sh shell.Shell
	sc scenes
}

func newModel() model {
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fill("main")),
		Base:   look.Plain(),
		Title:  "scene",
		Actions: []shell.Action{
			shell.Custom("n", "next scene", shell.Send(sceneMsg{})),
			shell.Quit(),
		},
	})
	th := sh.Theme()
	return model{sh: sh, sc: newScenes(th)}
}

// Init sets the first scene going.
func (m model) Init() tea.Cmd { return tea.Batch(m.sh.Init(), m.sc.scene.Tick()) }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	var cmd tea.Cmd
	if _, ok := msg.(sceneMsg); ok {
		m.sc, cmd = m.sc.next(m.sh.Theme())
		return m, cmd
	}
	m.sc, cmd = m.sc.update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"main": {Title: "every scene", Right: still(), Body: func(r layout.Rect) string {
			return m.sc.rows(m.sh.Theme(), m.label, r.W, r.H)
		}},
	})
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
	return m.sh.Theme().Accent.Render(fmt.Sprintf("%-11s", kind)) + m.sh.Theme().Muted.Render(fmt.Sprintf("%-14s", variant))
}

func main() {
	name := flag.String("scene", "", "show only this scene, as a recording does")
	flag.Parse()
	var m tea.Model = newModel()
	if *name != "" {
		s, err := newSolo(*name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		m = s
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
