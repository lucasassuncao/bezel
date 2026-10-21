package chart

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

func th() theme.Resolved { return theme.Resolve(theme.Gruvbox, true) }

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
