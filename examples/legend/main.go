// Legend: more keys than one row holds, some needing a write capability
// that w grants; the ones cut are counted and ? lists every one.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

const write shell.Capability = "write"

type (
	toggleWriteMsg struct{}
	didMsg         struct{ what string }
)

type model struct {
	sh shell.Shell
	// canWrite is shared with Config.Can, which reads it on every frame.
	canWrite *bool
}

// did is an action that only reports itself: the keys are what this shows.
func did(what string) func(shell.ActionContext) tea.Cmd { return shell.Send(didMsg{what}) }

// actions is every key the app has; a test runs shell.Check over it.
func actions() []shell.Action {
	return []shell.Action{
		shell.Help(), shell.Move(), shell.Quit(),
		shell.Custom("w", "write mode", shell.Send(toggleWriteMsg{})),
		shell.Custom("o", "open", did("open")),
		shell.Custom("c", "copy", did("copy")),
		shell.Custom("/", "filter", did("filter")),
		shell.Custom("s", "sort", did("sort")),
		shell.Custom("e", "export", did("export")),
		shell.Custom("n", "new", did("new"), shell.Needs(write)),
		shell.Custom("r", "rename", did("rename"), shell.Needs(write)),
		shell.Custom("d", "delete", did("delete"), shell.Needs(write)),
		shell.Custom("i", "import", did("import"), shell.Needs(write)),
		shell.Custom("u", "undo", did("undo"), shell.Needs(write)),
	}
}

func newModel() model {
	canWrite := new(bool)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("main")),
		Theme:  theme.Resolve(theme.ThemeMint, true),
		Title:  "legend",
		Can:    func(c shell.Capability) bool { return c != write || *canWrite },
		// One row, so the cut and its count show even on a wide terminal.
		LegendLines: 1,
		Actions:     actions(),
	})
	return model{sh: sh, canWrite: canWrite}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case toggleWriteMsg:
		*m.canWrite = !*m.canWrite
	case didMsg:
		m.sh, cmd = m.sh.SetStatus(msg.what, shell.Info, 2*time.Second)
		return m, cmd
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	mode := "read-only: the write keys are hidden"
	if *m.canWrite {
		mode = "write mode: every key is shown"
	}
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"main": {Title: "legend", Body: func(layout.Rect) string { return mode }},
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
