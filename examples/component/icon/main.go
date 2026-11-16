// Icon: every state mark in both sets, then a list drawn with the set in use;
// a swaps the list between Unicode and ASCII.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

var states = []struct {
	name string
	st   icon.State
}{
	{"OK", icon.OK}, {"Fail", icon.Fail}, {"Warn", icon.Warn}, {"Info", icon.Info},
	{"Pending", icon.Pending}, {"Running", icon.Running}, {"Skipped", icon.Skipped},
}

// packages is what the sample list reports, one state each.
var packages = []struct {
	name, note string
	st         icon.State
}{
	{"git", "2.47.0", icon.OK},
	{"neovim", "installer exited with 1603", icon.Fail},
	{"docker", "needs a restart", icon.Warn},
	{"lazygit", "downloading", icon.Running},
	{"ffmpeg", "queued", icon.Pending},
	{"jq", "action: skip", icon.Skipped},
}

type swapMsg struct{}

type model struct {
	sh    shell.Shell
	ascii bool
}

func newModel() model {
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("icons")),
		Base:   theme.Mint(),
		Title:  "icon",
		Actions: []shell.Action{
			shell.Custom("a", "swap set", shell.Send(swapMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh}
}

func (m model) Init() tea.Cmd { return m.sh.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	if _, ok := msg.(swapMsg); ok {
		m.ascii = !m.ascii
	}
	return m, nil
}

func (m model) set() (icon.Set, string) {
	if m.ascii {
		return icon.ASCII(), "ASCII"
	}
	return icon.Unicode(), "Unicode"
}

func (m model) body(r layout.Rect) string {
	th := m.sh.Theme()
	out := []string{draw.Heading("States", th.Section, r.W), ""}
	for _, s := range states {
		out = append(out, fmt.Sprintf("  %-9s %s   %s", s.name, icon.Unicode().Render(s.st, th), icon.ASCII().Render(s.st, th)))
	}
	set, name := m.set()
	out = append(out, "", draw.Heading("Packages · "+name, th.Section, r.W), "")
	for _, p := range packages {
		out = append(out, fmt.Sprintf("  %s %-8s %s", set.Render(p.st, th), p.name, th.Dim.Render(p.note)))
	}
	out = append(out, "", draw.Heading("Tree · "+name, th.Section, r.W), "")
	return strings.Join(append(out, m.tree(set)...), "\n")
}

// tree is a source chain drawn with the set's connectors: each source hangs off
// the application, and the one that answered carries why.
func (m model) tree(set icon.Set) []string {
	th := m.sh.Theme()
	conn := func(last bool) string { return th.Muted.Render(set.Connector(last)) }
	indent := func(last bool) string { return th.Muted.Render(set.Indent(last)) }
	return []string{
		"  lazygit",
		"  " + conn(false) + "winget  " + th.Dim.Render("not offered"),
		"  " + conn(false) + "scoop   " + th.Success.Render("installed 0.42.0"),
		"  " + indent(false) + conn(true) + th.Dim.Render("bucket extras"),
		"  " + conn(true) + "github  " + th.Dim.Render("not tried"),
	}
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"icons": {Title: "icons", Body: m.body},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
