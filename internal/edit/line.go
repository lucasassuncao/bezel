// Package edit is the text editing input and textbox share: a line of
// graphemes with a caret between them. Every operation returns a new Line
// and never writes to the slice a copy of the old one still holds.
package edit

import (
	"slices"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

// Line is one line of text and the caret in it, which sits before the
// grapheme at Cursor; at Len it is after the last one.
type Line struct {
	g      []string
	cursor int
}

// Graphemes splits s into what a user sees as characters: a flag, an emoji
// with its modifiers or a letter with its accents is one.
func Graphemes(s string) []string {
	var out []string
	for s != "" {
		c, _ := ansi.FirstGraphemeCluster(s, ansi.GraphemeWidth)
		if c == "" { // not expected; never loop forever on it
			c = s[:1]
		}
		out = append(out, c)
		s = s[len(c):]
	}
	return out
}

// NewLine is s with the caret at its end.
func NewLine(s string) Line {
	g := Graphemes(s)
	return Line{g: g, cursor: len(g)}
}

// String is the text of the line.
func (l Line) String() string { return strings.Join(l.g, "") }

// Len is how many graphemes the line has.
func (l Line) Len() int { return len(l.g) }

// Cursor is the index of the grapheme the caret sits before.
func (l Line) Cursor() int { return l.cursor }

// At is the grapheme at i, or "" past the end.
func (l Line) At(i int) string {
	if i < 0 || i >= len(l.g) {
		return ""
	}
	return l.g[i]
}

// Slice is the text of the graphemes [from, to).
func (l Line) Slice(from, to int) string {
	from, to = l.clamp(from), l.clamp(to)
	if from >= to {
		return ""
	}
	return strings.Join(l.g[from:to], "")
}

// Width is how many cells the graphemes [from, to) take.
func (l Line) Width(from, to int) int { return ansi.StringWidth(l.Slice(from, to)) }

// IndexAt is the grapheme whose cells hold column col, counted from the
// start: the caret a click or a move up lands on. Past the end it is Len.
func (l Line) IndexAt(col int) int {
	at := 0
	for i, g := range l.g {
		w := ansi.StringWidth(g)
		if col < at+w {
			return i
		}
		at += w
	}
	return len(l.g)
}

func (l Line) clamp(i int) int { return min(max(i, 0), len(l.g)) }

// SetCursor moves the caret to i, kept within the line.
func (l Line) SetCursor(i int) Line {
	l.cursor = l.clamp(i)
	return l
}

// Insert puts s at the caret and moves the caret past it. s is taken as it
// is: a caller that wants one line strips its newlines first.
func (l Line) Insert(s string) Line {
	if s == "" {
		return l
	}
	return rebuild(l.Slice(0, l.cursor)+s, l.Slice(l.cursor, len(l.g)))
}

// remove drops the graphemes [from, to) and leaves the caret at from.
func (l Line) remove(from, to int) Line {
	from, to = l.clamp(from), l.clamp(to)
	if from >= to {
		return l.SetCursor(from)
	}
	return rebuild(l.Slice(0, from), l.Slice(to, len(l.g)))
}

// rebuild is head and tail split afresh, the caret after the last grapheme
// that ends within head: an edit can join or part its neighbours, as a mark
// typed after a letter does.
func rebuild(head, tail string) Line {
	g := Graphemes(head + tail)
	cursor, n := 0, 0
	for i, c := range g {
		if n += len(c); n > len(head) {
			break
		}
		cursor = i + 1
	}
	return Line{g: g, cursor: cursor}
}

// DeleteBack removes the grapheme before the caret.
func (l Line) DeleteBack() Line { return l.remove(l.cursor-1, l.cursor) }

// DeleteForward removes the grapheme after the caret.
func (l Line) DeleteForward() Line { return l.remove(l.cursor, l.cursor+1) }

// KillToStart removes everything before the caret.
func (l Line) KillToStart() Line { return l.remove(0, l.cursor) }

// KillToEnd removes everything after the caret.
func (l Line) KillToEnd() Line { return l.remove(l.cursor, len(l.g)) }

// DeleteWordBack removes from the start of the word before the caret.
func (l Line) DeleteWordBack() Line { return l.remove(l.wordLeft(), l.cursor) }

// DeleteWordForward removes to the end of the word after the caret.
func (l Line) DeleteWordForward() Line { return l.remove(l.cursor, l.wordRight()) }

// WordLeft moves the caret to the start of the word before it.
func (l Line) WordLeft() Line { return l.SetCursor(l.wordLeft()) }

// WordRight moves the caret to the end of the word after it.
func (l Line) WordRight() Line { return l.SetCursor(l.wordRight()) }

func (l Line) space(i int) bool {
	r := []rune(l.g[i])
	return len(r) > 0 && unicode.IsSpace(r[0])
}

// wordLeft passes the spaces before the caret, then the word before them.
func (l Line) wordLeft() int {
	i := l.cursor
	for i > 0 && l.space(i-1) {
		i--
	}
	for i > 0 && !l.space(i-1) {
		i--
	}
	return i
}

// wordRight passes the spaces after the caret, then the word after them.
func (l Line) wordRight() int {
	i := l.cursor
	for i < len(l.g) && l.space(i) {
		i++
	}
	for i < len(l.g) && !l.space(i) {
		i++
	}
	return i
}

// Split cuts the line at the caret: what was before it, and what was after,
// with the caret at its start.
func (l Line) Split() (before, after Line) {
	before = Line{g: slices.Clone(l.g[:l.cursor]), cursor: l.cursor}
	after = Line{g: slices.Clone(l.g[l.cursor:])}
	return before, after
}

// Join appends next to the line, the caret where the two meet.
func (l Line) Join(next Line) Line { return rebuild(l.String(), next.String()) }

// Printable is s without the characters that would draw nothing or redraw
// the rest: controls, and format marks bar the joiners emoji and flags need.
// Bidi overrides go with them, so no text can reorder what follows it.
func Printable(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\u200c', r == '\u200d', r >= 0xE0020 && r <= 0xE007F:
			return r
		case unicode.IsControl(r), unicode.Is(unicode.Cf, r):
			return -1
		}
		return r
	}, s)
}
