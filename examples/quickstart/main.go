// Quickstart: the whole program the README shows. Its test fails when the
// two differ, so the README's copy keeps compiling.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/list"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

type model struct {
	sh   shell.Shell
	th   theme.Resolved
	list list.Model[string]
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	return model{
		th: th,
		sh: shell.New(shell.Config{
			Title: "fruit",
			Theme: th,
			Layout: layout.Columns(
				layout.Fixed("list", layout.Ratio(1, 3), layout.Min(20)),
				layout.Fill("detail"),
			),
			Actions: []shell.Action{shell.Move(), shell.Help(), shell.Quit()},
		}),
		list: list.New([]list.Row[string]{
			{Label: "apple", Value: "red and crisp"},
			{Label: "banana", Value: "yellow and soft"},
			{Label: "cherry", Value: "small and sweet"},
		}),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// The shell goes first: resize, dialogs and the keys of its actions.
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	// Whatever is left is yours: here, the arrows move the list.
	m.list = m.list.SetHeight(draw.InnerRect(m.sh.Rect("list")).H)
	m.list, r = m.list.Update(msg)
	return m, r.Cmd
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"list": {Title: "Fruit", Body: func(layout.Rect) string {
			return m.list.View(m.th)
		}},
		"detail": {Title: "Detail", Body: func(layout.Rect) string {
			row, _ := m.list.Selected()
			return row.Value
		}},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
