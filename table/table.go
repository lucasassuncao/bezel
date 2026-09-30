// Package table sizes columns to their contents and lays cells out in them:
// a row is as narrow as its widest values allow, and when even that is too
// wide, the columns give way in a declared order rather than all at once.
package table

import (
	"strings"

	"github.com/lucasassuncao/bezel/draw"
)

// Column declares one column; Min and Max bound its width (0 = unbounded).
// Flex marks the column that shrinks first: a truncated name is still
// recognisable, a truncated version is not.
type Column struct {
	Title    string
	Min, Max int
	Flex     bool
}

// Gap is the columns between cells.
const Gap = 2

// Fit returns each column's width: its widest cell (title included) capped at
// Max, then shrunk to fit width - Flex first down to Min, then the widest,
// never below minCell. Min is a floor for shrinking, not a width to grow to.
func Fit(cols []Column, rows [][]string, width int) []int {
	const minCell = 8
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = draw.Width(c.Title)
	}
	for _, r := range rows {
		for i := range cols {
			if i < len(r) {
				widths[i] = max(widths[i], draw.Width(r[i]))
			}
		}
	}
	for i, c := range cols {
		if c.Max > 0 {
			widths[i] = min(widths[i], c.Max)
		}
	}

	avail := width - Gap*(len(cols)-1)
	for total(widths) > avail {
		if i := flexOf(cols); i >= 0 && widths[i] > max(cols[i].Min, 1) {
			widths[i]--
			continue
		}
		widest := -1
		for i := range widths {
			if widest < 0 || widths[i] > widths[widest] {
				widest = i
			}
		}
		if widest < 0 || widths[widest] <= minCell {
			break // nothing left to give; the panel clips the tail
		}
		widths[widest]--
	}
	return widths
}

func flexOf(cols []Column) int {
	for i, c := range cols {
		if c.Flex {
			return i
		}
	}
	return -1
}

func total(ws []int) int {
	n := 0
	for _, w := range ws {
		n += w
	}
	return n
}

// Row lays cells out in widths, padded to their columns and separated by
// Gap; the last cell is cut but not padded, so the row ends where it ends.
func Row(widths []int, cells []string) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		if i == len(widths)-1 {
			parts[i] = draw.Truncate(cell, w)
		} else {
			parts[i] = draw.Fit(cell, w)
		}
	}
	return strings.Join(parts, strings.Repeat(" ", Gap))
}

// Titles is the header row: every column's title in its width.
func Titles(cols []Column, widths []int) string {
	titles := make([]string, len(cols))
	for i, c := range cols {
		titles[i] = c.Title
	}
	return Row(widths, titles)
}
