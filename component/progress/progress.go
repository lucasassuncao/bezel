// Package progress draws how far along something is: bars, a meter and a
// stepper of named stages, each one row exactly the width asked for, and
// Model, a bar that springs to the share it is set to.
package progress

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/component/chart"
	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// Bar is a done/total bar width cells wide: filled in the accent colour, the
// rest muted. A zero total reads as nothing done.
func Bar(done, total, width int, th theme.Resolved) string {
	width = max(width, 1)
	filled := 0
	if total > 0 {
		filled = min(max(done, 0)*width/total, width)
	}
	return th.Accent.Render(strings.Repeat("█", filled)) + th.Muted.Render(strings.Repeat("░", width-filled))
}

// minBar is the narrowest bar still worth drawing; below it Line drops the bar.
const minBar = 5

// Line is label, a bar and "done/total" in one row width cells wide. The bar
// takes what the label and the count leave; when space runs out the bar goes
// first, then the label is cut, and the count stays.
func Line(label string, done, total, width int, th theme.Resolved) string {
	count := th.Dim.Render(fmt.Sprintf("%d/%d", done, total))
	barW := width - ansi.StringWidth(label) - ansi.StringWidth(count) - 2*2
	if barW < minBar {
		label = draw.Truncate(label, max(0, width-ansi.StringWidth(count)-2))
		return draw.Fit(label+"  "+count, width)
	}
	return draw.Fit(label+"  "+Bar(done, total, barW, th)+"  "+count, width)
}

// Stepper is a row of named stages: those before current marked done, current
// marked running, the rest pending. current == len(steps) marks every stage
// done. When the row does not fit width it becomes "Step 2 of 4 · Name".
func Stepper(steps []string, current, width int, th theme.Resolved, set icon.Set) string {
	if len(steps) == 0 || width < 1 {
		return ""
	}
	current = min(max(current, 0), len(steps))
	parts := make([]string, len(steps))
	for i, name := range steps {
		switch {
		case i < current:
			parts[i] = set.Render(icon.OK, th) + " " + th.Dim.Render(name)
		case i == current:
			parts[i] = set.Render(icon.Running, th) + " " + th.Bold.Render(name)
		default:
			parts[i] = set.Render(icon.Pending, th) + " " + th.Dim.Render(name)
		}
	}
	row := strings.Join(parts, th.Muted.Render(" "+set.Line+" "))
	if ansi.StringWidth(row) <= width {
		return draw.Fit(row, width)
	}
	if current == len(steps) {
		return draw.Fit(set.Render(icon.OK, th)+" "+fmt.Sprintf("All %d steps done", len(steps)), width)
	}
	compact := fmt.Sprintf("Step %d of %d · ", current+1, len(steps))
	return draw.Fit(set.Render(icon.Running, th)+" "+th.Dim.Render(compact)+th.Bold.Render(steps[current]), width)
}

// Meter is done/total as width squares: the filled ones coloured along the
// theme's ramp from the left, the rest hollow, so it reads without colour.
func Meter(done, total, width int, th theme.Resolved) string {
	width = max(width, 1)
	filled := 0
	if total > 0 {
		filled = min(max(done, 0)*width/total, width)
	}
	var b strings.Builder
	for i := 0; i < filled; {
		step := i * len(th.Ramp) / width
		j := i
		for j < filled && j*len(th.Ramp)/width == step {
			j++
		}
		b.WriteString(th.Ramp[step].Render(strings.Repeat("■", j-i)))
		i = j
	}
	return b.String() + th.Muted.Render(strings.Repeat("□", width-filled))
}

// eighths are a cell filled from the left, one eighth at a time.
var eighths = []rune("▏▎▍▌▋▊▉")

// Smooth is done/total as a bar width cells wide whose edge is cut to an
// eighth of a cell, over a muted track as Bar's.
func Smooth(done, total, width int, th theme.Resolved, opts ...chart.Option) string {
	fill := chart.Fill(th, opts...)
	width = max(width, 1)
	n := 0
	if total > 0 {
		n = min(max(done, 0)*width*8/total, width*8)
	}
	cells := strings.Repeat("█", n/8)
	if n%8 > 0 {
		cells += string(eighths[n%8-1])
	}
	return fill.Render(cells) + th.Muted.Render(strings.Repeat("░", width-ansi.StringWidth(cells)))
}
