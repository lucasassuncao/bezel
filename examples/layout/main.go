// Layout: a header row over two columns; w narrows the window and the side
// column collapses away.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// narrowWidth is below the Collapse limit of 72.
const narrowWidth = 60

type toggleWidthMsg struct{}

type model struct {
	sh           shell.Shell
	realW, realH int
	narrow       bool
}

func newModel() model {
	sh := shell.New(shell.Config{
		Layout: layout.Rows(
			layout.Fixed("header", layout.Lines(3)),
			layout.Columns(
				layout.Fixed("side", layout.Ratio(1, 3), layout.Min(20)),
				layout.Fill("main"),
			).Collapse(72, "main"),
		),
		Theme: theme.Resolve(theme.ThemeMint, true),
		Title: "layout",
		Actions: []shell.Action{
			shell.Custom("w", "narrow/wide", shell.Send(toggleWidthMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh}
}

func (m model) Init() tea.Cmd { return nil }

// resize hands the shell the width to draw at: the real one or narrowWidth.
// VHS cannot resize a terminal mid-recording, so this stands in for it.
func (m model) resize() model {
	w := m.realW
	if m.narrow {
		w = min(narrowWidth, m.realW)
	}
	m.sh, _, _ = m.sh.Update(tea.WindowSizeMsg{Width: w, Height: m.realH})
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.realW, m.realH = size.Width, size.Height
		return m.resize(), nil
	}
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case toggleWidthMsg:
		m.narrow = !m.narrow
		return m.resize(), nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// rectBody prints the leaf's own name and the rect the layout gave it.
func rectBody(name string) func(layout.Rect) string {
	return func(r layout.Rect) string { return fmt.Sprintf("%s %dx%d", name, r.W, r.H) }
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"header": {Title: "header", Body: rectBody("header")},
		"side":   {Title: "side", Body: rectBody("side")},
		"main":   {Title: "main", Body: rectBody("main")},
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
