package animation

import (
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// attr is a cell's text attributes, one bit each.
type attr uint8

const (
	bold attr = 1 << iota
	faint
	italic
	underline
	blink
	reverse
	strike
)

// cell is one grapheme of a drawn block and how it was drawn.
type cell struct {
	text   string
	width  int
	fg, bg color.Color // nil is the terminal's own
	attrs  attr
}

// grid is a drawn block read back into cells, a row per line.
type grid [][]cell

// parse reads block, as lipgloss draws it, into cells. Colours and attributes
// set by SGR follow the text; every other escape sequence draws nothing and
// is dropped.
func parse(block string) grid {
	p := ansi.NewParser()
	var g grid
	row := []cell{}
	var cur cell
	state := byte(0)
	for len(block) > 0 {
		seq, width, n, next := ansi.DecodeSequence(block, state, p)
		state, block = next, block[n:]
		switch {
		case seq == "\n":
			g, row = append(g, row), []cell{}
		case width > 0:
			c := cur
			c.text, c.width = seq, width
			row = append(row, c)
		case strings.HasPrefix(seq, "\x1b[") && ansi.Cmd(p.Command()).Final() == 'm':
			cur = sgr(cur, p.Params())
		}
	}
	return append(g, row)
}

// sgr is style changed by one Select Graphic Rendition sequence's params.
func sgr(style cell, params ansi.Params) cell {
	if len(params) == 0 {
		return cell{}
	}
	for i := 0; i < len(params); i++ {
		switch code := params[i].Param(0); {
		case code == 0:
			style = cell{}
		case code >= 1 && code <= 9:
			style.attrs |= map[int]attr{1: bold, 2: faint, 3: italic, 4: underline, 5: blink, 7: reverse, 9: strike}[code]
		case code == 22:
			style.attrs &^= bold | faint
		case code >= 23 && code <= 29:
			style.attrs &^= map[int]attr{23: italic, 24: underline, 25: blink, 27: reverse, 29: strike}[code]
		case code >= 30 && code <= 37:
			style.fg = ansi.BasicColor(code - 30)
		case code >= 40 && code <= 47:
			style.bg = ansi.BasicColor(code - 40)
		case code >= 90 && code <= 97:
			style.fg = ansi.BasicColor(code - 90 + 8)
		case code >= 100 && code <= 107:
			style.bg = ansi.BasicColor(code - 100 + 8)
		case code == 39:
			style.fg = nil
		case code == 49:
			style.bg = nil
		case code == 38, code == 48, code == 58:
			var c color.Color
			n := ansi.ReadStyleColor(params[i:], &c)
			if n == 0 {
				return style
			}
			switch code {
			case 38:
				style.fg = c
			case 48:
				style.bg = c
			}
			i += n - 1
		}
	}
	return style
}

// sameStyle reports whether a and b are drawn alike, text aside.
func sameStyle(a, b cell) bool {
	return a.attrs == b.attrs && sameColor(a.fg, b.fg) && sameColor(a.bg, b.bg)
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
func open(c cell) string {
	s := ansi.Style{}
	if c.fg != nil {
		s = s.ForegroundColor(c.fg)
	}
	if c.bg != nil {
		s = s.BackgroundColor(c.bg)
	}
	for _, a := range []struct {
		bit attr
		set func(ansi.Style) ansi.Style
	}{
		{bold, ansi.Style.Bold}, {faint, ansi.Style.Faint},
		{italic, func(s ansi.Style) ansi.Style { return s.Italic(true) }},
		{underline, func(s ansi.Style) ansi.Style { return s.Underline(true) }},
		{blink, func(s ansi.Style) ansi.Style { return s.Blink(true) }},
		{reverse, func(s ansi.Style) ansi.Style { return s.Reverse(true) }},
		{strike, func(s ansi.Style) ansi.Style { return s.Strikethrough(true) }},
	} {
		if c.attrs&a.bit != 0 {
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
func (g grid) String() string {
	var b strings.Builder
	for y, row := range g {
		if y > 0 {
			b.WriteByte('\n')
		}
		var cur cell
		for _, c := range row {
			if !sameStyle(c, cur) {
				if cur.fg != nil || cur.bg != nil || cur.attrs != 0 {
					b.WriteString(ansi.ResetStyle)
				}
				b.WriteString(open(c))
				cur = c
			}
			b.WriteString(c.text)
		}
		if cur.fg != nil || cur.bg != nil || cur.attrs != 0 {
			b.WriteString(ansi.ResetStyle)
		}
	}
	return b.String()
}

// width is the widest row, in columns.
func (g grid) width() int {
	w := 0
	for _, row := range g {
		n := 0
		for _, c := range row {
			n += c.width
		}
		w = max(w, n)
	}
	return w
}
