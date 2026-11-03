// Tags: chips that filter a list. Type a topic and enter adds it, tab takes
// the suggestion shown, backspace on an empty field takes the last chip off;
// the packages left are the ones with every topic.
package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/tags"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// pkg is a package and the topics it is tagged with.
type pkg struct {
	name   string
	topics []string
}

var pkgs = []pkg{
	{"bubbletea", []string{"tui", "framework"}},
	{"lipgloss", []string{"tui", "style"}},
	{"glamour", []string{"markdown", "style"}},
	{"cobra", []string{"cli", "framework"}},
	{"viper", []string{"cli", "config"}},
	{"koanf", []string{"config"}},
	{"chroma", []string{"syntax", "style"}},
	{"termenv", []string{"tui", "color"}},
}

// topics is every topic, for the suggestions.
func topics() []string {
	var out []string
	for _, p := range pkgs {
		for _, t := range p.topics {
			if !slices.Contains(out, t) {
				out = append(out, t)
			}
		}
	}
	slices.Sort(out)
	return out
}

type model struct {
	sh shell.Shell
	th theme.Resolved
	tg tags.Model
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fixed("filter", layout.Lines(3)), layout.Fill("packages")),
		Theme:  th,
		Title:  "tags",
		Actions: []shell.Action{
			shell.Custom("enter", "add", nil, shell.DisplayOnly()),
			shell.Custom("tab", "complete", nil, shell.DisplayOnly()),
			shell.Custom("backspace", "remove", nil, shell.DisplayOnly()),
			// Every letter is text here, so esc quits.
			shell.Quit(shell.WithKey("esc")),
		},
	})
	tg := tags.New(th).WithPlaceholder("a topic, then enter").SetSuggestions(topics())
	tg, _ = tg.Focus()
	return model{sh: sh, th: th, tg: tg}
}

func (m model) Init() tea.Cmd { return m.tg.Blink() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "ctrl+c" {
		return m, tea.Quit
	}
	var r bezel.Result
	if m.tg, r = m.tg.Update(msg); r.Handled {
		return m, r.Cmd
	}
	m.sh, r = m.sh.Update(msg)
	return m, r.Cmd
}

// matching is the packages tagged with every chip.
func (m model) matching() []pkg {
	var out []pkg
	for _, p := range pkgs {
		if !slices.ContainsFunc(m.tg.Tags(), func(t string) bool { return !slices.Contains(p.topics, t) }) {
			out = append(out, p)
		}
	}
	return out
}

func (m model) list(r layout.Rect) string {
	var out []string
	for _, p := range m.matching() {
		out = append(out, fmt.Sprintf("%-11s", p.name)+m.th.Dim.Render(strings.Join(p.topics, ", ")))
	}
	if len(out) == 0 {
		out = append(out, m.th.Dim.Render("no package has every topic"))
	}
	return draw.FitBlock(strings.Join(out, "\n"), r.W, r.H)
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"filter":   {Title: "topics", Body: func(r layout.Rect) string { return m.tg.SetWidth(r.W).View() }},
		"packages": {Title: "packages", Bottom: fmt.Sprintf("%d of %d", len(m.matching()), len(pkgs)), Body: m.list},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
