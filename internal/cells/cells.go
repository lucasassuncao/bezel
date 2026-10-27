// Package cells reads a drawn block back into styled cells and draws them
// again, for the animations that change a block after it was drawn.
package cells

import (
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Attr is a cell's text attributes, one bit each.
type Attr uint8

const (
	Bold Attr = 1 << iota
	Faint
	Italic
	Underline
	Blink
	Reverse
	Strike
)

// Cell is one grapheme of a drawn block and how it was drawn.
type Cell struct {
	Text   string
	Width  int
	Fg, Bg color.Color // nil is the terminal's own
	Attrs  Attr
}

// Grid is a drawn block read back into cells, a row per line.
type Grid [][]Cell

// Parse reads block, as lipgloss draws it, into cells. Colours and attributes
// set by SGR follow the text; every other escape sequence draws nothing and
// is dropped.
func Parse(block string) Grid {
	p := ansi.NewParser()
	var g Grid
	row := []Cell{}
	var cur Cell
	state := byte(0)
	for len(block) > 0 {
		seq, width, n, next := ansi.DecodeSequence(block, state, p)
		state, block = next, block[n:]
		switch {
		case seq == "\n":
			g, row = append(g, row), []Cell{}
		case width > 0:
			c := cur
			c.Text, c.Width = seq, width
			row = append(row, c)
		case strings.HasPrefix(seq, "\x1b[") && ansi.Cmd(p.Command()).Final() == 'm':
			cur = sgr(cur, p.Params())
		}
	}
	return append(g, row)
}

// setAttr and clearAttr are the attributes SGR codes 1-9 turn on and 20-29
// turn off, the latter indexed from 20.
var (
	setAttr   = [10]Attr{1: Bold, 2: Faint, 3: Italic, 4: Underline, 5: Blink, 7: Reverse, 9: Strike}
	clearAttr = [10]Attr{3: Italic, 4: Underline, 5: Blink, 7: Reverse, 9: Strike}
)

// sgr is style changed by one Select Graphic Rendition sequence's params.
func sgr(style Cell, params ansi.Params) Cell {
	if len(params) == 0 {
		return Cell{}
	}
	for i := 0; i < len(params); i++ {
		switch code := params[i].Param(0); {
		case code == 0:
			style = Cell{}
		case code >= 1 && code <= 9:
			style.Attrs |= setAttr[code]
		case code == 22:
			style.Attrs &^= Bold | Faint
		case code >= 23 && code <= 29:
			style.Attrs &^= clearAttr[code-20]
		case code >= 30 && code <= 37:
			style.Fg = ansi.BasicColor(code - 30)
		case code >= 40 && code <= 47:
			style.Bg = ansi.BasicColor(code - 40)
		case code >= 90 && code <= 97:
			style.Fg = ansi.BasicColor(code - 90 + 8)
		case code >= 100 && code <= 107:
			style.Bg = ansi.BasicColor(code - 100 + 8)
		case code == 39:
			style.Fg = nil
		case code == 49:
			style.Bg = nil
		case code == 38, code == 48, code == 58:
			var c color.Color
			n := ansi.ReadStyleColor(params[i:], &c)
			if n == 0 {
				return style
			}
			switch code {
			case 38:
				style.Fg = c
			case 48:
				style.Bg = c
			}
			i += n - 1
		}
	}
	return style
}

// sameStyle reports whether a and b are drawn alike, text aside.
func sameStyle(a, b Cell) bool {
	return a.Attrs == b.Attrs && sameColor(a.Fg, b.Fg) && sameColor(a.Bg, b.Bg)
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

// open is the SGR that draws in c's style, from a reset.
func open(c Cell) string {
	s := ansi.Style{}
	if c.Fg != nil {
		s = s.ForegroundColor(c.Fg)
	}
	if c.Bg != nil {
		s = s.BackgroundColor(c.Bg)
	}
	for _, a := range []struct {
		bit Attr
		set func(ansi.Style) ansi.Style
	}{
		{Bold, ansi.Style.Bold}, {Faint, ansi.Style.Faint},
		{Italic, func(s ansi.Style) ansi.Style { return s.Italic(true) }},
		{Underline, func(s ansi.Style) ansi.Style { return s.Underline(true) }},
		{Blink, func(s ansi.Style) ansi.Style { return s.Blink(true) }},
		{Reverse, func(s ansi.Style) ansi.Style { return s.Reverse(true) }},
		{Strike, func(s ansi.Style) ansi.Style { return s.Strikethrough(true) }},
	} {
		if c.Attrs&a.bit != 0 {
			s = a.set(s)
		}
	}
	if len(s) == 0 {
		return ""
	}
	return s.String()
}

// String draws g, changing style only where it changes and resetting at the
// end of every styled line.
func (g Grid) String() string {
	var b strings.Builder
	for y, row := range g {
		if y > 0 {
			b.WriteByte('\n')
		}
		var cur Cell
		for _, c := range row {
			if !sameStyle(c, cur) {
				if cur.Fg != nil || cur.Bg != nil || cur.Attrs != 0 {
					b.WriteString(ansi.ResetStyle)
				}
				b.WriteString(open(c))
				cur = c
			}
			b.WriteString(c.Text)
		}
		if cur.Fg != nil || cur.Bg != nil || cur.Attrs != 0 {
			b.WriteString(ansi.ResetStyle)
		}
	}
	return b.String()
}

// Width is the widest row, in columns.
func (g Grid) Width() int {
	w := 0
	for _, row := range g {
		w = max(w, RowWidth(row))
	}
	return w
}

// RowWidth is how many columns row's cells take.
func RowWidth(row []Cell) int {
	n := 0
	for _, c := range row {
		n += c.Width
	}
	return n
}
