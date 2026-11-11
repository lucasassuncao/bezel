package canvas

import (
	"math"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

func th() theme.Resolved { return theme.Resolve(theme.Theme{}, true) }

// grid is a canvas of 4 by 2 cells over 0..7 both ways: one unit is one dot.
func grid() Model { return New(0, 7, 0, 7).SetWidth(4).SetHeight(2) }

func plain(m Model) []string { return strings.Split(ansi.Strip(m.View(th())), "\n") }

func TestAnUnsizedCanvasDrawsNothing(t *testing.T) {
	m := New(0, 1, 0, 1).Draw(Line{X1: 1, Y1: 1})
	require.Empty(t, m.View(th()))
	require.Empty(t, m.SetWidth(4).View(th()))
}

func TestAnEmptyCanvasIsBlankCellsOfItsSize(t *testing.T) {
	require.Equal(t, []string{"    ", "    "}, plain(grid()))
}

func TestYGrowsUpwards(t *testing.T) {
	m := grid().Draw(Points{Coords: []Point{{0, 0}, {7, 7}}})
	require.Equal(t, []string{"   ⠈", "⡀   "}, plain(m))
}

func TestDrawAddsAndClearRemoves(t *testing.T) {
	m := grid().Draw(Line{X0: 0, Y0: 7, X1: 7, Y1: 7})
	require.Len(t, m.Shapes(), 1)
	m = m.Draw(Line{X0: 0, Y0: 0, X1: 7, Y1: 0})
	require.Equal(t, []string{"⠉⠉⠉⠉", "⣀⣀⣀⣀"}, plain(m))
	require.Equal(t, []string{"    ", "    "}, plain(m.Clear()))
	require.Empty(t, m.Clear().Shapes())
}

// Drawing on a copy must not show on the canvas it came from.
func TestDrawLeavesTheOriginalAlone(t *testing.T) {
	base := grid().Draw(Line{X1: 7})
	a := base.Draw(Points{Coords: []Point{{7, 7}}})
	b := base.Draw(Points{Coords: []Point{{0, 7}}})
	require.Len(t, base.Shapes(), 1)
	require.Equal(t, []string{"   ⠈", "⣀⣀⣀⣀"}, plain(a))
	require.Equal(t, []string{"⠁   ", "⣀⣀⣀⣀"}, plain(b))
}

func TestSetBoundsMovesTheView(t *testing.T) {
	m := grid().Draw(Points{Coords: []Point{{10, 0}}})
	require.Equal(t, []string{"    ", "    "}, plain(m), "outside the region")
	m = m.SetBounds(10, 17, 0, 7)
	require.Equal(t, []string{"    ", "⡀   "}, plain(m))
	x0, x1, y0, y1 := m.Bounds()
	require.Equal(t, [4]float64{10, 17, 0, 7}, [4]float64{x0, x1, y0, y1})
}

func TestAnEmptyOrBrokenRegionDrawsBlank(t *testing.T) {
	for _, m := range []Model{New(1, 1, 0, 7), New(0, 7, 7, 0), New(math.NaN(), 7, 0, 7), New(0, math.Inf(1), 0, 7)} {
		m = m.SetWidth(2).SetHeight(1).Draw(Line{X1: 7, Y1: 7}, Points{Coords: []Point{{1, 1}}})
		require.Equal(t, "  ", ansi.Strip(m.View(th())))
	}
}

func TestSizesAndNegativeSizes(t *testing.T) {
	m := grid()
	require.Equal(t, 4, m.Width())
	require.Equal(t, 2, m.Height())
	m = m.SetWidth(-3).SetHeight(-1)
	require.Zero(t, m.Width())
	require.Zero(t, m.Height())
}

func TestSeriesColoursAShapeAndWithStylesEditsThem(t *testing.T) {
	m := grid().Draw(Points{Coords: []Point{{0, 7}}, Series: 1})
	want := th().Series[1].Render("⠁")
	require.True(t, strings.HasPrefix(m.View(th()), want))

	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m = m.WithStyles(func(s Styles) Styles {
		s.Series = []lipgloss.Style{red}
		return s
	})
	require.True(t, strings.HasPrefix(m.View(th()), red.Render("⠁")), "series 1 wraps round one style")
}

func TestALaterShapeTakesASharedCellsColour(t *testing.T) {
	m := grid().Draw(Points{Coords: []Point{{0, 7}}, Series: 0}, Points{Coords: []Point{{1, 7}}, Series: 2})
	require.True(t, strings.HasPrefix(m.View(th()), th().Series[2].Render("⠉")))
}

func TestANilShapeIsSkipped(t *testing.T) {
	require.NotPanics(t, func() { grid().Draw(nil, Line{}).View(th()) })
}

func TestClipKeepsTheShareInsideTheBox(t *testing.T) {
	t0, t1, ok := clip(-10, 0, 20, 0, 0, -1, 10, 1)
	require.True(t, ok)
	require.InDelta(t, 0.5, t0, 1e-9)
	require.InDelta(t, 1.0, t1, 1e-9)
	_, _, ok = clip(-10, 5, 20, 0, 0, -1, 10, 1)
	require.False(t, ok, "a line above the box")
	_, _, ok = clip(-10, 5, 0, 0, 0, -1, 10, 1)
	require.False(t, ok, "a point outside")
}

func BenchmarkView(b *testing.B) {
	m := New(0, 100, 0, 100).SetWidth(80).SetHeight(24)
	for i := range 50 {
		f := float64(i * 2)
		m = m.Draw(Line{X0: f, Y1: 100 - f, X1: 100 - f, Y0: f, Series: i}, Circle{X: 50, Y: 50, R: f})
	}
	for b.Loop() {
		_ = m.View(th())
	}
}
