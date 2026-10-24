// Package themebrowser provides a small, inline (not full-screen) terminal
// table listing bezel's built-in theme names next to their category and palette.
package themebrowser

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"

	"github.com/lucasassuncao/bezel/component/table"
	"github.com/lucasassuncao/bezel/theme"
)

// maxVisibleRows caps the table's height before it scrolls, so the ~50
// built-in themes don't grow the box past a reasonable size.
const maxVisibleRows = 15

// wide is more than any row of the table needs: each line ends where its
// last cell does, and the box sizes itself to the longest.
const wide = 512

var cols = []table.Column{{Title: "ID"}, {Title: "Theme"}, {Title: "Category"}, {Title: "Palette"}}

type browserModel struct {
	tbl table.Model[string]
	th  theme.Resolved
}

// newBrowserModel lists the built-in themes in th's colours. dark picks the
// variant of the adaptive default, for the table and for each palette.
func newBrowserModel(th theme.Theme, dark bool) *browserModel {
	all := theme.All()
	var items []table.Item[string]
	for _, cat := range theme.Categories() {
		for _, name := range cat.Themes {
			id := strconv.Itoa(len(items) + 1)
			cells := []string{id, name, cat.Name, palette(all[name], dark)}
			items = append(items, table.Item[string]{Cells: cells, Value: name})
		}
	}
	height := min(max(len(items), 1), maxVisibleRows) + 1 // and the header
	return &browserModel{tbl: table.New(cols, items, height), th: theme.Resolve(th, dark)}
}

// palette is one block per colour that tells themes apart: accent, selection,
// border, dim, success and danger, as t resolves on this background.
func palette(t theme.Theme, dark bool) string {
	c := theme.ResolveColors(t, dark)
	blocks := make([]string, 0, 6)
	for _, hex := range []string{c.Accent, c.Selection, c.Border, c.Dim, c.Success, c.Danger} {
		blocks = append(blocks, lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render("██"))
	}
	return strings.Join(blocks, " ")
}

func (m *browserModel) Init() tea.Cmd { return nil }

// Update quits on q or ctrl+c and hands every other key to the table.
func (m *browserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	m.tbl, _ = m.tbl.Update(keyMsg)
	return m, nil
}

// View renders inline, without AltScreen. The border sizes itself to the
// table's content, so there is no manual width/height math to drift.
func (m *browserModel) View() tea.View {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.th.Colors.Border)).
		Padding(0, 1)
	return tea.NewView(box.Render(m.tbl.View(m.th, wide)))
}

// List writes the built-in themes, one line per category: what --list-themes
// prints where there is no terminal to browse in, such as a pipe or a script.
func List(w io.Writer) error {
	for _, c := range theme.Categories() {
		names := slices.Sorted(slices.Values(c.Themes))
		if _, err := fmt.Fprintf(w, "  %-14s %s\n", c.Name, strings.Join(names, ", ")); err != nil {
			return err
		}
	}
	return nil
}

// BrowseInTerminal renders an inline, scrollable table (↑/↓, q/ctrl+c to
// quit) of every built-in theme and its theme.Categories() category. An
// optional theme.Theme sets colors; without a terminal on stdout it Lists.
func BrowseInTerminal(t ...theme.Theme) error {
	if !term.IsTerminal(os.Stdout.Fd()) {
		return List(os.Stdout)
	}
	th := theme.Theme{}
	if len(t) > 0 {
		th = t[0]
	}
	m := newBrowserModel(th, theme.DarkTerminal())
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run theme browser: %w", err)
	}
	return nil
}
