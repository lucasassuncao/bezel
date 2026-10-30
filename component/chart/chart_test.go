package chart

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

// rich is a theme with a whole series and ramp of distinct colours.
var rich = theme.Theme{Colors: theme.Colors{
	Accent: "#FE8019", Selection: "#FABD2F", Border: "#7C6F64", Dim: "#A89984", Text: "#EBDBB2", Background: "#282828",
	Success: "#B8BB26", Warning: "#FABD2F", Danger: "#FB4934", Info: "#83A598", OnAccent: "#282828",
	Series: [6]string{"#B8BB26", "#8EC07C", "#FE8019", "#83A598", "#FABD2F", "#D3869B"},
	Ramp:   [3]string{"#B8BB26", "#FABD2F", "#FE8019"},
}}

func th() theme.Resolved { return theme.Resolve(rich, true) }

func TestAreaStacksEighthsFromTheBottom(t *testing.T) {
	require.Equal(t, "  █▄\n ███", ansi.Strip(Area([]float64{0, 8, 16, 12}, 16, 4, 2, th())))
}

func TestAreaShowsTheNewestValues(t *testing.T) {
	require.Equal(t, "▄█", ansi.Strip(Area([]float64{8, 8, 4, 8}, 8, 2, 1, th())))
	require.Equal(t, "  ▄█", ansi.Strip(Area([]float64{4, 8}, 8, 4, 1, th())), "fewer values than columns sit on the right")
}

func TestAreaColoursItsRowsAlongTheRamp(t *testing.T) {
	rows := strings.Split(Area([]float64{16, 16}, 16, 2, 2, th()), "\n")
	require.Contains(t, rows[1], th().Ramp[0].Render("██"), "the bottom row")
	require.Contains(t, rows[0], th().Ramp[2].Render("██"), "the top row")
}

func TestAreaInASeriesColour(t *testing.T) {
	require.Contains(t, Area([]float64{8}, 8, 1, 1, th(), InSeries(1)), th().Series[1].Render("█"))
}

func TestSparkIsOneRow(t *testing.T) {
	require.Equal(t, " ▂▄█", ansi.Strip(Spark([]float64{0, 2, 4, 8}, 8, 4, th())))
	require.Equal(t, "█", ansi.Strip(Spark([]float64{99}, 8, 1, th())), "over the top stays in the row")
}

func TestChartsOfNothing(t *testing.T) {
	require.Empty(t, Area([]float64{1}, 8, 0, 3, th()))
	require.Empty(t, Spark([]float64{1}, 8, 0, th()))
	require.Equal(t, "    ", ansi.Strip(Spark([]float64{1, 2}, 0, 4, th())), "no top draws nothing")
}

func TestANilOptionIsSkipped(t *testing.T) {
	require.Equal(t, ansi.Strip(Spark([]float64{1, 2}, 2, 2, th())), ansi.Strip(Spark([]float64{1, 2}, 2, 2, th(), nil)))
}

// Whatever the values, a chart keeps to the cells it was given.
func FuzzChartsStayInTheirBox(f *testing.F) {
	f.Add(3.0, -2.0, 1e9, 10.0, 12, 4)
	f.Add(0.0, 0.0, 0.0, 0.0, 1, 1)
	f.Add(-5.0, 7.5, 2.0, 0.0, 0, 3)
	f.Fuzz(func(t *testing.T, a, b, c, top float64, w, h int) {
		w, h = w%100, h%20
		vs := []float64{a, b, c, a + b, c - a}
		within := func(name, out string, rows int) {
			lines := strings.Split(out, "\n")
			if rows > 0 {
				require.LessOrEqual(t, len(lines), rows, name)
			}
			for i, l := range lines {
				require.LessOrEqual(t, ansi.StringWidth(l), max(w, 0), "%s row %d", name, i)
			}
		}
		within("area", Area(vs, top, w, h, th()), max(h, 1))
		within("spark", Spark(vs, top, w, th()), 1)
		within("bars", Bars([]Bar{{Label: "a", Value: a}, {Label: "bb", Value: b}}, top, w, th()), 0)
	})
}
