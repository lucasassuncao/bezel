package draw

import (
	"strings"
	"testing"
	"unicode"
)

// Fit is what every pane cell goes through: whatever the text, the cell is
// exactly width wide, or the columns beside it shift.
func FuzzFitIsExactlyWidth(f *testing.F) {
	for _, s := range []string{"", "plain", "a\x1b[31mred\x1b[0m", "日本語", "e\u0301", "tab\there", "\xff", "0\xe8"} {
		f.Add(s, uint8(5))
	}
	f.Fuzz(func(t *testing.T, s string, w uint8) {
		s = Sanitize(s)
		// Format runes such as U+0600 glue to the next cluster, padding
		// included: the width is the terminal's guess, not a promise.
		if strings.ContainsFunc(s, func(r rune) bool { return unicode.Is(unicode.Cf, r) }) {
			t.Skip()
		}
		width := int(w)
		got := Fit(s, width)
		if Width(got) != width {
			t.Fatalf("Fit(%q, %d) = %q is %d wide", s, width, got, Width(got))
		}
		if Width(Truncate(s, width)) > max(width, 0) && width > 0 {
			t.Fatalf("Truncate(%q, %d) overflows", s, width)
		}
	})
}

// A modal drawn over a screen changes the cells it covers and nothing else:
// the line count stays, and rows it does not reach are left byte for byte.
func FuzzCompositeKeepsTheBackground(f *testing.F) {
	f.Add("[ok]", 3, 1)
	f.Add("a\nb\nc", -2, -1)
	f.Add("日本", 9, 2)
	f.Fuzz(func(t *testing.T, fg string, x, y int) {
		fg = strings.Join(SanitizeLines(fg), "\n")
		x, y = x%20, y%20
		bg := strings.TrimSuffix(strings.Repeat(strings.Repeat(".", 10)+"\n", 6), "\n")
		got := strings.Split(Composite(fg, bg, x, y), "\n")
		bgLines := strings.Split(bg, "\n")
		if len(got) != len(bgLines) {
			t.Fatalf("Composite(%q, x=%d, y=%d) has %d lines, want %d", fg, x, y, len(got), len(bgLines))
		}
		fgRows := strings.Count(fg, "\n") + 1
		for i := range got {
			if (i < y || i >= y+fgRows) && got[i] != bgLines[i] {
				t.Fatalf("Composite(%q, x=%d, y=%d) changed row %d: %q", fg, x, y, i, got[i])
			}
		}
	})
}
