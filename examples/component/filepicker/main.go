// Filepicker: the repository browsed inside a pane. enter opens a directory
// or picks a file and names it on the status row, ← climbs, and p shows or
// hides the permissions column.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/filepicker"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// permsMsg shows or hides the permissions column.
type permsMsg struct{}

type model struct {
	sh    shell.Shell
	th    theme.Resolved
	fp    filepicker.Model
	perms bool
	root  string
}

func newModel(dir string) model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("files")),
		Theme:  th,
		Title:  "filepicker",
		Actions: []shell.Action{
			shell.Custom("↑/↓", "move", nil, shell.DisplayOnly()),
			shell.Custom("←/→", "up/open", nil, shell.DisplayOnly()),
			shell.Custom("enter", "pick", nil, shell.DisplayOnly()),
			shell.Custom("p", "permissions", shell.Send(permsMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, fp: filepicker.New(dir), root: dir}
}

func (m model) Init() tea.Cmd { return m.fp.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.fp = m.fp.SetHeight(size.Height - 7) // the pane's rows, less the chrome around it
	}
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case permsMsg:
		m.perms = !m.perms
		m.fp = m.fp.WithPermissions(m.perms)
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	var r bezel.Result
	m.fp, r = m.fp.Update(msg)
	cmd = r.Cmd
	if path, ok := m.fp.Picked(); ok {
		if rel, err := filepath.Rel(m.root, path); err == nil {
			path = rel
		}
		var status tea.Cmd
		m.sh, status = m.sh.SetStatus("picked "+path, shell.OK, 3*time.Second)
		return m, tea.Batch(cmd, status)
	}
	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"files": {Title: m.fp.Dir(), Body: func(r layout.Rect) string { return m.fp.SetHeight(r.H).View(m.th) }},
	}))
	v.AltScreen = true
	v.BackgroundColor, v.ForegroundColor = m.th.Background, m.th.Foreground
	return v
}

func main() {
	dir, err := os.Getwd()
	if err == nil {
		_, err = tea.NewProgram(newModel(dir)).Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
