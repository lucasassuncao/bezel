// Package chart draws numbers in cells: Area and Spark a run of them over
// time, the newest on the right, and Bars a few side by side, one row each.
// Every cell is cut to an eighth, so a bar or a column grows smoothly.
package chart

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/theme"
)

// Option changes how a chart is drawn.
type Option func(*look)

type look struct{ series int }

// InSeries draws the chart whole in the theme's i-th series colour, instead
// of along the ramp (Area) or in the accent (Spark).
func InSeries(i int) Option { return func(l *look) { l.series = i } }

// Fill is the style a bar or a line in opts is filled with: the theme's i-th
// series colour under InSeries(i), the accent otherwise. progress.Smooth reads
// it too, so one InSeries serves both packages.
func Fill(th theme.Resolved, opts ...Option) lipgloss.Style {
	if l := options(opts); l.series >= 0 && len(th.Series) > 0 {
		return th.Series[l.series%len(th.Series)]
	}
	return th.Accent
}

func options(opts []Option) look {
	l := look{series: -1}
	for _, o := range opts {
		if o != nil {
			o(&l)
		}
	}
	return l
}

// blocks are a cell filled from the bottom, one eighth at a time.
var blocks = []rune("▁▂▃▄▅▆▇█")

// Area is the last width values as columns height rows tall, scaled so top
// fills them, each row coloured along the theme's ramp from the bottom up.
// Fewer values than columns leave the left blank.
func Area(values []float64, top float64, width, height int, th theme.Resolved, opts ...Option) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	l := options(opts)
	heights := columns(values, top, width, height*8)
	rows := make([]string, height)
	for r := range height { // r counts from the bottom
		st := th.Ramp[0]
		if height > 1 {
			st = th.Ramp[int(math.Round(float64(r)*float64(len(th.Ramp)-1)/float64(height-1)))]
		}
		if l.series >= 0 {
			st = th.Series[l.series%len(th.Series)]
		}
		cells := make([]rune, width)
		for c, h := range heights {
			cells[c] = cell(h - r*8)
		}
		rows[height-1-r] = paint(cells, st)
	}
	return strings.Join(rows, "\n")
}

// Spark is the last width values as one row, scaled to top, in the accent.
func Spark(values []float64, top float64, width int, th theme.Resolved, opts ...Option) string {
	if width <= 0 {
		return ""
	}
	st := Fill(th, opts...)
	cells := make([]rune, width)
	for c, h := range columns(values, top, width, 8) {
		cells[c] = cell(h)
	}
	return paint(cells, st)
}

// columns are the last width values in steps of top, right-aligned; none
// at all when top is not positive.
func columns(values []float64, top float64, width, steps int) []int {
	out := make([]int, width)
	if top <= 0 {
		return out
	}
	if len(values) > width {
		values = values[len(values)-width:]
	}
	for i, v := range values {
		out[width-len(values)+i] = int(math.Round(min(max(v/top, 0), 1) * float64(steps)))
	}
	return out
}

// cell is a cell n eighths full from the bottom.
func cell(n int) rune {
	switch {
	case n <= 0:
		return ' '
	case n >= 8:
		return '█'
	}
	return blocks[n-1]
}

// paint is cells with every run of marks in st and the gaps left bare.
func paint(cells []rune, st lipgloss.Style) string {
	var b strings.Builder
	for i := 0; i < len(cells); {
		j := i
		for j < len(cells) && (cells[j] == ' ') == (cells[i] == ' ') {
			j++
		}
		if cells[i] == ' ' {
			b.WriteString(string(cells[i:j]))
		} else {
			b.WriteString(st.Render(string(cells[i:j])))
		}
		i = j
	}
	return b.String()
}
