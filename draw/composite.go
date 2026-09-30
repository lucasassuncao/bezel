package draw

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Composite draws fg over bg at (x, y), replacing only the cells it covers and
// keeping the rest of each bg line, escapes included. This is what lets a
// modal float over the panes instead of blanking them.
func Composite(fg, bg string, x, y int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	for i, fgLine := range fgLines {
		row := y + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		bgLine := bgLines[row]

		left := ansi.Truncate(bgLine, x, "")
		if pad := x - ansi.StringWidth(left); pad > 0 {
			left += strings.Repeat(" ", pad)
		}
		right := ansi.TruncateLeft(bgLine, x+ansi.StringWidth(fgLine), "")

		bgLines[row] = left + fgLine + right
	}
	return strings.Join(bgLines, "\n")
}

// CompositeCenter centres fg over bg.
func CompositeCenter(fg, bg string) string {
	fgW, fgH := BlockSize(fg)
	bgW, bgH := BlockSize(bg)
	return Composite(fg, bg, max(0, (bgW-fgW)/2), max(0, (bgH-fgH)/2))
}
