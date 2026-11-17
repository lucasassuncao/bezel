// The gallery narrowed to one kind by --text, so a recording can show that
// kind alone, as the GIFs under docs/recordings/animation/text/solo do.
package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// tabs is how many families the gallery has.
const tabs = 4

// replay is the message that plays a kind again; a kind missing here loops.
var replay = map[string]tea.Msg{
	"typewriter": typeMsg{}, "shake": shakeMsg{}, "scramble": scrambleMsg{},
	"odometer": odometerMsg{}, "decrypt": decryptMsg{}, "ticker": tickerMsg{},
	"highlight": markMsg{}, "strike": strikeMsg{}, "rise": copyMsg{},
	"stream": streamMsg{}, "morph": morphMsg{}, "typist": typistMsg{},
	"flip": flipMsg{}, "big": flipMsg{}, "burn": burnMsg{}, "fall": fallMsg{},
}

// block is the lines one kind takes in its tab; an empty kind is a gap.
type block struct {
	kind  string
	lines []string
}

// one is a block of a single line.
func one(kind, line string) block { return block{kind, []string{line}} }

// join is a tab's body: the lines of its blocks, one after another.
func join(blocks []block) string {
	var lines []string
	for _, b := range blocks {
		lines = append(lines, b.lines...)
	}
	return strings.Join(lines, "\n")
}

// bare is a row without its kind and variant columns: the title names it.
func bare(string, string) string { return "" }

// tabBlocks are the kinds of family tab, as the gallery draws them.
func (m model) tabBlocks(tab int, th theme.Resolved, label func(kind, variant string) string) []block {
	switch tab {
	case 1:
		return m.ro.blocks(th, label)
	case 2:
		return m.tx.blocks(th, label)
	case 3:
		return m.wd.blocks(th, label)
	}
	return m.ty.blocks(th, label)
}

// find returns the tab and the lines of kind, bare of their labels.
func (m model) find(kind string) (tab int, lines []string, ok bool) {
	for t := 0; t < tabs; t++ {
		for _, b := range m.tabBlocks(t, m.sh.Theme(), bare) {
			if b.kind == kind {
				return t, b.lines, true
			}
		}
	}
	return 0, nil, false
}

// kinds lists every kind the gallery has, in the order it draws them.
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

// solo is the gallery showing one kind, with no tabs.
type solo struct {
	m    model
	sh   shell.Shell
	kind string
	tab  int
}

func newSolo(kind string) (solo, error) {
	m := newModel()
	tab, _, ok := m.find(kind)
	if !ok {
		return solo{}, fmt.Errorf("no text kind %q; try one of: %s", kind, strings.Join(m.kinds(), ", "))
	}
	acts := []shell.Action{shell.Quit()}
	if msg, again := replay[kind]; again {
		acts = []shell.Action{shell.Custom("r", "again", shell.Send(msg)), shell.Quit()}
	}
	sh := shell.New(shell.Config{
		Layout:  layout.Rows(layout.Fill("main")),
		Base:    theme.Mint(),
		Title:   "text",
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
