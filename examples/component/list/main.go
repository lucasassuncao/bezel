// List: rows under section headings, a / filter, enter showing the payload
// the chosen row carries, and a cursor that glides from row to row.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/component/list"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// food is the payload a row carries; the list never looks inside it.
type food struct {
	name     string
	calories int
}

func rows() []list.Row[food] {
	row := func(name string, kcal int) list.Row[food] {
		return list.Row[food]{Label: name, Mark: "•", Value: food{name, kcal}}
	}
	return []list.Row[food]{
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
	list   list.Model[food]
	chosen *food
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
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
	return model{sh: sh, th: th, list: list.New(rows(), 0).WithMotion(120 * time.Millisecond)}
}

func (m model) Init() tea.Cmd { return nil }

// toList hands a key to the list, reads what it did, and starts the cursor
// gliding to where the key moved it.
func (m model) toList(key tea.KeyPressMsg) (model, tea.Cmd) {
	var act list.Action
	m.list = m.list.SetHeight(draw.InnerRect(m.sh.Rect("list")).H)
	if m.list, act = m.list.Update(key); act == list.Chosen {
		f := m.list.Selected().Value
		m.chosen = &f
	}
	return m, m.list.Tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// While filtering every key is text for the filter, q included.
	if key, ok := msg.(tea.KeyPressMsg); ok && m.list.IsFiltering() {
		return m.toList(key)
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
		return m.toList(key)
	}
	var cmd2 tea.Cmd
	if m.list, cmd2 = m.list.Animate(msg); cmd2 != nil {
		return m, cmd2
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
	// Sized before the pane is drawn: the scrollbar is the pane's, not the body's.
	l := m.list.SetHeight(draw.InnerRect(m.sh.Rect("list")).H)
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"list": {Title: "food", Scroll: l.Scroll(), Body: func(layout.Rect) string {
			return l.View(m.th)
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
