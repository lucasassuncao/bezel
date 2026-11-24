package chart

import (
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestHeatDrawsTwoColumnsACell(t *testing.T) {
	out := Heat([][]float64{{0, 5, 10}, {10, 0, 5}}, 10, th())
	lines := strings.Split(ansi.Strip(out), "\n")
	require.Equal(t, []string{"··████", "██··██"}, lines)
	require.NotEqual(t, ansi.Strip(out), out, "the cells are coloured")
}

func TestHeatWithoutATopIsAllDots(t *testing.T) {
	require.Equal(t, "····", ansi.Strip(Heat([][]float64{{3, 4}}, 0, th())))
}

// NaN passes a share <= 0 check and indexed the ramp with int(NaN).
func TestHeatDrawsANumberItCannotPlaceAsADot(t *testing.T) {
	require.Equal(t, "··██", ansi.Strip(Heat([][]float64{{math.NaN(), 10}}, 10, th())))
	require.Equal(t, "··", ansi.Strip(Heat([][]float64{{math.Inf(1)}}, math.Inf(1), th())))
}
