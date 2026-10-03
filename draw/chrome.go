package draw

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/layout"
)

// ChromeStyle colours the rows around the panes.
type ChromeStyle struct {
	Title     lipgloss.Style // header title
	Info      lipgloss.Style // header subtitle and right-hand text
	Tab       lipgloss.Style
	TabActive lipgloss.Style
	Status    lipgloss.Style
	Dim       lipgloss.Style // empty-state message and hint
}

// Header is one row: title and subtitle on the left, right pinned to the far
// edge, the whole thing cut to width.
func Header(width int, title, subtitle, right string, st ChromeStyle) string {
	left := st.Title.Render(title)
	if subtitle != "" {
		left += st.Info.Render(" · " + subtitle)
	}
	rightR := ""
	if right != "" {
		rightR = st.Info.Render(right)
	}
	gap := max(0, width-ansi.StringWidth(left)-ansi.StringWidth(rightR))
	return Fit(left+strings.Repeat(" ", gap)+rightR, width)
}

// Tabs is one row of names, the active one styled apart.
func Tabs(width int, names []string, active int, st ChromeStyle) string {
	parts := make([]string, len(names))
	for i, n := range names {
		if i == active {
			parts[i] = st.TabActive.Render(" " + n + " ")
		} else {
			parts[i] = st.Tab.Render(" " + n + " ")
		}
	}
	return Fit(strings.Join(parts, " "), width)
}

// Divider is a horizontal rule width cells long, with label set into it after a
// short lead when one is given: "── Label ─────".
func Divider(width int, label string, line, text lipgloss.Style) string {
	if width < 1 {
		return ""
	}
	const lead = 2
	// Too narrow for the lead, the two spaces and a cell of label: the bare rule.
	if label == "" || width < lead+3 {
		return line.Render(strings.Repeat("─", width))
	}
	// Measured rendered: a style with padding takes cells the raw label does not.
	room := width - lead - 2
	pad := ansi.StringWidth(text.Render("x")) - 1
	styled := text.Render(Truncate(label, max(1, room-pad)))
	rest := max(0, room-ansi.StringWidth(styled))
	return line.Render(strings.Repeat("─", lead)+" ") + styled + line.Render(" "+strings.Repeat("─", rest))
}

// StatusLine is one row of text in a style, exactly width cells.
func StatusLine(width int, text string, style lipgloss.Style) string {
	return Fit(style.Render(text), width)
}

// EmptyState centres a message, and a hint under it, inside r.
func EmptyState(r layout.Rect, message, hint string, st ChromeStyle) string {
	if r.W < 1 || r.H < 1 {
		return ""
	}
	wrap := lipgloss.NewStyle().Width(r.W).Align(lipgloss.Center)
	body := wrap.Render(st.Dim.Render(message))
	if hint != "" {
		body += "\n\n" + wrap.Render(st.Dim.Render(hint))
	}
	return FitBlock(lipgloss.Place(r.W, r.H, lipgloss.Center, lipgloss.Center, body), r.W, r.H)
}
