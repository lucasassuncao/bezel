// The gallery narrowed to one scene by --scene, so a recording can show that
// scene alone, as the GIFs under docs/recordings/animation/scene/solo do.
package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/examples/internal/look"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
)

// solo is one scene filling the pane, with no way to change it.
type solo struct {
	sh shell.Shell
	sc scenes
}

func newSolo(name string) (solo, error) {
	// The key runs but none is printed: a recording has no use for a legend.
	sh := shell.New(shell.Config{
		Layout:      layout.Rows(layout.Fill("main")),
		Base:        look.Plain(),
		Title:       "scene",
		Actions:     []shell.Action{shell.Quit(shell.HelpOnly())},
		LegendLines: 1,
	})
	sc := newScenes(sh.Theme())
	for i, k := range sceneKinds {
		if k.name == name {
			sc.kind, sc.scene = i, k.make(sh.Theme())
			return solo{sh: sh, sc: sc}, nil
		}
	}
	names := make([]string, len(sceneKinds))
	for i, k := range sceneKinds {
		names[i] = k.name
	}
	return solo{}, fmt.Errorf("no scene %q; try one of: %s", name, strings.Join(names, ", "))
}

func (s solo) Init() tea.Cmd { return tea.Batch(s.sh.Init(), s.sc.scene.Tick()) }

func (s solo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if s.sh, r = s.sh.Update(msg); r.Handled {
		return s, r.Cmd
	}
	var cmd tea.Cmd
	s.sc, cmd = s.sc.update(msg)
	return s, cmd
}

func (s solo) View() tea.View {
	return s.sh.View(map[string]shell.Pane{
		"main": {Title: sceneKinds[s.sc.kind].name, Right: still(), Body: func(r layout.Rect) string {
			return s.sc.under(s.sh.Theme(), r.W, r.H)
		}},
	})
}
