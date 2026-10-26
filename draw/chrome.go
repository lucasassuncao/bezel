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
	parts := tabParts(names, active, st)
	all := strings.Join(parts, " ")
	if ansi.StringWidth(all) <= width || active < 0 || active >= len(parts) {
		return Fit(all, width)
	}
	return Fit(scrolledTabs(parts, active, width, st), width)
}

// TabAt is the tab Tabs draws at column x with the same arguments, -1 over
// a gap, a scroll mark or past the last tab.
func TabAt(width int, names []string, active int, st ChromeStyle, x int) int {
	parts := tabParts(names, active, st)
	lo, hi, at := 0, len(parts), 0
	if ansi.StringWidth(strings.Join(parts, " ")) > width && active >= 0 && active < len(parts) {
		if lo, hi = tabWindow(parts, active, width); lo > 0 {
			at = 2 // "‹ "
		}
	}
	for i := lo; i < hi; i++ {
		w := ansi.StringWidth(parts[i])
		if x >= at && x < at+w && x < width {
			return i
		}
		at += w + 1
	}
	return -1
}

// tabParts is each name drawn as its tab.
func tabParts(names []string, active int, st ChromeStyle) []string {
	parts := make([]string, len(names))
	for i, n := range names {
		if i == active {
			parts[i] = st.TabActive.Render(" " + n + " ")
		} else {
			parts[i] = st.Tab.Render(" " + n + " ")
		}
	}
	return parts
}

// scrolledTabs is as many of parts as fit width around active, with ‹ or ›
// where tabs are hidden on that side, so the active tab is never cut off.
func scrolledTabs(parts []string, active, width int, st ChromeStyle) string {
	lo, hi := tabWindow(parts, active, width)
	s := strings.Join(parts[lo:hi], " ")
	if lo > 0 {
		s = st.Tab.Render("‹") + " " + s
	}
	if hi < len(parts) {
		s += " " + st.Tab.Render("›")
	}
	return s
}

// tabWindow is the parts [lo, hi) that fit width around active, room kept
// for the marks of the hidden ones.
func tabWindow(parts []string, active, width int) (lo, hi int) {
	lo, hi = active, active+1
	fits := func(lo, hi int) bool {
		w := ansi.StringWidth(strings.Join(parts[lo:hi], " "))
		if lo > 0 {
			w += 2
		}
		if hi < len(parts) {
			w += 2
		}
		return w <= width
	}
	for grew := true; grew; {
		grew = false
		if hi < len(parts) && fits(lo, hi+1) {
			hi, grew = hi+1, true
		}
		if lo > 0 && fits(lo-1, hi) {
			lo, grew = lo-1, true
		}
	}
	return lo, hi
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
