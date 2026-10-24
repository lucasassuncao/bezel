// Paginator: 42 items, 8 to a page, with the page under them. ← and → turn
// the page (pgup/pgdown and h/l too), and t switches between "2/6" and dots.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/component/paginator"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

const total = 42

// kindMsg switches the paginator between numbers and dots.
type kindMsg struct{}

type model struct {
	sh   shell.Shell
	th   theme.Resolved
	pg   paginator.Model
	dots bool
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("paginator")),
		Theme:  th,
		Title:  "paginator",
		Actions: []shell.Action{
			shell.Custom("←/→", "page", nil, shell.DisplayOnly()),
			shell.Custom("t", "dots/numbers", shell.Send(kindMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, pg: paginator.New(8).SetItems(total)}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case kindMsg:
		m.dots = !m.dots
		kind := paginator.Numbers
		if m.dots {
			kind = paginator.Dots
		}
		m.pg = m.pg.WithKind(kind)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.pg, _ = m.pg.Update(msg)
	}
	return m, nil
}

func (m model) body(r layout.Rect) string {
	out := []string{draw.Heading("Items", m.th.Section, r.W), ""}
	start, end := m.pg.Bounds(total)
	for i := start; i < end; i++ {
		out = append(out, fmt.Sprintf("  item %d", i+1))
	}
	for range m.pg.PerPage() - (end - start) {
		out = append(out, "")
	}
	return strings.Join(append(out, "", "  "+m.pg.View(m.th)), "\n")
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"paginator": {Title: "paginator", Body: m.body},
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
