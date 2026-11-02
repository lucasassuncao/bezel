package chart

import (
	"math"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// Line is one series of Lines: its values, oldest first, and its name for
// the key under the chart; a NaN leaves a gap.
type Line struct {
	Label  string
	Values []float64
}

// dotBits are the braille bits of a cell's dots, by column then by row from
// the top: a cell is two dots wide and four tall.
var dotBits = [2][4]rune{{0x01, 0x02, 0x04, 0x40}, {0x08, 0x10, 0x20, 0x80}}

// Lines draws each line in braille, width by height cells: the y axis on the
// left from lo at the bottom to hi at the top, the newest values on the right
// and, when a line has a Label, a key on the last row.
func Lines(lines []Line, lo, hi float64, width, height int, th theme.Resolved, opts ...Option) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	styles := lineStyles(len(lines), th, options(opts))
	var key string
	if named(lines) && height > 1 {
		key = draw.Truncate(lineKey(lines, styles), width)
		height--
	}
	ticks := axisTicks(lo, hi, height)
	labelW := 0
	for _, t := range ticks {
		labelW = max(labelW, draw.Width(t))
	}
	plotW := width - labelW - 1 // the axis takes a cell
	if plotW < 1 {
		labelW, plotW = 0, width-1
	}
	if plotW < 1 {
		return strings.TrimSuffix(strings.Repeat("\n", height)+key, "\n")
	}

	g := newGrid(plotW, height)
	for i, l := range lines {
		g.plot(l.Values, lo, hi, i)
	}
	axis := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Colors.Border))
	rows := make([]string, height)
	for r := range height {
		lead, tick := "", "│"
		if labelW > 0 {
			lead = strings.Repeat(" ", labelW)
			if ticks[r] != "" {
				lead, tick = th.Dim.Render(draw.Pad(strings.Repeat(" ", labelW-draw.Width(ticks[r]))+ticks[r], labelW)), "┤"
			}
		}
		rows[r] = lead + axis.Render(tick) + g.row(r, styles)
	}
	if key != "" {
		rows = append(rows, key)
	}
	return strings.Join(rows, "\n")
}

// lineStyles is each line's colour: the accent for a lone line, else the
// series colours in turn, from InSeries' when given.
func lineStyles(n int, th theme.Resolved, l look) []lipgloss.Style {
	out := make([]lipgloss.Style, n)
	for i := range out {
		switch {
		case n == 1 && l.series < 0:
			out[i] = th.Accent
		default:
			out[i] = th.Series[(max(l.series, 0)+i)%len(th.Series)]
		}
	}
	return out
}

func named(lines []Line) bool {
	for _, l := range lines {
		if l.Label != "" {
			return true
		}
	}
	return false
}

// lineKey is "● name" for every named line, in its colour.
func lineKey(lines []Line, styles []lipgloss.Style) string {
	parts := make([]string, 0, len(lines))
	for i, l := range lines {
		if l.Label != "" {
			parts = append(parts, styles[i].Render("●")+" "+l.Label)
		}
	}
	return strings.Join(parts, "  ")
}

// axisTicks is the label of each row from the top: hi, lo and, when a row
// sits halfway, the middle; "" for the rest, and all "" unless hi > lo.
func axisTicks(lo, hi float64, height int) []string {
	ticks := make([]string, height)
	if hi <= lo || math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(hi-lo, 0) {
		return ticks
	}
	ticks[0] = number(hi)
	if height > 1 {
		ticks[height-1] = number(lo)
	}
	if height >= 5 && (height-1)%2 == 0 {
		ticks[(height-1)/2] = number((lo + hi) / 2)
	}
	return ticks
}

// number is v with one decimal at most, as Bars prints a value.
func number(v float64) string {
	return strings.TrimSuffix(strconv.FormatFloat(v, 'f', 1, 64), ".0")
}

// grid is the braille dots of the plot, and in each cell the last line that
// put a dot there, whose colour the cell takes.
type grid struct {
	w, h  int
	bits  []rune
	owner []int
}

func newGrid(w, h int) grid {
	return grid{w: w, h: h, bits: make([]rune, w*h), owner: make([]int, w*h)}
}

// plot draws the last two dots a cell of values, joined by straight runs,
// right-aligned; a NaN breaks the line and hi not above lo draws nothing.
func (g grid) plot(values []float64, lo, hi float64, line int) {
	dw, dh := g.w*2, g.h*4
	if hi <= lo || math.IsInf(hi-lo, 0) {
		return
	}
	if len(values) > dw {
		values = values[len(values)-dw:]
	}
	x0 := dw - len(values)
	prevX, prevY, have := 0, 0, false
	for i, v := range values {
		if math.IsNaN(v) {
			have = false
			continue
		}
		frac := min(max((v-lo)/(hi-lo), 0), 1)
		x, y := x0+i, dh-1-int(math.Round(frac*float64(dh-1)))
		if have {
			g.join(prevX, prevY, x, y, line)
		} else {
			g.dot(x, y, line)
		}
		prevX, prevY, have = x, y, true
	}
}

// join lights the dots from one point to the next: the column steps halfway
// at the middle of the rise, so a jump reads as a line, not two dots.
func (g grid) join(x0, y0, x1, y1, line int) {
	mid := (y0 + y1) / 2
	step := 1
	if y1 < y0 {
		step = -1
	}
	for y := y0; ; y += step {
		x := x0
		if (step > 0 && y > mid) || (step < 0 && y < mid) {
			x = x1
		}
		g.dot(x, y, line)
		if y == y1 {
			break
		}
	}
	g.dot(x1, y1, line)
}

func (g grid) dot(x, y, line int) {
	if x < 0 || y < 0 || x >= g.w*2 || y >= g.h*4 {
		return
	}
	c := (y/4)*g.w + x/2
	g.bits[c] |= dotBits[x%2][y%4]
	g.owner[c] = line
}

// row is cell row r, each run of one line's cells in its colour.
func (g grid) row(r int, styles []lipgloss.Style) string {
	var b strings.Builder
	cells := g.bits[r*g.w : (r+1)*g.w]
	owners := g.owner[r*g.w : (r+1)*g.w]
	for i := 0; i < len(cells); {
		j := i
		for j < len(cells) && (cells[j] == 0) == (cells[i] == 0) && (cells[i] == 0 || owners[j] == owners[i]) {
			j++
		}
		if cells[i] == 0 {
			b.WriteString(strings.Repeat(" ", j-i))
		} else {
			run := make([]rune, j-i)
			for k := range run {
				run[k] = 0x2800 + cells[i+k]
			}
			b.WriteString(styles[owners[i]].Render(string(run)))
		}
		i = j
	}
	return b.String()
}
