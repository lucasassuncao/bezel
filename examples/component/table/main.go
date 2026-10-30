// Table: a cursor over rows that scroll under a fixed header, with the position
// and a scrollbar in the pane's edges and the row's detail below. w narrows the
// window: the description gives way first, version and licence stay whole.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/table"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
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
	{"testify", "v1.11.1", "MIT", "Assertions and mocks for Go tests"},
	{"clipboard", "v0.1.4", "BSD-3", "Clipboard access for Go"},
	{"x/ansi", "v0.11.9", "MIT", "Parse, measure and cut strings with ANSI escapes"},
	{"x/term", "v0.2.2", "MIT", "Terminal size and raw mode"},
	{"x/termios", "v0.1.1", "MIT", "termios for Unix terminals"},
	{"x/windows", "v0.2.2", "MIT", "Windows console calls"},
	{"colorprofile", "v0.4.3", "MIT", "The colours a terminal can show"},
	{"displaywidth", "v0.11.0", "MIT", "How many cells a string takes on screen"},
	{"uax29", "v2.7.0", "MIT", "Unicode text segmentation"},
	{"go-colorful", "v1.4.1", "MIT", "Colour spaces and blending"},
	{"go-runewidth", "v0.0.31", "MIT", "Cell width of runes and strings"},
	{"cancelreader", "v0.2.2", "MIT", "A reader that can be cancelled"},
	{"uniseg", "v0.4.7", "MIT", "Unicode grapheme clusters and line breaks"},
	{"terminfo", "v1.2.0", "MIT", "A terminfo database reader"},
	{"x/sys", "v0.49.0", "BSD-3", "Low-level operating system calls"},
	{"x/sync", "v0.24.0", "BSD-3", "errgroup, semaphore and singleflight"},
}

// items carry each row's name, for the status enter sets.
func items() []table.Item[string] {
	out := make([]table.Item[string], len(rows))
	for i, r := range rows {
		out[i] = table.Item[string]{Cells: r, Value: r[0]}
	}
	return out
}

type toggleWidthMsg struct{}

type model struct {
	sh           shell.Shell
	th           theme.Resolved
	tbl          table.Model[string]
	realW, realH int
	narrow       bool
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fixed("table", layout.Ratio(3, 5)), layout.Fill("detail").Info()),
		Theme:  th,
		Title:  "table",
		Actions: []shell.Action{
			shell.Move(),
			shell.Custom("enter", "choose", nil, shell.DisplayOnly()),
			shell.Custom("w", "narrow/wide", shell.Send(toggleWidthMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, tbl: table.New(cols, items())}
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
	m.tbl = m.tbl.SetHeight(draw.InnerRect(m.sh.Rect("table")).H)
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
		var r bezel.Result
		m.tbl, r = m.tbl.Update(msg)
		if e, ok := r.Event.(table.Chosen[string]); ok {
			m.sh, cmd = m.sh.SetStatus("chose "+e.Item.Value, shell.OK, 2*time.Second)
			return m, cmd
		}
	}
	return m, nil
}

// detail is the row under the cursor, its description wrapped to the pane.
func (m model) detail(r layout.Rect) string {
	it, ok := m.tbl.Selected()
	if !ok {
		return ""
	}
	c := it.Cells
	out := draw.KV("version", c[1], m.th.Dim, r.W) + draw.KV("licence", c[2], m.th.Dim, r.W) + "\n"
	return out + "  " + strings.Join(draw.Wrap(c[3], r.W-2), "\n  ")
}

func (m model) View() tea.View {
	it, _ := m.tbl.Selected()
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"table": {
			Title:  "dependencies",
			Bottom: fmt.Sprintf("%d/%d", m.tbl.Cursor()+1, m.tbl.Len()),
			Scroll: m.tbl.Scroll(),
			Body:   func(r layout.Rect) string { return m.tbl.SetWidth(r.W).View(m.th) },
		},
		"detail": {Title: it.Value, Body: m.detail},
	}))
	v.AltScreen = true
	v.BackgroundColor, v.ForegroundColor = m.th.Background, m.th.Foreground
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
