package table

import (
	"strings"

	"github.com/lucasassuncao/bezel/theme"
)

// Static is the whole table as text to print, the header on top: no cursor,
// no scrolling. Columns fit width as in Fit; width <= 0 keeps their own.
func Static(cols []Column, rows [][]string, width int, th theme.Resolved) string {
	widths := natural(cols, rows)
	if width > 0 {
		widths = shrink(cols, widths, width)
	}
	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, th.TableHeader.Render(Titles(cols, widths)))
	for _, r := range rows {
		lines = append(lines, Row(widths, r))
	}
	return strings.Join(lines, "\n")
}
