package draw

import (
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

// ViewportGutter is a viewport line-number column with the same shape as
// Gutter, so a viewport and a NumberLines block line up side by side. The
// viewport probes the width with a zero context, hence the fixed format.
func ViewportGutter(style lipgloss.Style) viewport.GutterFunc {
	return func(ctx viewport.GutterContext) string {
		if ctx.Index >= ctx.TotalLines {
			return style.Render("     │ ")
		}
		return Gutter(ctx.Index+1, style)
	}
}
