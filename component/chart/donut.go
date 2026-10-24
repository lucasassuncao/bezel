package chart

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/theme"
)

// Donut is a ring size rows tall and twice as wide, filled clockwise from the
// top by value's share of total, with the share in its middle: a quota, a
// disk, a goal. Animate value with a Float to sweep it in.
func Donut(value, total float64, size int, th theme.Resolved) string {
	size = max(size, 3)
	w, h := size*2, size
	share := 0.0
	if total > 0 {
		share = min(max(value/total, 0), 1)
	}
	r := float64(h) / 2
	thick := max(r*0.4, 1)
	rows := make([]string, h)
	label := fmt.Sprintf("%d%%", int(math.Round(share*100)))
	for y := range h {
		var b strings.Builder
		for x := 0; x < w; x++ {
			if y == h/2 && x == (w-ansi.StringWidth(label))/2 && ansi.StringWidth(label) <= w-2 {
				b.WriteString(th.Bold.Render(label))
				x += ansi.StringWidth(label) - 1
				continue
			}
			dx := (float64(x) + 0.5 - float64(w)/2) / 2
			dy := float64(y) + 0.5 - float64(h)/2
			d := math.Hypot(dx, dy)
			if d > r || d < r-thick {
				b.WriteByte(' ')
				continue
			}
			angle := math.Atan2(dx, -dy)
			if angle < 0 {
				angle += 2 * math.Pi
			}
			if angle/(2*math.Pi) < share {
				b.WriteString(th.Accent.Render("█"))
			} else {
				b.WriteString(th.Muted.Render("█"))
			}
		}
		rows[y] = b.String()
	}
	return strings.Join(rows, "\n")
}
