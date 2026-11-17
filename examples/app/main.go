// Demo: a small task list built from a few bezel components. docs/EXAMPLE.md
// walks through it step by step.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/list"
	"github.com/lucasassuncao/bezel/component/progress"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// task is what the app keeps; the list only draws it.
type task struct {
	title string
	done  bool
}

// The messages the keys send. The model handles them in Update, like any other.
type (
	addMsg     struct{ title string }
	toggleMsg  struct{}
	askDropMsg struct{}
	dropMsg    struct{}
	askAddMsg  struct{}
)

type model struct {
	sh    shell.Shell
	tasks []task
	list  list.Model[int] // each row carries its task's index
}

func newModel(readOnly bool) model {
	sh := shell.New(shell.Config{
		Layout: layout.Columns(
			layout.Fixed("tasks", layout.Ratio(1, 2), layout.Min(30)),
			layout.Fill("detail").Info(),
		).Collapse(70, "tasks"),
		Base:     theme.Plain(),
		Title:    "tasks",
		Subtitle: "a bezel demo",
		Can:      func(c shell.Capability) bool { return c != "write" || !readOnly },
		Actions: []shell.Action{
			shell.Move(),
			shell.Custom("n", "new", shell.Send(askAddMsg{})),
			shell.Custom("space", "done", shell.Send(toggleMsg{})),
			shell.Custom("d", "delete", shell.Send(askDropMsg{}), shell.Needs("write")),
			shell.Help(),
			shell.Quit(),
		},
	})
	m := model{sh: sh, list: list.New[int](nil), tasks: []task{
		{title: "write the release notes", done: true},
		{title: "record the demo"},
		{title: "answer the issue about themes"},
		{title: "bump the apps to the new tag"},
	}}
	return m.refresh()
}

// refresh rebuilds the list's rows from the tasks, keeping where the cursor is.
func (m model) refresh() model {
	rows := make([]list.Row[int], len(m.tasks))
	for i, t := range m.tasks {
		rows[i] = list.Row[int]{Label: t.title, Mark: "○", Value: i}
		if t.done {
			rows[i].Mark = "✓"
		}
	}
	m.list = m.list.SetRows(rows)
	return m
}

// current is the index of the task under the cursor, false when there is none.
func (m model) current() (int, bool) {
	if row, ok := m.list.Selected(); ok {
		return row.Value, true
	}
	return 0, false
}

func (m model) Init() tea.Cmd { return m.sh.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	// The list scrolls by how many rows its pane shows.
	m.list = m.list.SetHeight(m.sh.Inner("tasks").H)
	switch msg := msg.(type) {
	case askAddMsg:
		ask := overlay.NewPrompt("New task", func(title string) tea.Cmd {
			return bezel.Emit(addMsg{title: title})
		}).WithValidate(func(typed string) error {
			if strings.TrimSpace(typed) == "" {
				return errors.New("a task needs a title")
			}
			return nil
		})
		return m, overlay.Open(ask)
	case addMsg:
		m.tasks = append(slices.Clone(m.tasks), task{title: strings.TrimSpace(msg.title)})
		return m.refresh(), shell.Notify("added "+msg.title, shell.OK)
	case toggleMsg:
		if i, ok := m.current(); ok {
			m.tasks = slices.Clone(m.tasks)
			m.tasks[i].done = !m.tasks[i].done
			m = m.refresh()
		}
	case askDropMsg:
		if i, ok := m.current(); ok {
			yes := bezel.Emit(dropMsg{})
			ask := overlay.NewConfirm("Delete task", m.tasks[i].title, yes).WithKind(draw.Danger)
			return m, overlay.Open(ask)
		}
	case dropMsg:
		if i, ok := m.current(); ok {
			title := m.tasks[i].title
			m.tasks = slices.Delete(slices.Clone(m.tasks), i, i+1)
			return m.refresh(), shell.Notify("deleted "+title, shell.OK)
		}
	case tea.KeyPressMsg:
		// Every key no action took is the list's: the arrows, home and end.
		m.list, _ = m.list.Update(msg)
	}
	return m, nil
}

// doneCount is how many tasks are done.
func (m model) doneCount() int {
	n := 0
	for _, t := range m.tasks {
		if t.done {
			n++
		}
	}
	return n
}

// detail is the selected task, and how far along the whole list is.
func (m model) detail(r layout.Rect) string {
	th := m.sh.Theme()
	i, ok := m.current()
	if !ok {
		return draw.EmptyState(r, "nothing to do", "n adds a task", th.Chrome)
	}
	t := m.tasks[i]
	state := th.Warning.Render("to do")
	if t.done {
		state = th.Success.Render("done")
	}
	return strings.Join([]string{
		th.Accent.Render(t.title),
		"",
		th.Dim.Render("state  ") + state,
		"",
		progress.Line("all tasks", m.doneCount(), len(m.tasks), r.W, th),
	}, "\n")
}

func (m model) View() tea.View { return m.sh.View(m.panes()) }

// panes is what View draws, apart so a test can check its names.
func (m model) panes() map[string]shell.Pane {
	return map[string]shell.Pane{
		"tasks": {
			Title:  "tasks",
			Bottom: fmt.Sprintf("%d/%d done", m.doneCount(), len(m.tasks)),
			Scroll: m.list.Scroll(),
			Body:   func(layout.Rect) string { return m.list.View(m.sh.Theme()) },
		},
		"detail": {Title: "detail", Body: m.detail},
	}
}

func main() {
	readOnly := flag.Bool("read-only", false, "hide the keys that change the list")
	flag.Parse()
	if _, err := tea.NewProgram(newModel(*readOnly)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
