package progress

import (
	"math"
	"strings"

	"github.com/lucasassuncao/bezel/theme"
)

// Gauge is value out of total as a needle under a ruled track width cells
// wide, two rows: the track with a tick every quarter, then the needle. The
// value is a float, so a Spring can carry the needle smoothly.
func Gauge(value, total float64, width int, th theme.Resolved) string {
	width = max(width, 2)
	track := []rune(strings.Repeat("─", width))
	for q := range 5 {
		at := int(math.Round(float64(q) * float64(width-1) / 4))
		switch q {
		case 0:
			track[at] = '├'
		case 4:
			track[at] = '┤'
		default:
			track[at] = '┼'
		}
	}
	at := 0
	if total > 0 {
		at = int(math.Round(min(max(value/total, 0), 1) * float64(width-1)))
	}
	needle := strings.Repeat(" ", at) + th.Accent.Render("▲") + strings.Repeat(" ", width-1-at)
	return th.Muted.Render(string(track)) + "\n" + needle
}
