// The gallery narrowed to one effect by --effect, so a recording can show that
// effect alone, as the GIFs under docs/recordings/effect/solo do.
package main

import (
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/examples/internal/look"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// family is the tab of the gallery a kind belongs to.
type family int

const (
	drawFamily family = iota
	lightFamily
	swapFamily
	retroFamily
)

// kind is one effect the gallery can show alone. i is its place in its tab's
// kinds, or -1 for the ones a tab shows beside them. again is the message r
// plays it with, nil for one that runs until the program ends.
type kind struct {
	name   string
	family family
	i      int
	again  tea.Msg
}

// slug is a kind's name as a flag takes it: "wipe radial" is wipe-radial, and
// a sign like the arrow of "push ←" is dropped.
func slug(name string) string {
	var words []string
	for _, w := range strings.Fields(name) {
		if strings.IndexFunc(w, func(r rune) bool { return !unicode.IsLetter(r) }) < 0 {
			words = append(words, w)
		}
	}
	return strings.Join(words, "-")
}

// kinds is every effect the gallery has, in the order its tabs show them.
func kinds() []kind {
	var ks []kind
	for i, k := range effectKinds {
		ks = append(ks, kind{slug(k.name), drawFamily, i, againMsg{}})
	}
	for i, k := range lightKinds {
		ks = append(ks, kind{slug(k.name), lightFamily, i, shineMsg{}})
	}
	ks = append(ks, kind{"diff", lightFamily, -1, diffMsg{}})
	for i, k := range swapKinds {
		ks = append(ks, kind{slug(k.name), swapFamily, i, againSwap{}})
	}
	for i, k := range retroKinds {
		ks = append(ks, kind{slug(k.name), retroFamily, i, rewindMsg{}})
	}
	return append(ks, kind{"skeleton", retroFamily, -1, nil}, kind{"trace", retroFamily, -1, nil})
}

// solo is the gallery showing one effect, with no tabs.
type solo struct {
	m  model
	sh shell.Shell
	k  kind
}

func newSolo(name string) (solo, error) {
	var found *kind
	names := make([]string, 0, len(effectKinds))
	for _, k := range kinds() {
		names = append(names, k.name)
		if k.name == name {
			k := k
			found = &k
		}
	}
	if found == nil {
		return solo{}, fmt.Errorf("no effect %q; try one of: %s", name, strings.Join(names, ", "))
	}
	m := newModel()
	th := m.sh.Theme()
	switch found.family {
	case drawFamily:
		m.dr.kind = found.i
	case lightFamily:
		if found.i >= 0 {
			m.lt.kind = found.i
			m.lt = m.lt.again(th)
		}
	case swapFamily:
		m.sw.kind = found.i
	case retroFamily:
		if found.i >= 0 {
			m.rt.kind = found.i
			m.rt = m.rt.again(th)
		}
	}
	// The keys run but none is printed: a recording has no use for a legend.
	acts := []shell.Action{shell.Quit(shell.HelpOnly())}
	if found.again != nil {
		acts = []shell.Action{shell.Custom("r", "again", shell.Send(found.again), shell.HelpOnly()), shell.Quit(shell.HelpOnly())}
	}
	sh := shell.New(shell.Config{
		Layout:      layout.Rows(layout.Fill("main")),
		Base:        look.Plain(),
		Title:       "effect",
		Actions:     acts,
		LegendLines: 1,
	})
	return solo{m: m, sh: sh, k: *found}, nil
}

// stage is what the effect plays on, as it shows it now.
func (s solo) stage(th theme.Resolved) string {
	switch s.k.family {
	case drawFamily:
		return s.m.dr.stage(th)
	case lightFamily:
		if s.k.i < 0 {
			return s.m.lt.diffBlock(th)
		}
		return s.m.lt.stage(th)
	case swapFamily:
		return s.m.sw.stage(th)
	}
	switch s.k.name {
	case "skeleton":
		return s.m.rt.shape.Within(2, 1, 26, 3).Apply(report(th))
	case "trace":
		return s.m.rt.trace.Apply(report(th))
	}
	return s.m.rt.stage(th)
}

func (s solo) Init() tea.Cmd { return tea.Batch(s.sh.Init(), s.m.Init()) }

func (s solo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if s.sh, r = s.sh.Update(msg); r.Handled {
		return s, r.Cmd
	}
	next, cmd := s.m.Update(msg)
	s.m = next.(model)
	return s, cmd
}

func (s solo) View() tea.View {
	return s.sh.View(map[string]shell.Pane{
		"main": {Title: s.k.name, Right: still(), Body: func(layout.Rect) string {
			return s.stage(s.sh.Theme())
		}},
	})
}
