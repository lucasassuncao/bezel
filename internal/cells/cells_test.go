package cells

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlainTextComesBackAsItWent(t *testing.T) {
	g := Parse("ab\ncd")
	require.Len(t, g, 2)
	assert.Equal(t, "ab\ncd", g.String())
	assert.Equal(t, 2, g.Width())
}

func TestStylesSurviveTheRoundTrip(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true)
	g := Parse(red.Render("hi") + "x")
	c := g[0][0]
	r, gr, b, _ := c.Fg.RGBA()
	assert.Equal(t, [3]uint32{0xffff, 0, 0}, [3]uint32{r, gr, b})
	assert.True(t, c.Attrs&Bold != 0)
	assert.Nil(t, g[0][2].Fg, "the reset after hi left x plain")
	assert.Equal(t, "hix", ansi.Strip(g.String()))
	assert.Equal(t, g, Parse(g.String()), "drawing what was read reads the same")
}

func TestEveryColourFormReads(t *testing.T) {
	g := Parse("\x1b[31ma\x1b[38;5;200mb\x1b[48;2;1;2;3mc\x1b[93;44md\x1b[39;49me")
	row := g[0]
	assert.Equal(t, ansi.BasicColor(1), row[0].Fg)
	assert.Equal(t, ansi.IndexedColor(200), row[1].Fg)
	r, gr, b, _ := row[2].Bg.RGBA()
	assert.Equal(t, [3]uint32{1 * 257, 2 * 257, 3 * 257}, [3]uint32{r, gr, b})
	assert.Equal(t, ansi.BasicColor(11), row[3].Fg, "bright yellow")
	assert.Equal(t, ansi.BasicColor(4), row[3].Bg)
	assert.Nil(t, row[4].Fg)
	assert.Nil(t, row[4].Bg)
}

func TestWideGraphemesTakeTwoColumns(t *testing.T) {
	g := Parse("日x")
	require.Len(t, g[0], 2)
	assert.Equal(t, 2, g[0][0].Width)
	assert.Equal(t, 3, g.Width())
}

// Sequences that draw nothing, a hyperlink say, are dropped, not drawn as text.
func TestOtherSequencesAreDropped(t *testing.T) {
	g := Parse("\x1b]8;;https://example.com\x07link\x1b]8;;\x07")
	assert.Equal(t, "link", g.String())
}
