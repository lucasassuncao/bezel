package overlay

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
)

// MaxInner stops a modal spanning a wide terminal: past this a line stops
// being scannable anyway.
const MaxInner = 120

// Inner is the content width a modal gets inside body.
func Inner(body layout.Rect) int { return min(max(10, body.W-8), MaxInner) }

// Box frames lines in style and clips them to body, so an overflowing modal
// never pushes its own border off screen. Lines cut are counted on the last row.
func Box(lines []string, body layout.Rect, style lipgloss.Style) string {
	return BoxWidth(lines, body, Inner(body), style)
}

// BoxWidth is Box with the content width chosen by the caller, for a form
// that reads better narrower than the terminal allows.
func BoxWidth(lines []string, body layout.Rect, inner int, style lipgloss.Style) string {
	clipped := make([]string, 0, len(lines))
	for _, line := range lines {
		for _, part := range strings.Split(line, "\n") {
			clipped = append(clipped, draw.Truncate(part, inner))
		}
	}
	// The frame is border and padding both; the cut spares the last line, which
	// is where a dialog keeps its buttons and so how to leave it.
	if maxLines := body.H - style.GetVerticalFrameSize(); maxLines > 2 && len(clipped) > maxLines {
		cut := len(clipped) - maxLines + 1
		last := clipped[len(clipped)-1]
		clipped = append(clipped[:maxLines-2], fmt.Sprintf("… %d more lines, resize to see them", cut), last)
	}
	return style.MaxWidth(body.W).Render(strings.Join(clipped, "\n"))
}
