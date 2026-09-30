// Tree: a directory tree over the app's own node type; arrows expand and
// collapse, and :reveal opens the way down to a path.
package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
	"github.com/lucasassuncao/bezel/tree"
)

// entry is one file or directory. The tree reads it through TreeInfo and
// never sees the path.
type entry struct {
	path     string
	depth    int
	dir      bool
	expanded bool
}

func (e entry) TreeInfo() tree.Info {
	return tree.Info{Depth: e.depth, Leaf: !e.dir, Expanded: e.expanded}
}

func (e entry) Expand(open bool) entry { e.expanded = open; return e }

// entries lists paths depth first; a trailing slash marks a directory.
func entries(paths ...string) []entry {
	out := make([]entry, len(paths))
	for i, p := range paths {
		dir := strings.HasSuffix(p, "/")
		p = strings.TrimSuffix(p, "/")
		out[i] = entry{path: p, depth: strings.Count(p, "/"), dir: dir}
	}
	return out
}

var files = entries(
	"cmd/", "cmd/app/", "cmd/app/main.go",
	"internal/", "internal/store/", "internal/store/store.go", "internal/store/store_test.go",
	"internal/tui/", "internal/tui/model.go", "internal/tui/keys/", "internal/tui/keys/keys.go",
	"docs/", "docs/README.md",
	"go.mod", "Makefile",
)

type revealMsg struct{ path string }

type model struct {
	sh   shell.Shell
	th   theme.Resolved
	tree tree.Model[entry]
}

func actions() []shell.Action {
	return []shell.Action{
		shell.Commands(), shell.Move(),
		shell.Custom("←/→", "collapse/expand", nil, shell.DisplayOnly()),
		shell.Command("reveal", "open the way to a path", func(ctx shell.ActionContext) tea.Cmd {
			return func() tea.Msg { return revealMsg{strings.TrimSpace(ctx.Arg)} }
		}, shell.Arg("path")),
		shell.Quit(),
	}
}

func newModel() model {
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout:  layout.Columns(layout.Fixed("tree", layout.Ratio(1, 2)), layout.Fill("detail")),
		Theme:   th,
		Title:   "tree",
		Actions: actions(),
	})
	return model{sh: sh, th: th, tree: tree.New(files, 0)}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	m.tree.Height = draw.InnerRect(m.sh.Rect("tree")).H
	switch msg := msg.(type) {
	case revealMsg:
		var ok bool
		if m.tree, ok = m.tree.Reveal(func(e entry) bool { return e.path == msg.path }); !ok {
			m.sh, cmd = m.sh.SetStatus("no such path: "+msg.path, shell.Error, 0)
			return m, cmd
		}
		m.tree = m.tree.ClampOffset()
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.tree, _ = m.tree.Update(msg)
	}
	return m, nil
}

func (m model) row(e entry, _ int, selected bool) string {
	icon := "  "
	if e.dir {
		icon = "▸ "
		if e.expanded {
			icon = "▾ "
		}
	}
	line := strings.Repeat("  ", e.depth) + icon + path.Base(e.path)
	if selected {
		return m.th.Cursor.Render(line)
	}
	return line
}

func (m model) detail() string {
	idx := m.tree.CurrentIdx()
	if idx < 0 {
		return ""
	}
	e := m.tree.Nodes[idx]
	kind := "file"
	if e.dir {
		kind = "directory"
	}
	return fmt.Sprintf("%s\n\n%s", m.th.Accent.Render(e.path), m.th.Dim.Render(kind))
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"tree": {Title: "files", Body: func(r layout.Rect) string {
			t := m.tree
			t.Height = r.H
			return t.View(m.th, m.row)
		}},
		"detail": {Title: "detail", Body: func(layout.Rect) string { return m.detail() }},
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
