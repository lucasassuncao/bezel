package draw

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// Sanitize renders server-supplied text so drawing it can only draw: OSC 52
// sets the clipboard, CSI forges a confirmation modal, bidi (CVE-2021-42574)
// reorders a line. Rendering only - the clipboard still sees the real bytes.
func Sanitize(s string) string {
	if utf8.ValidString(s) && !strings.ContainsFunc(s, unsafeToDraw) {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for i, r := range s {
		switch {
		case invalidAt(s, i):
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case !unsafeToDraw(r):
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == 0x1b:
			b.WriteString(`\e`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\x%02x`, r)
		default:
			fmt.Fprintf(&b, `\u%04x`, r)
		}
	}
	return b.String()
}

// Clean is s with only what a terminal draws: text, line breaks, tabs,
// colours and links. Other controls and escapes go, and those cut short with
// what they swallowed; bad UTF-8 is U+FFFD, the one cell a terminal draws.
func Clean(s string) string {
	s = strings.ToValidUTF8(s, "�")
	var b strings.Builder
	start := 0 // the first byte not yet copied to b
	for i := 0; i < len(s); {
		c := s[i]
		c1 := c == 0xc2 && i+1 < len(s) && s[i+1] <= 0x9f // U+0080 to U+009F
		if (c >= 0x20 && c != 0x7f && !c1) || c == '\t' || c == '\n' {
			i++
			continue
		}
		n, keep := 1, false
		switch {
		case c1:
			n = 2
		case c == 0x1b:
			n, keep = escape(s[i:])
		}
		if !keep {
			b.WriteString(s[start:i])
			start = i + n
		}
		i += n
	}
	if start == 0 {
		return s
	}
	b.WriteString(s[start:])
	return strings.ToValidUTF8(b.String(), "�")
}

// escape is how many bytes the escape at the start of s takes, and whether it
// is a whole colour or link, the only escapes Clean keeps.
func escape(s string) (int, bool) {
	p := ansi.GetParser()
	defer ansi.PutParser(p)
	seq, _, n, state := ansi.DecodeSequence(s, ansi.NormalState, p)
	cmd := p.Command()
	if state != ansi.NormalState || !whole(seq) {
		return n, false
	}
	return n, ansi.HasCsiPrefix(seq) && cmd == 'm' || ansi.HasOscPrefix(seq) && cmd == 8
}

// whole reports whether seq, decoded on its own, ends complete: the decoder
// hands back an escape cut short by a control character as if it were one.
func whole(seq string) bool {
	_, _, n, st := ansi.DecodeSequence(seq, ansi.NormalState, nil)
	return n == len(seq) && st == ansi.NormalState
}

// SanitizeLines handles legitimately multi-line text - a policy, a role's
// configuration - keeping the breaks and neutralising the rest. A trailing
// carriage return is CRLF and not content, so a Windows document reads clean.
func SanitizeLines(s string) []string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = Sanitize(strings.TrimSuffix(line, "\r"))
	}
	return lines
}

// invalidAt reports whether s[i] is a byte that is not UTF-8. The terminal
// draws it as one U+FFFD, but ansi measures it with the cells after it.
func invalidAt(s string, i int) bool {
	r, size := utf8.DecodeRuneInString(s[i:])
	return r == utf8.RuneError && size == 1
}

// unsafeToDraw reports whether a rune changes the terminal's behaviour or the
// reading order of the line, rather than putting a glyph on the screen.
func unsafeToDraw(r rune) bool {
	switch {
	case r < 0x20, r == 0x7f:
		// C0 controls and DEL: escapes, carriage returns, NUL.
		return true
	case r >= 0x80 && r <= 0x9f:
		// C1 controls. A UTF-8 terminal will not act on these, but not every
		// terminal is in UTF-8 mode, and the cost of covering them is nothing.
		return true
	case r == 0x200e, r == 0x200f:
		// LEFT-TO-RIGHT and RIGHT-TO-LEFT MARK.
		return true
	case r >= 0x202a && r <= 0x202e:
		// The embedding and override pair, and POP DIRECTIONAL FORMATTING.
		return true
	case r >= 0x2066 && r <= 0x2069:
		// The isolates, which do the same job with tidier scoping.
		return true
	default:
		return false
	}
}
