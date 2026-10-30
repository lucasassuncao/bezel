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
	lo, hi, before, after := tabSpan(parts, active, width, st)
	return Fit(before+strings.Join(parts[lo:hi], " ")+after, width)
}

// TabAt is the tab Tabs draws at column x with the same arguments, -1 over
// a gap, a scroll mark or past the last tab.
func TabAt(width int, names []string, active int, st ChromeStyle, x int) int {
	parts := tabParts(names, active, st)
	lo, hi, before, _ := tabSpan(parts, active, width, st)
	at := ansi.StringWidth(before)
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

// tabSpan is the parts [lo, hi) a row width wide shows, all when they fit,
// else as many as fit around active with ‹ or › where tabs are hidden on
// that side, so the active tab is never cut off.
func tabSpan(parts []string, active, width int, st ChromeStyle) (lo, hi int, before, after string) {
	if ansi.StringWidth(strings.Join(parts, " ")) <= width || active < 0 || active >= len(parts) {
		return 0, len(parts), "", ""
	}
	lo, hi = tabWindow(parts, active, width)
	if lo > 0 {
		before = st.Tab.Render("‹") + " "
	}
	if hi < len(parts) {
		after = " " + st.Tab.Render("›")
	}
	return lo, hi, before, after
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

// DividerStyle paints a Divider: Line the rule, Label the text set into it.
type DividerStyle struct {
	Line, Label lipgloss.Style
}

// Divider is a horizontal rule width cells long, with label set into it after a
// short lead when one is given: "── Label ─────".
func Divider(width int, label string, st DividerStyle) string {
	if width < 1 {
		return ""
	}
	const lead = 2
	// Too narrow for the lead, the two spaces and a cell of label: the bare rule.
	if label == "" || width < lead+3 {
		return st.Line.Render(strings.Repeat("─", width))
	}
	// Measured rendered: a style with padding takes cells the raw label does not.
	room := width - lead - 2
	pad := ansi.StringWidth(st.Label.Render("x")) - 1
	styled := st.Label.Render(Truncate(label, max(1, room-pad)))
	rest := max(0, room-ansi.StringWidth(styled))
	return st.Line.Render(strings.Repeat("─", lead)+" ") + styled + st.Line.Render(" "+strings.Repeat("─", rest))
}

// StatusLine is one row of text in a style, exactly width cells.
func StatusLine(width int, text string, style lipgloss.Style) string {
	return Fit(style.Render(OneLine(text)), width)
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
