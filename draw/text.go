// Package draw renders the frame: text fitting, panels, header, tabs, status
// and overlay compositing. Everything measures screen cells, truncates, and
// never wraps, so callers can count rows and get the terminal's answer.
package draw

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/layout"
)

const ellipsis = "…"

// Truncate cuts s to width cells with an ellipsis. width <= 0 leaves s alone.
func Truncate(s string, width int) string {
	if width <= 0 || ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, ellipsis)
}

// Pad fills s out to width cells.
func Pad(s string, width int) string {
	if gap := width - ansi.StringWidth(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

// Fit makes s exactly width cells: cut when longer, padded when shorter.
func Fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return Pad(Truncate(s, width), width)
}

// Wrap folds plain text into rows of at most width cells, breaking after
// spaces and hyphens, cutting a word wider than a row; a rune wider than a row
// gets one alone. Newlines stay; style the rows after, as styles do not re-open.
func Wrap(s string, width int) []string {
	if s == "" {
		return nil
	}
	if width <= 0 {
		return strings.Split(s, "\n")
	}
	var rows []string
	for _, line := range strings.Split(s, "\n") {
		rows = append(rows, wrapLine(line, width)...)
	}
	return rows
}

// wrapLine is one line of Wrap. ansi.Wrap leaves a rune wider than the row
// with what follows it, and blank rows before it; those lines are cut instead.
func wrapLine(line string, width int) []string {
	rows := strings.Split(ansi.Wrap(line, width, ""), "\n")
	if !slices.ContainsFunc(rows, func(r string) bool { return ansi.StringWidth(r) > width }) {
		return rows
	}
	rows = slices.DeleteFunc(strings.Split(ansi.Hardwrap(line, width, true), "\n"), func(r string) bool { return r == "" })
	if len(rows) == 0 {
		return []string{""}
	}
	return rows
}

// BlockSize measures a rendered block: widest line and line count.
func BlockSize(s string) (width, height int) {
	lines := strings.Split(s, "\n")
	for _, l := range lines {
		width = max(width, ansi.StringWidth(l))
	}
	return width, len(lines)
}

// PadHeight grows a block to height lines so a Composite has rows to land on.
func PadHeight(s string, height int) string {
	if missing := height - strings.Count(s, "\n") - 1; missing > 0 {
		return s + strings.Repeat("\n", missing)
	}
	return s
}

// FitBlock cuts a block to height rows and width cells per row. A too-wide
// line would wrap inside lipgloss and push a border down, so cut first.
func FitBlock(s string, width, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.SplitN(s, "\n", height+1) // a long body is never split past what shows
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, l := range lines {
		lines[i] = Truncate(l, width)
	}
	return strings.Join(lines, "\n")
}

// InnerRect is what a Panel leaves for content: one border cell and one
// padding cell on each side horizontally, one border row above and below.
func InnerRect(r layout.Rect) layout.Rect {
	return layout.Rect{X: r.X + 2, Y: r.Y + 1, W: max(1, r.W-4), H: max(1, r.H-2)}
}

// GutterWidth is the columns Gutter and NumberLines take: "%4d │ ".
const GutterWidth = 7

// Gutter is one line-number cell, for a textarea prompt or NumberLines.
func Gutter(n int, style lipgloss.Style) string { return style.Render(fmt.Sprintf("%4d │ ", n)) }

// NumberLines prefixes every line of s with its 1-based Gutter, before any
// scrolling, so the numbers stay absolute.
func NumberLines(s string, style lipgloss.Style) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = Gutter(i+1, style) + line
	}
	return strings.Join(lines, "\n")
}

// TabStop is the cells between the tab stops ExpandTabs runs a tab to.
const TabStop = 4

// ExpandTabs turns each tab into the spaces up to the next TabStop, counting
// the cells before it, escapes left out. Width counts a tab as no cells, and
// the terminal would jump it to a tab stop of its own.
func ExpandTabs(s string) string {
	if !strings.Contains(s, "\t") {
		return s
	}
	parts := strings.Split(s, "\t")
	var b strings.Builder
	col := 0
	for i, p := range parts {
		b.WriteString(p)
		if i == len(parts)-1 {
			break
		}
		col += ansi.StringWidth(p)
		pad := TabStop - col%TabStop
		b.WriteString(strings.Repeat(" ", pad))
		col += pad
	}
	return b.String()
}

// Width is how many cells s takes on screen, escapes ignored.
func Width(s string) int { return ansi.StringWidth(s) }

// Cut is Truncate without the ellipsis: the line ends where the width does.
func Cut(s string, width int) string {
	if width <= 0 || ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, "")
}

// Heading is a block heading inside a pane, upper-cased and cut to width.
func Heading(name string, style lipgloss.Style, width int) string {
	return style.Render(Cut(strings.ToUpper(name), width))
}

// OneLine is s with its line breaks as spaces, rune for rune, for text that
// must keep to one row: a label, a status, a title.
func OneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, s)
}
