package draw

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompositeReplacesOnlyTheCoveredCells(t *testing.T) {
	bg := "0123456789\n0123456789\n0123456789"
	fg := "XX\nXX"
	got := Composite(fg, bg, 3, 1)
	require.Equal(t, "0123456789\n012XX56789\n012XX56789", got)
}

func TestCompositePadsAShortBackgroundLine(t *testing.T) {
	got := Composite("XX", "ab", 5, 0)
	require.Equal(t, "ab   XX", got)
}

func TestCompositeCenterPlacesInTheMiddle(t *testing.T) {
	bg := strings.Repeat(".........\n", 4) + "........."
	got := CompositeCenter("XXX\nXXX", bg)
	lines := strings.Split(got, "\n")
	require.Equal(t, "...XXX...", lines[1])
	require.Equal(t, "...XXX...", lines[2])
	require.Equal(t, ".........", lines[0])
}
