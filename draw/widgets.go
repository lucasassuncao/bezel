package draw

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
)

// Button is a filled label: ink's colour behind it when focused, grey otherwise.
// Dialogs draw their answers with it; a form draws its submit with it.
func Button(label string, ink lipgloss.Style, focused bool) string {
	bg := lipgloss.Color("240")
	if focused {
		bg = ink.GetForeground()
	}
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(bg).Padding(0, 2).Render("  " + label + "  ")
}

// Card is a Panel whose last row is footer, set off from the body by a rule:
// a block of facts with what can be done about them underneath.
func Card(r layout.Rect, title, body, footer string, st PanelStyle) string {
	if footer == "" {
		return Panel(r, title, body, st)
	}
	inner := InnerRect(r)
	rows := max(0, inner.H-2) // the rule and the footer take the last two rows
	lines := strings.Split(FitBlock(body, inner.W, rows), "\n")
	lines = lines[:min(len(lines), rows)]
	for len(lines) < rows { // a short body still leaves the footer at the bottom
		lines = append(lines, "")
	}
	lines = append(lines, Divider(inner.W, "", st.Border, st.Title), footer)
	return Panel(r, title, strings.Join(lines, "\n"), st)
}
