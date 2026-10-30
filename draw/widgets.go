package draw

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
)

// More is the row a scrolled list shows in place of its last while n rows
// hide below, set in indent cells to line up with the rows' text.
func More(n, indent int, st lipgloss.Style) string {
	return st.Render(fmt.Sprintf("%s↓ %d more", strings.Repeat(" ", max(indent, 0)), n))
}

// ButtonStyle paints a button: Focused the one enter would press, Blurred the
// rest. The theme fills both from its palette, with the same padding so focus
// never moves a button.
type ButtonStyle struct {
	Focused, Blurred lipgloss.Style
}

// Button is a filled label in st.Focused when focused, st.Blurred otherwise.
// Dialogs draw their answers with it; a form draws its submit with it.
func Button(label string, st ButtonStyle, focused bool) string {
	s := st.Blurred
	if focused {
		s = st.Focused
	}
	return s.Render("  " + label + "  ")
}

// Card is a Panel whose last row is footer, set off from the body by a rule:
// a block of facts with what can be done about them underneath.
func Card(r layout.Rect, title, body, footer string, st PanelStyle) string {
	if footer == "" {
		return Panel(r, Edges{Title: title}, body, st)
	}
	inner := InnerRect(r)
	rows := max(0, inner.H-2) // the rule and the footer take the last two rows
	lines := strings.Split(FitBlock(body, inner.W, rows), "\n")
	lines = lines[:min(len(lines), rows)]
	for len(lines) < rows { // a short body still leaves the footer at the bottom
		lines = append(lines, "")
	}
	lines = append(lines, Divider(inner.W, "", DividerStyle{Line: st.Border, Label: st.Title}), footer)
	return Panel(r, Edges{Title: title}, strings.Join(lines, "\n"), st)
}
