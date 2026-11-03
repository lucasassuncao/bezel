// List: rows under section headings, a / filter, enter showing the payload,
// a gliding cursor, space checking rows for a batch and a spinner on a busy
// row, which b turns on and off.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
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
		row("banana", 89), row("mango", 60), busy(row("orange", 47)), row("pineapple", 50),
		{Section: true},
		{Label: "Vegetables", Section: true},
		row("carrot", 41), row("spinach", 23), row("tomato", 18), row("onion", 40),
	}
}

// busy is r with a spinner turning in place of its mark.
func busy(r list.Row[food]) list.Row[food] { r.Busy = true; return r }

type toggleBusyMsg struct{}

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	list   list.Model[food]
	chosen *food
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fixed("list", layout.Ratio(1, 2)), layout.Fill("detail")),
		Theme:  th,
		Title:  "list",
		Actions: []shell.Action{
			shell.Move(),
			shell.Custom("/", "filter", nil, shell.DisplayOnly()),
			shell.Custom("enter", "choose", nil, shell.DisplayOnly()),
			shell.Custom("space", "check", nil, shell.DisplayOnly()),
			shell.Custom("b", "busy", shell.Send(toggleBusyMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, list: list.New(rows()).WithMotion(120 * time.Millisecond).WithMulti(true)}
}

// Init turns the spinner of the row that starts busy.
func (m model) Init() tea.Cmd { return m.list.Tick() }

// toList hands a message to the list and reads what it did; the Cmd it
// returns starts the cursor gliding to where the key moved it.
func (m model) toList(msg tea.Msg) (model, tea.Cmd) {
	m.list = m.list.SetHeight(draw.InnerRect(m.sh.Rect("list")).H)
	var r bezel.Result
	m.list, r = m.list.Update(msg)
	if c, ok := r.Event.(list.Chosen[food]); ok {
		m.chosen = &c.Row.Value
	}
	return m, r.Cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// While filtering every key is text for the filter, q included.
	if key, ok := msg.(tea.KeyPressMsg); ok && m.list.Filtering() {
		return m.toList(key)
	}
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if _, ok := msg.(toggleBusyMsg); ok {
		return m.toggleBusy()
	}
	return m.toList(msg)
}

func (m model) detail() string {
	out := m.th.Dim.Render("enter shows the row's payload")
	if m.chosen != nil {
		out = fmt.Sprintf("%s\n\n%d kcal per 100 g", m.th.Accent.Render(m.chosen.name), m.chosen.calories)
	}
	var names []string
	for _, r := range m.list.Checked() {
		names = append(names, r.Label)
	}
	if len(names) > 0 {
		out += "\n\n" + m.th.Dim.Render("checked: ") + strings.Join(names, ", ")
	}
	return out
}

func (m model) View() tea.View {
	// Sized before the pane is drawn: the scrollbar is the pane's, not the body's.
	l := m.list.SetHeight(draw.InnerRect(m.sh.Rect("list")).H)
	return m.sh.View(map[string]shell.Pane{
		"list": {Title: "food", Bottom: fmt.Sprintf("%d checked", len(l.Checked())), Scroll: l.Scroll(), Body: func(layout.Rect) string {
			return l.View(m.th)
		}},
		"detail": {Title: "payload", Body: func(layout.Rect) string { return m.detail() }},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// toggleBusy starts or stops the spinner on the row under the cursor.
func (m model) toggleBusy() (model, tea.Cmd) {
	sel, ok := m.list.Selected()
	if !ok {
		return m, nil
	}
	for i, r := range m.list.Rows() {
		if !r.Section && r.Label == sel.Label {
			r.Busy = !r.Busy
			m.list = m.list.SetRow(i, r)
		}
	}
	return m, m.list.Tick()
}
