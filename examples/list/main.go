// List: rows under section headings, a / filter, and enter showing the
// payload the chosen row carries.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/list"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// food is the payload a row carries; the list never looks inside it.
type food struct {
	name     string
	calories int
}

func rows() []list.Row {
	row := func(name string, kcal int) list.Row {
		return list.Row{Label: name, Mark: "•", Value: food{name, kcal}}
	}
	return []list.Row{
		{Label: "Fruit", Section: true},
		row("banana", 89), row("mango", 60), row("orange", 47), row("pineapple", 50),
		{Section: true},
		{Label: "Vegetables", Section: true},
		row("carrot", 41), row("spinach", 23), row("tomato", 18), row("onion", 40),
	}
}

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	list   list.Model
	chosen *food
}

func newModel() model {
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fixed("list", layout.Ratio(1, 2)), layout.Fill("detail")),
		Theme:  th,
		Title:  "list",
		Actions: []shell.Action{
			shell.Move(),
			shell.Custom("/", "filter", nil, shell.DisplayOnly()),
			shell.Custom("enter", "choose", nil, shell.DisplayOnly()),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, list: list.New(rows(), 0)}
}

func (m model) Init() tea.Cmd { return nil }

// toList hands a key to the list and reads what it did.
func (m model) toList(key tea.KeyPressMsg) model {
	var act list.Action
	m.list = m.list.SetHeight(draw.InnerRect(m.sh.Rect("list")).H)
	if m.list, act = m.list.Update(key); act == list.Chosen {
		f := m.list.Selected().Value.(food)
		m.chosen = &f
	}
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// While filtering every key is text for the filter, q included.
	if key, ok := msg.(tea.KeyPressMsg); ok && m.list.IsFiltering() {
		return m.toList(key), nil
	}
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.toList(key), nil
	}
	return m, nil
}

func (m model) detail() string {
	if m.chosen == nil {
		return m.th.Dim.Render("enter shows the row's payload")
	}
	return fmt.Sprintf("%s\n\n%d kcal per 100 g", m.th.Accent.Render(m.chosen.name), m.chosen.calories)
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"list": {Title: "food", Body: func(r layout.Rect) string {
			return m.list.SetHeight(r.H).View(m.th)
		}},
		"detail": {Title: "payload", Body: func(layout.Rect) string { return m.detail() }},
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
