// Package draw renders the frame: text fitting, panels, header, tabs, status
// and overlay compositing. Everything measures screen cells, truncates, and
// never wraps, so callers can count rows and get the terminal's answer.
package draw

import (
	"fmt"
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
// spaces and hyphens and cutting a word wider than a row. Newlines stay;
// style the rows after wrapping, since a style is not re-opened per row.
func Wrap(s string, width int) []string {
	if s == "" {
		return nil
	}
	if width > 0 {
		s = ansi.Wrap(s, width, "")
	}
	return strings.Split(s, "\n")
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
