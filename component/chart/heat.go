package chart

import (
	"math"
	"strings"

	"github.com/lucasassuncao/bezel/theme"
)

// Heat is values as a grid of cells two columns wide, each coloured along the
// theme's ramp by its share of top; a value at or below zero is a muted dot.
func Heat(values [][]float64, top float64, th theme.Resolved) string {
	rows := make([]string, len(values))
	for y, row := range values {
		var b strings.Builder
		for _, v := range row {
			share := 0.0
			if top > 0 {
				share = min(max(v/top, 0), 1)
			}
			if share <= 0 {
				b.WriteString(th.Muted.Render("··"))
				continue
			}
			i := int(math.Round(share * float64(len(th.Ramp)-1)))
			b.WriteString(th.Ramp[i].Render("██"))
		}
		rows[y] = b.String()
	}
	return strings.Join(rows, "\n")
}
