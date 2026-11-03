package input

import (
	"strings"
	"unicode"

	"github.com/lucasassuncao/bezel/internal/edit"
)

// slot is one place of a format: a class a typed character must be of, or a
// literal the field writes itself.
type slot struct {
	class rune // '9' a digit, 'a' a letter, '*' any; 0 for a literal
	lit   string
}

// parseFormat reads a pattern: 9, a and * are slots, a backslash makes the
// next character a literal, and every other character is one.
func parseFormat(pattern string) []slot {
	var out []slot
	escaped := false
	for _, g := range edit.Graphemes(pattern) {
		switch {
		case escaped:
			out, escaped = append(out, slot{lit: g}), false
		case g == `\`:
			escaped = true
		case g == "9" || g == "a" || g == "*":
			out = append(out, slot{class: rune(g[0])})
		default:
			out = append(out, slot{lit: g})
		}
	}
	return out
}

// takes reports whether g fits the slot.
func (s slot) takes(g string) bool {
	r := []rune(g)[0]
	switch s.class {
	case '9':
		return unicode.IsDigit(r)
	case 'a':
		return unicode.IsLetter(r)
	}
	return !unicode.IsSpace(r)
}

// WithFormat holds the text to pattern, as for a date or a document number:
// 9 takes a digit, a a letter and * anything, while the rest is written for
// the user. "9999-99-99" is a date; a backslash makes 9, a or * literal.
func (m Model) WithFormat(pattern string) Model {
	m.format = parseFormat(pattern)
	return m.SetValue(m.Value())
}

// Complete reports whether every slot of WithFormat is filled; without a
// format, whether there is any text.
func (m Model) Complete() bool {
	if len(m.format) == 0 {
		return m.line.Len() > 0
	}
	return m.line.Len() == len(m.format)
}

// formatted is s laid on the format: each slot takes the next character of s
// that fits it, skipping the ones that do not, and a literal is written once
// there is more of s to place, taking a typed copy of itself along.
func (m *Model) formatted(s string) string {
	if len(m.format) == 0 {
		return s
	}
	in := edit.Graphemes(s)
	var b strings.Builder
	for _, sl := range m.format {
		if len(in) == 0 {
			break
		}
		if sl.class == 0 {
			b.WriteString(sl.lit)
			if in[0] == sl.lit {
				in = in[1:]
			}
			continue
		}
		for len(in) > 0 && !sl.takes(in[0]) {
			in = in[1:]
		}
		if len(in) == 0 {
			break
		}
		b.WriteString(in[0])
		in = in[1:]
	}
	return b.String()
}

// hint is what the format still asks for past the text: each slot as an
// underscore, each literal as itself.
func (m *Model) hint() []string {
	if len(m.format) == 0 || m.line.Cursor() != m.line.Len() {
		return nil
	}
	var out []string
	for _, sl := range m.format[min(m.line.Len(), len(m.format)):] {
		if sl.class == 0 {
			out = append(out, sl.lit)
		} else {
			out = append(out, "_")
		}
	}
	return out
}
