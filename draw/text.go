// Package draw renders the frame: text fitting, panels, header, tabs, status
// and overlay compositing. Everything measures screen cells, truncates, and
// never wraps, so callers can count rows and get the terminal's answer.
package draw

import (
	"fmt"
	"slices"
	"strings"
	"time"

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
	lines := strings.Split(s, "\n")
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

// Breadcrumb joins path segments the way a header subtitle shows them.
func Breadcrumb(segs []string) string { return strings.Join(segs, " › ") }

// ScrollTo returns a window of at most height lines from s that keeps line
// target (1-based) visible, roughly centred. target < 1 yields the top.
func ScrollTo(s string, height, target int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= height {
		return s
	}
	offset := min(max(target-1-height/2, 0), len(lines)-height)
	return strings.Join(lines[offset:offset+height], "\n")
}

// GutterWidth is the columns Gutter and NumberLines take: "%4d │ ".
const GutterWidth = 7

// Gutter is one line-number cell, for a textarea prompt or NumberLines.
func Gutter(n int, style lipgloss.Style) string { return style.Render(fmt.Sprintf("%4d │ ", n)) }

// NumberLines prefixes every line of s with its 1-based Gutter. Run it before
// ScrollTo so the numbers stay absolute.
func NumberLines(s string, style lipgloss.Style) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = Gutter(i+1, style) + line
	}
	return strings.Join(lines, "\n")
}

// Width is how many cells s takes on screen, escapes ignored.
func Width(s string) int { return ansi.StringWidth(s) }

// Tail is a copy of the last height lines of lines, the way a log scrolls.
// A copy, so holding it does not keep the whole log alive.
func Tail(lines []string, height int) []string {
	if height <= 0 {
		return nil
	}
	return slices.Clone(lines[max(len(lines)-height, 0):])
}

// Cut is Truncate without the ellipsis: the line ends where the width does.
func Cut(s string, width int) string {
	if width <= 0 || ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, "")
}

// OrDash is s, or a dash when it is empty.
func OrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// KV is one labelled row: the key padded to 14, the value after it, the
// whole line cut to width. Both halves are sanitized, since either may be
// something a server chose.
func KV(key, value string, keyStyle lipgloss.Style, width int) string {
	return Cut("  "+keyStyle.Render(fmt.Sprintf("%-14s ", Sanitize(key)))+Sanitize(value), width) + "\n"
}

// Heading is a block heading inside a pane, upper-cased and cut to width.
func Heading(name string, style lipgloss.Style, width int) string {
	return style.Render(Cut(strings.ToUpper(name), width))
}

// HeadingBadged is a heading wearing a badge beside it, or the heading alone
// when the two do not fit: a badge that says what the block is has to be
// read with it.
func HeadingBadged(name, badge string, style lipgloss.Style, width int) string {
	label := strings.ToUpper(name)
	if ansi.StringWidth(label)+1+ansi.StringWidth(badge) > width {
		return Heading(name, style, width)
	}
	return style.Render(label) + " " + badge
}

// ShortDuration renders a lifetime the way a header has room for: the two
// largest units that say anything, truncated rather than rounded, so a
// countdown never shows more time than is left.
func ShortDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	switch {
	case d >= 24*time.Hour:
		days := d / (24 * time.Hour)
		if h := (d % (24 * time.Hour)) / time.Hour; h > 0 {
			return fmt.Sprintf("%dd%dh", days, h)
		}
		return fmt.Sprintf("%dd", days)
	case d >= time.Hour:
		hours := d / time.Hour
		if mins := (d % time.Hour) / time.Minute; mins > 0 {
			return fmt.Sprintf("%dh%dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	case d >= time.Minute:
		return fmt.Sprintf("%dm", d/time.Minute)
	default:
		return fmt.Sprintf("%ds", d/time.Second)
	}
}
