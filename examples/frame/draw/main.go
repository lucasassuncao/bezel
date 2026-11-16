// Draw: one scrolling pane with a section per helper, each drawn at the
// pane's own width.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

const paragraph = "Wrap folds plain text into rows of at most the given width, breaking after spaces and hyphens, and cuts a word wider than a row."

const source = "package main\n\nfunc main() {\n    println(\"hi\")\n}"

type model struct {
	sh     shell.Shell
	offset int
}

func newModel() model {
	sh := shell.New(shell.Config{
		Layout:  layout.Columns(layout.Fill("draw")),
		Base:    theme.Mint(),
		Title:   "draw",
		Actions: []shell.Action{shell.Scroll(), shell.Quit()},
	})
	return model{sh: sh}
}

func (m model) Init() tea.Cmd { return m.sh.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "up":
			m.offset = max(0, m.offset-1)
		case "down":
			r := m.sh.Inner("draw")
			m.offset = min(m.offset+1, max(0, len(m.lines(r.W))-r.H))
		}
	}
	return m, nil
}

// sections draws every helper at width w, one block per helper.
func (m model) sections(w int) []string {
	th := m.sh.Theme()
	var out []string
	add := func(name string, lines ...string) {
		out = append(out, draw.Heading(name, th.Section, w))
		out = append(out, lines...)
		out = append(out, "")
	}
	add("Header", draw.Header(w, "my app", "kv/app/prod", "v1.2.0", th.Chrome))
	add("Tabs", draw.Tabs(w, []string{"secrets", "policies", "auth"}, 1, th.Chrome))
	add("Divider", draw.Divider(w, "", draw.DividerStyle{Line: th.Muted}), draw.Divider(w, "Packages", draw.DividerStyle{Line: th.Muted, Label: th.Accent}))
	card := draw.Card(layout.Rect{W: min(w, 40), H: 7}, "git", "version  2.47.0\nsource   winget\nscope    machine",
		th.Key.Render("[u]")+" upgrade  "+th.Key.Render("[r]")+" remove", th.Panel)
	add("Card", strings.Split(card, "\n")...)
	add("Button", draw.Button("Save", th.Button, true)+"  "+draw.Button("Cancel", th.Button, false))
	add("Wrap", draw.Wrap(paragraph, min(w, 40))...)
	add("Truncate", draw.Truncate("a line much longer than the room it has to fit in", 24))
	add("Link", draw.Link("bezel on GitHub", "https://github.com/lucasassuncao/bezel", th.Accent.Underline(true)))
	add("NumberLines", draw.NumberLines(source, th.Muted))
	add("EmptyState", draw.EmptyState(layout.Rect{W: w, H: 4}, "nothing here", "press n to add one", th.Chrome))
	return out
}

// lines is the whole content at width w, one entry per screen row.
func (m model) lines(w int) []string {
	return strings.Split(strings.Join(m.sections(w), "\n"), "\n")
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"draw": {Title: "draw", Body: func(r layout.Rect) string {
			lines := m.lines(r.W)
			return strings.Join(lines[min(m.offset, len(lines)-1):], "\n")
		}},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
