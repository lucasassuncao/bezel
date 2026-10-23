package animation

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlainTextComesBackAsItWent(t *testing.T) {
	g := parse("ab\ncd")
	require.Len(t, g, 2)
	assert.Equal(t, "ab\ncd", g.String())
	assert.Equal(t, 2, g.width())
}

func TestStylesSurviveTheRoundTrip(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true)
	g := parse(red.Render("hi") + "x")
	c := g[0][0]
	r, gr, b, _ := c.fg.RGBA()
	assert.Equal(t, [3]uint32{0xffff, 0, 0}, [3]uint32{r, gr, b})
	assert.True(t, c.attrs&bold != 0)
	assert.Nil(t, g[0][2].fg, "the reset after hi left x plain")
	assert.Equal(t, "hix", ansi.Strip(g.String()))
	assert.Equal(t, g, parse(g.String()), "drawing what was read reads the same")
}

func TestEveryColourFormReads(t *testing.T) {
	g := parse("\x1b[31ma\x1b[38;5;200mb\x1b[48;2;1;2;3mc\x1b[93;44md\x1b[39;49me")
	row := g[0]
	assert.Equal(t, ansi.BasicColor(1), row[0].fg)
	assert.Equal(t, ansi.IndexedColor(200), row[1].fg)
	r, gr, b, _ := row[2].bg.RGBA()
	assert.Equal(t, [3]uint32{1 * 257, 2 * 257, 3 * 257}, [3]uint32{r, gr, b})
	assert.Equal(t, ansi.BasicColor(11), row[3].fg, "bright yellow")
	assert.Equal(t, ansi.BasicColor(4), row[3].bg)
	assert.Nil(t, row[4].fg)
	assert.Nil(t, row[4].bg)
}

func TestWideGraphemesTakeTwoColumns(t *testing.T) {
	g := parse("日x")
	require.Len(t, g[0], 2)
	assert.Equal(t, 2, g[0][0].width)
	assert.Equal(t, 3, g.width())
}

// Sequences that draw nothing, a hyperlink say, are dropped, not drawn as text.
func TestOtherSequencesAreDropped(t *testing.T) {
	g := parse("\x1b]8;;https://example.com\x07link\x1b]8;;\x07")
	assert.Equal(t, "link", g.String())
}
