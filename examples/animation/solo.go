// The gallery narrowed to one animation by --animation, so a recording can
// show that animation alone, as the GIFs under docs/recordings/animation/solo do.
package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/examples/internal/look"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// tabs is how many families the gallery has.
const tabs = 7

// replay is the message that plays a kind again; a kind missing here loops,
// or, as the spring does, answers the arrow keys.
var replay = map[string]tea.Msg{
	"easing": replayMsg{}, "float": replayMsg{}, "count-up": replayMsg{}, "cascade": replayMsg{},
	"fade": fadeMsg{}, "pulse": pulseMsg{}, "curves": curvesMsg{},
	"point": moveMsg{}, "path": routeMsg{},
	"drop": dropMsg{}, "flock": flockMsg{}, "rubber": rubberMsg{},
	"bars": growMsg{}, "spark": pushMsg{}, "area": growMsg{}, "gauge": gaugeMsg{}, "heat": warmMsg{},
	"timeline": flowMsg{}, "keyframes": flowMsg{},
	"burst": burstMsg{}, "donut": sweepMsg{}, "race": rankMsg{}, "squash": squashMsg{},
}

// chunk is the lines one kind takes in its tab; an empty kind is a gap.
type chunk struct {
	kind  string
	lines []string
}

// one is a block of a single line.
func one(kind, line string) chunk { return chunk{kind, []string{line}} }

// join is a tab's body: the lines of its blocks, one after another.
func join(blocks []chunk) string {
	var lines []string
	for _, b := range blocks {
		lines = append(lines, b.lines...)
	}
	return strings.Join(lines, "\n")
}

// bare is a row without its kind and variant columns: the title names it.
func bare(string, string) string { return "" }

// tabBlocks are the kinds of family tab, as the gallery draws them.
func (m model) tabBlocks(tab int, th theme.Resolved, label func(kind, variant string) string) []chunk {
	switch tab {
	case 1:
		return m.cv.blocks(th, label)
	case 2:
		return m.mv.blocks(th, label)
	case 3:
		return m.ph.blocks(th, label)
	case 4:
		return m.dt.blocks(th, label)
	case 5:
		return m.fl.blocks(th, label)
	case 6:
		return m.pp.blocks(th, label)
	}
	return m.basicBlocks(label)
}

// find returns the tab holding kind.
func (m model) find(kind string) (tab int, ok bool) {
	for t := 0; t < tabs; t++ {
		for _, b := range m.tabBlocks(t, m.sh.Theme(), bare) {
			if b.kind == kind {
				return t, true
			}
		}
	}
	return 0, false
}

// kinds lists every animation the gallery has, in the order it draws them.
func (m model) kinds() []string {
	var out []string
	for t := 0; t < tabs; t++ {
		for _, b := range m.tabBlocks(t, m.sh.Theme(), bare) {
			if b.kind != "" {
				out = append(out, b.kind)
			}
		}
	}
	return out
}

// solo is the gallery showing one animation, with no tabs.
type solo struct {
	m    model
	sh   shell.Shell
	kind string
	tab  int
}

func newSolo(kind string) (solo, error) {
	m := newModel()
	tab, ok := m.find(kind)
	if !ok {
		return solo{}, fmt.Errorf("no animation %q; try one of: %s", kind, strings.Join(m.kinds(), ", "))
	}
	acts := []shell.Action{shell.Quit()}
	if msg, again := replay[kind]; again {
		acts = []shell.Action{shell.Custom("r", "again", shell.Send(msg)), shell.Quit()}
	} else if kind == "spring" {
		acts = []shell.Action{shell.Custom("←/→", "move", nil, shell.DisplayOnly()), shell.Quit()}
	}
	sh := shell.New(shell.Config{
		Layout:  layout.Rows(layout.Fill("main")),
		Base:    look.Plain(),
		Title:   "animation",
		Actions: acts,
	})
	return solo{m: m, sh: sh, kind: kind, tab: tab}, nil
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
	th := s.sh.Theme()
	return s.sh.View(map[string]shell.Pane{
		"main": {Title: s.kind, Right: still(), Body: func(layout.Rect) string {
			for _, b := range s.m.tabBlocks(s.tab, th, bare) {
				if b.kind == s.kind {
					return strings.Join(b.lines, "\n")
				}
			}
			return ""
		}},
	})
}
