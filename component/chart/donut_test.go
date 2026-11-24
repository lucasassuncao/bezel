package chart

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

func TestDonutIsARingTwiceAsWideAsItIsTall(t *testing.T) {
	th := theme.Resolve(theme.Plain(), true)
	rows := strings.Split(ansi.Strip(Donut(1, 4, 8, th)), "\n")
	require.Len(t, rows, 8)
	for _, r := range rows {
		require.Equal(t, 16, ansi.StringWidth(r), "%q", r)
	}
	require.Contains(t, rows[4], "25%")
	require.NotContains(t, rows[0][:2], "█", "the corners are outside the ring")
}

func TestDonutFillsClockwiseFromTheTop(t *testing.T) {
	th := theme.Resolve(theme.Plain(), true)
	quarter := Donut(1, 4, 10, th)
	full := Donut(4, 4, 10, th)
	empty := Donut(0, 4, 10, th)
	require.NotEqual(t, quarter, full)
	require.NotEqual(t, quarter, empty)
	require.Contains(t, ansi.Strip(Donut(5, 0, 6, th)), "0%", "no total is nothing filled")
	require.Contains(t, ansi.Strip(Donut(9, 4, 6, th)), "100%", "past the total is full")
}

func TestATinyDonutStillDrawsARing(t *testing.T) {
	th := theme.Resolve(theme.Plain(), true)
	rows := strings.Split(ansi.Strip(Donut(1, 2, 1, th)), "\n")
	require.Len(t, rows, 3, "three rows at least")
}
