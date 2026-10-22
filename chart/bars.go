package chart

import (
	"math"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// Bar is one row of Bars: what it measures, how much, and the text after it,
// Value printed when empty. Style paints this bar alone, as a threshold
// would; nil keeps the chart's colour.
type Bar struct {
	Label string
	Value float64
	Text  string
	Style *lipgloss.Style
}

// eighths are a cell filled from the left, one eighth at a time.
var eighths = []rune("▏▎▍▌▋▊▉")

// Bars is one row per bar, width cells wide: the labels in a column, each bar
// scaled so top fills it, its edge cut to an eighth, and its text right-aligned
// at the end. A long label is cut before the bars give up any room.
func Bars(bars []Bar, top float64, width int, th theme.Resolved, opts ...Option) string {
	if width <= 0 || len(bars) == 0 {
		return ""
	}
	st := th.Accent
	if l := options(opts); l.series >= 0 {
		st = th.Series[l.series%len(th.Series)]
	}
	labelW, textW := 0, 0
	for _, b := range bars {
		labelW = max(labelW, draw.Width(b.Label))
		textW = max(textW, draw.Width(b.text()))
	}
	labelW = min(labelW, width/3)
	barW := width - labelW - textW - 2 // a space after the label and before the text

	rows := make([]string, len(bars))
	for i, b := range bars {
		text := strings.Repeat(" ", textW-draw.Width(b.text())) + b.text()
		if barW < 1 {
			rows[i] = draw.Fit(draw.Fit(b.Label, labelW)+" "+text, width)
			continue
		}
		fill := st
		if b.Style != nil {
			fill = *b.Style
		}
		rows[i] = draw.Fit(b.Label, labelW) + " " + bar(b.Value, top, barW, fill) + " " + text
	}
	return strings.Join(rows, "\n")
}

// text is what the row prints after the bar.
func (b Bar) text() string {
	if b.Text != "" {
		return b.Text
	}
	return strings.TrimSuffix(strconv.FormatFloat(b.Value, 'f', 1, 64), ".0")
}

// bar is v of top as width cells, the filled ones in st and the rest blank.
func bar(v, top float64, width int, st lipgloss.Style) string {
	n := 0
	if frac := v / top; top > 0 && !math.IsNaN(frac) { // a NaN would turn into a count below zero
		n = int(min(max(frac, 0), 1)*float64(width*8) + 0.5)
	}
	cells := strings.Repeat("█", n/8)
	if n%8 > 0 {
		cells += string(eighths[n%8-1])
	}
	out := ""
	if cells != "" {
		out = st.Render(cells)
	}
	return out + strings.Repeat(" ", width-draw.Width(cells))
}
