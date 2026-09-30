// Animation: h slides a hints pane open and shut; a tween gives its height
// each frame and SetLayout applies it.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// hintsHeight is the open pane, borders included.
const hintsHeight = 8

const slide = 300 * time.Millisecond

type (
	toggleMsg struct{}
	frameMsg  time.Time
)

type model struct {
	sh    shell.Shell
	open  bool
	tween animation.Tween
	// height is what the layout holds now: the tween's last sample.
	height int
}

// shape is the layout with the hints pane at h rows; at 0 it is left out.
func shape(h int) layout.Node {
	if h == 0 {
		return layout.Rows(layout.Fill("main"))
	}
	return layout.Rows(layout.Fill("main"), layout.Fixed("hints", layout.Lines(h)))
}

func newModel() model {
	sh := shell.New(shell.Config{
		Layout: shape(0),
		Theme:  theme.Resolve(theme.ThemeMint, true),
		Title:  "animation",
		Actions: []shell.Action{
			shell.Custom("h", "slide open/shut", shell.Send(toggleMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh}
}

func nextFrame() tea.Cmd {
	return tea.Tick(animation.Frame, func(t time.Time) tea.Msg { return frameMsg(t) })
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case toggleMsg:
		m.open = !m.open
		target := 0
		if m.open {
			target = hintsHeight
		}
		// A toggle mid-slide starts from where the pane is, not where it was going.
		m.tween = animation.New(m.height, target, slide)
		return m, nextFrame()
	case frameMsg:
		done := m.tween.Advance(time.Time(msg))
		m.height = m.tween.Cur
		m.sh = m.sh.SetLayout(shape(m.height))
		if !done {
			return m, nextFrame()
		}
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"main": {Title: "main", Body: func(layout.Rect) string { return "press h" }},
		"hints": {Title: "hints", Body: func(layout.Rect) string {
			return "h      toggle this pane\nq      quit\n\nIt slides because a Tween\neases its height, frame by frame."
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
