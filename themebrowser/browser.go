// Package themebrowser provides a small, inline (not full-screen) terminal
// table listing bezel's built-in theme names next to their category.
package themebrowser

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"

	"github.com/lucasassuncao/bezel/theme"
)

const colID = "ID"

// maxVisibleRows caps the table's height before it scrolls, so the ~50
// built-in themes don't grow the box past a reasonable size.
const maxVisibleRows = 15

type browserModel struct {
	tbl    table.Model
	colors theme.Colors
}

func newBrowserModel(colors theme.Colors) *browserModel {
	const (
		colTheme    = "Theme"
		colCategory = "Category"
	)
	themeW := lipgloss.Width(colTheme)
	catW := lipgloss.Width(colCategory)

	var names [][2]string // theme name, category name
	for _, cat := range theme.Categories() {
		for _, name := range cat.Themes {
			names = append(names, [2]string{name, cat.Name})
			themeW = max(themeW, lipgloss.Width(name))
			catW = max(catW, lipgloss.Width(cat.Name))
		}
	}

	idW := len(strconv.Itoa(len(names)))
	if idW < len(colID) {
		idW = len(colID)
	}
	rows := make([]table.Row, len(names))
	for i, n := range names {
		rows[i] = table.Row{strconv.Itoa(i + 1), n[0], n[1]}
	}

	visibleRows := len(rows)
	if visibleRows > maxVisibleRows {
		visibleRows = maxVisibleRows
	}
	if visibleRows < 1 {
		visibleRows = 1
	}

	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: colID, Width: idW},
			{Title: colTheme, Width: themeW},
			{Title: colCategory, Width: catW},
		}),
		table.WithRows(rows),
		table.WithStyles(tableStyles(colors)),
		table.WithWidth((idW+2)+(themeW+2)+(catW+2)),
		table.WithHeight(visibleRows+1), // +1: WithHeight subtracts the 1-line header row internally
	)

	return &browserModel{tbl: tbl, colors: colors}
}

// tableStyles keeps Padding off Selected: renderRow pads each cell, then
// wraps the row in Selected, so padding there widens the cursor row. Selected
// gets a Background too, since a text-color-only cue is easy to miss on a wide row.
func tableStyles(c theme.Colors) table.Styles {
	selected := lipgloss.NewStyle().Bold(true).
		Background(lipgloss.Color(c.Selection)).
		Foreground(contrastText(c.Selection))
	return table.Styles{
		Header: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.Border)).Padding(0, 1),
		// Cell carries no Foreground: its reset code would cut Selected's
		// Background at every cell boundary, highlighting only the first column.
		// Padding emits no ANSI codes, so the highlight spans the whole row.
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: selected,
	}
}

// contrastText picks black or white text readable on bg by perceived
// luminance, since Selection ranges from near-black to pale cream across
// themes. Non-hex values (ANSI codes such as ThemeTerminal's) get white.
func contrastText(bg string) color.Color {
	if len(bg) == 7 && bg[0] == '#' {
		r, rErr := strconv.ParseInt(bg[1:3], 16, 64)
		g, gErr := strconv.ParseInt(bg[3:5], 16, 64)
		b, bErr := strconv.ParseInt(bg[5:7], 16, 64)
		if rErr == nil && gErr == nil && bErr == nil {
			luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			if luminance > 140 {
				return lipgloss.Color("#000000")
			}
		}
	}
	return lipgloss.Color("#FFFFFF")
}

func (m *browserModel) Init() tea.Cmd { return nil }

// Update handles navigation directly via table.Model's MoveUp/MoveDown
// rather than forwarding to table.Model.Update - one less layer (and its
// focus-state gate) between a keypress and the cursor actually moving.
func (m *browserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up":
		m.tbl.MoveUp(1)
	case "down":
		m.tbl.MoveDown(1)
	}
	return m, nil
}

// View renders inline, without AltScreen. The border sizes itself to the
// table's content, so there is no manual width/height math to drift.
func (m *browserModel) View() tea.View {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.colors.Border))
	return tea.NewView(box.Render(m.tbl.View()))
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
	m := newBrowserModel(theme.ResolveColors(th, theme.DarkTerminal()))
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run theme browser: %w", err)
	}
	return nil
}
