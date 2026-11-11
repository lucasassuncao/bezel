package chart

import (
	"math"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/braille"
	"github.com/lucasassuncao/bezel/theme"
)

// Line is one series of Lines: its values, oldest first, and its name for
// the key under the chart; a NaN leaves a gap.
type Line struct {
	Label  string
	Values []float64
}

// Lines draws each line in braille, width by height cells: the y axis on the
// left from lo at the bottom to hi at the top, the newest values on the right
// and a key of the Labels on the last row. lo equal to hi draws at the middle.
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

	g := braille.New(plotW, height)
	for i, l := range lines {
		plot(g, l.Values, lo, hi, i)
	}
	rows := make([]string, height)
	for r := range height {
		lead, tick := "", "│"
		if labelW > 0 {
			lead = strings.Repeat(" ", labelW)
			if ticks[r] != "" {
				lead, tick = th.Dim.Render(draw.Pad(strings.Repeat(" ", labelW-draw.Width(ticks[r]))+ticks[r], labelW)), "┤"
			}
		}
		rows[r] = lead + th.Muted.Render(tick) + g.Row(r, styles)
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
// sits halfway, the middle; lo alone at the middle when it equals hi, and
// all "" for a range that is not one.
func axisTicks(lo, hi float64, height int) []string {
	ticks := make([]string, height)
	switch {
	case !valid(lo, hi):
		return ticks
	case hi == lo:
		ticks[(height-1)/2] = number(lo, 1)
		return ticks
	}
	d := decimals(hi - lo)
	ticks[0] = number(hi, d)
	if height > 1 {
		ticks[height-1] = number(lo, d)
	}
	if height >= 5 && (height-1)%2 == 0 {
		ticks[(height-1)/2] = number((lo+hi)/2, d+1)
	}
	return ticks
}

// valid reports whether lo to hi is a range to plot: finite, hi not below lo.
func valid(lo, hi float64) bool {
	return hi >= lo && !math.IsNaN(lo) && !math.IsNaN(hi) && !math.IsInf(hi-lo, 0)
}

// decimals is how many a range of span needs for its ends to read apart:
// one, and more below 1.
func decimals(span float64) int {
	if span >= 1 {
		return 1
	}
	return min(int(math.Ceil(-math.Log10(span))), 8)
}

// number is v with d decimals at most, trailing zeros dropped.
func number(v float64, d int) string {
	s := strconv.FormatFloat(v, 'f', d, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if s == "-0" {
		s = "0"
	}
	return s
}

// plot draws the last two dots a cell of values, joined by straight runs,
// right-aligned; a NaN breaks the line, lo equal to hi draws at the middle.
func plot(g braille.Grid, values []float64, lo, hi float64, line int) {
	dw, dh := g.Dots()
	if !valid(lo, hi) {
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
		frac := 0.5
		if hi > lo {
			frac = min(max((v-lo)/(hi-lo), 0), 1)
		}
		x, y := x0+i, dh-1-int(math.Round(frac*float64(dh-1)))
		if have {
			join(g, prevX, prevY, x, y, line)
		} else {
			g.Set(x, y, line)
		}
		prevX, prevY, have = x, y, true
	}
}

// join lights the dots from one point to the next: the column steps halfway
// at the middle of the rise, so a jump reads as a line, not two dots.
func join(g braille.Grid, x0, y0, x1, y1, line int) {
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
		g.Set(x, y, line)
		if y == y1 {
			break
		}
	}
	g.Set(x1, y1, line)
}
