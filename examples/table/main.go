// Table: columns sized to their contents; w narrows the window and the
// description gives way first, while version and licence stay whole.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/table"
	"github.com/lucasassuncao/bezel/theme"
)

// narrowWidth is small enough that the table no longer fits whole.
const narrowWidth = 60

var cols = []table.Column{
	{Title: "NAME", Min: 8},
	{Title: "VERSION"},
	{Title: "LICENCE"},
	{Title: "DESCRIPTION", Min: 10, Flex: true},
}

var rows = [][]string{
	{"bubbletea", "v2.0.9", "MIT", "A framework for terminal apps, after The Elm Architecture"},
	{"lipgloss", "v2.0.6", "MIT", "Style definitions for nice terminal layouts"},
	{"bubbles", "v2.2.1", "MIT", "Components for Bubble Tea: text inputs, viewports, spinners"},
	{"testify", "v1.11.1", "MIT", "Assertions and mocks for Go tests"},
	{"clipboard", "v0.1.4", "BSD-3", "Clipboard access for Go"},
}

type toggleWidthMsg struct{}

type model struct {
	sh           shell.Shell
	th           theme.Resolved
	realW, realH int
	narrow       bool
}

func newModel() model {
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("table")),
		Theme:  th,
		Title:  "table",
		Actions: []shell.Action{
			shell.Custom("w", "narrow/wide", shell.Send(toggleWidthMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th}
}

func (m model) Init() tea.Cmd { return nil }

// resize hands the shell the width to draw at: the real one or narrowWidth.
// VHS cannot resize a terminal mid-recording, so this stands in for it.
func (m model) resize() model {
	w := m.realW
	if m.narrow {
		w = min(narrowWidth, m.realW)
	}
	m.sh, _, _ = m.sh.Update(tea.WindowSizeMsg{Width: w, Height: m.realH})
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.realW, m.realH = size.Width, size.Height
		return m.resize(), nil
	}
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case toggleWidthMsg:
		m.narrow = !m.narrow
		return m.resize(), nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// render fits the columns to width and lays out the header and every row.
func (m model) render(width int) string {
	widths := table.Fit(cols, rows, width)
	lines := []string{m.th.TableHeader.Render(table.Titles(cols, widths))}
	for _, r := range rows {
		lines = append(lines, table.Row(widths, r))
	}
	return strings.Join(lines, "\n")
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"table": {Title: "dependencies", Body: func(r layout.Rect) string { return m.render(r.W) }},
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
