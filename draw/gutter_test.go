package draw

import (
	"testing"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// A viewport's gutter lines up with Gutter, and past the last line keeps the
// same width with no number, so the rule runs straight down.
func TestViewportGutterMatchesGutterAndBlanksPastTheEnd(t *testing.T) {
	st := lipgloss.NewStyle()
	g := ViewportGutter(st)
	require.Equal(t, Gutter(3, st), g(viewport.GutterContext{Index: 2, TotalLines: 5}))
	past := ansi.Strip(g(viewport.GutterContext{Index: 5, TotalLines: 5}))
	require.Equal(t, "     │ ", past)
	require.Equal(t, GutterWidth, ansi.StringWidth(past))
}
