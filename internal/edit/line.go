// Package edit is the text editing input and textbox share: a line of
// graphemes with a caret between them. Every operation returns a new Line
// and never writes to the slice a copy of the old one still holds.
package edit

import (
	"math"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// Line is one line of text and the caret in it, which sits before the
// grapheme at Cursor; at Len it is after the last one. It keeps the text
// whole and where each grapheme starts, four bytes each, not a string apiece.
type Line struct {
	s      string
	b      []int32 // each grapheme's start in s, then len(s); nil when s is empty
	cursor int
}

// bounds is where each grapheme of s starts, then len(s): what a user sees
// as one character, a flag, an emoji with its modifiers or an accented letter.
func bounds(s string) []int32 {
	if s == "" {
		return nil
	}
	b := make([]int32, 1, len(s)+1)
	for off := 0; off < len(s); {
		c, _ := ansi.FirstGraphemeCluster(s[off:], ansi.GraphemeWidth)
		off += max(len(c), 1) // never loop forever on a cluster that came back empty
		b = append(b, at32(off))
	}
	return b
}

// Cells is how many cells grapheme g takes. A printable ASCII byte is one,
// answered without the full measure, which costs a microsecond a call.
func Cells(g string) int {
	if len(g) == 1 && g[0] >= 0x20 && g[0] < 0x7f {
		return 1
	}
	return ansi.StringWidth(g)
}

// at32 is a byte offset as Line keeps it. No line typed or pasted nears
// 2GiB, but one past it is clamped there rather than wrapped to a negative.
func at32(off int) int32 {
	switch {
	case off > math.MaxInt32:
		return math.MaxInt32
	case off < 0:
		return 0
	}
	return int32(off)
}

// Graphemes splits s into what a user sees as characters.
func Graphemes(s string) []string {
	b := bounds(s)
	out := make([]string, max(len(b)-1, 0))
	for i := range out {
		out[i] = s[b[i]:b[i+1]]
	}
	return out
}

// NewLine is s with the caret at its end. s is copied: a line cut from a
// large text would otherwise keep the whole text alive.
func NewLine(s string) Line {
	return rebuild(strings.Clone(s), "")
}

// String is the text of the line.
func (l Line) String() string { return l.s }

// Len is how many graphemes the line has.
func (l Line) Len() int { return max(len(l.b)-1, 0) }

// Cursor is the index of the grapheme the caret sits before.
func (l Line) Cursor() int { return l.cursor }

// off is the byte where grapheme i starts, i kept within the line.
func (l Line) off(i int) int {
	if l.b == nil {
		return 0
	}
	return int(l.b[l.clamp(i)])
}

// At is the grapheme at i, or "" past the end.
func (l Line) At(i int) string {
	if i < 0 || i >= l.Len() {
		return ""
	}
	return l.s[l.b[i]:l.b[i+1]]
}

// Slice is the text of the graphemes [from, to).
func (l Line) Slice(from, to int) string {
	from, to = l.clamp(from), l.clamp(to)
	if from >= to {
		return ""
	}
	return l.s[l.off(from):l.off(to)]
}

// Width is how many cells the graphemes [from, to) take.
func (l Line) Width(from, to int) int { return ansi.StringWidth(l.Slice(from, to)) }

// IndexAt is the grapheme whose cells hold column col, counted from the
// start: the caret a click or a move up lands on. Past the end it is Len.
func (l Line) IndexAt(col int) int {
	at := 0
	for i := range l.Len() {
		w := Cells(l.At(i))
		if col < at+w {
			return i
		}
		at += w
	}
	return l.Len()
}

func (l Line) clamp(i int) int { return min(max(i, 0), l.Len()) }

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
	at := l.off(l.cursor)
	return rebuild(l.s[:at]+s, l.s[at:])
}

