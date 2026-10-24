package effect

import (
	"image/color"

	"github.com/lucasassuncao/bezel/internal/cells"
)

// Ghost is block faded towards bg, alpha of it left in [0,1]: the copy that
// follows the cursor while something is dragged, the old place of a moved item.
func Ghost(block string, bg color.Color, alpha float64) string {
	g := cells.Parse(block)
	alpha = min(max(alpha, 0), 1)
	for _, row := range g {
		for i, c := range row {
			row[i] = tinted(c, bg, alpha)
		}
	}
	return g.String()
}