// remove drops the graphemes [from, to) and leaves the caret at from.
func (l Line) remove(from, to int) Line {
	from, to = l.clamp(from), l.clamp(to)
	if from >= to {
		return l.SetCursor(from)
	}
	return rebuild(l.s[:l.off(from)], l.s[l.off(to):])
}

// rebuild is head and tail split afresh, the caret after the last grapheme
// that ends within head: an edit can join or part its neighbours, as a mark
// typed after a letter does.
func rebuild(head, tail string) Line {
	s := head + tail
	b := bounds(s)
	cursor, found := slices.BinarySearch(b, at32(len(head)))
	if !found {
		cursor-- // head ends inside a grapheme: the caret goes before it
	}
	return Line{s: s, b: b, cursor: max(cursor, 0)}
}

// DeleteBack removes the grapheme before the caret.
func (l Line) DeleteBack() Line { return l.remove(l.cursor-1, l.cursor) }

// DeleteForward removes the grapheme after the caret.
func (l Line) DeleteForward() Line { return l.remove(l.cursor, l.cursor+1) }

// KillToStart removes everything before the caret.
func (l Line) KillToStart() Line { return l.remove(0, l.cursor) }

// KillToEnd removes everything after the caret.
func (l Line) KillToEnd() Line { return l.remove(l.cursor, l.Len()) }

// DeleteWordBack removes from the start of the word before the caret.
func (l Line) DeleteWordBack() Line { return l.remove(l.wordLeft(), l.cursor) }

// DeleteWordForward removes to the end of the word after the caret.
func (l Line) DeleteWordForward() Line { return l.remove(l.cursor, l.wordRight()) }

// WordLeft moves the caret to the start of the word before it.
func (l Line) WordLeft() Line { return l.SetCursor(l.wordLeft()) }

// WordRight moves the caret to the end of the word after it.
func (l Line) WordRight() Line { return l.SetCursor(l.wordRight()) }

func (l Line) space(i int) bool {
	r, _ := utf8.DecodeRuneInString(l.At(i))
	return unicode.IsSpace(r)
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
	i, n := l.cursor, l.Len()
	for i < n && l.space(i) {
		i++
	}
	for i < n && !l.space(i) {
		i++
	}
	return i
}

// MapWord passes the word after the caret, spaces before it skipped, through
// f, and leaves the caret after it, as alt+u does in a shell.
func (l Line) MapWord(f func(string) string) Line {
	from := l.cursor
	for from < l.Len() && l.space(from) {
		from++
	}
	to := l.SetCursor(from).wordRight()
	if from >= to {
		return l.SetCursor(to)
	}
	return rebuild(l.s[:l.off(from)]+f(l.s[l.off(from):l.off(to)]), l.s[l.off(to):])
}

// Transpose swaps the characters either side of the caret, the last two at
// the end of the line, and moves the caret past them, as ctrl+t does.
func (l Line) Transpose() Line {
	c := l.cursor
	if c == l.Len() {
		c--
	}
	if c < 1 || c >= l.Len() {
		return l
	}
	a, b := l.At(c-1), l.At(c)
	return rebuild(l.s[:l.off(c-1)]+b+a, l.s[l.off(c+1):])
}

// WordAt is the word the caret is in or touches, "" between spaces.
func (l Line) WordAt() string {
	from, to := l.cursor, l.cursor
	for from > 0 && !l.space(from-1) {
		from--
	}
	for to < l.Len() && !l.space(to) {
		to++
	}
	return l.Slice(from, to)
}

// Split cuts the line at the caret: what was before it, and what was after,
// with the caret at its start. Each half is a copy of its own text.
func (l Line) Split() (before, after Line) {
	at := l.off(l.cursor)
	return rebuild(strings.Clone(l.s[:at]), ""), rebuild("", strings.Clone(l.s[at:]))
}

// Join appends next to the line, the caret where the two meet.
func (l Line) Join(next Line) Line { return rebuild(l.s, next.s) }

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
