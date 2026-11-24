package text

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestBigDrawsThreeEvenRows(t *testing.T) {
	rows := strings.Split(Big("12:30"), "\n")
	require.Len(t, rows, 3)
	for _, r := range rows {
		require.Equal(t, 5*3+4, ansi.StringWidth(r), "%q", r)
	}
}

func TestBigDrawsAGlyphInHalfBlocks(t *testing.T) {
	require.Equal(t, "█▀█\n█ █\n▀▀▀", Big("0"))
	require.Equal(t, "▀█▀\n █ \n ▀ ", Big("t"), "letters are drawn in capitals")
}

func TestBigLeavesWhatItLacksBlank(t *testing.T) {
	require.Equal(t, "   \n   \n   ", Big("é"))
	require.Equal(t, "\n\n", Big(""))
}

func TestEveryGlyphIsThreeWideAndFiveTall(t *testing.T) {
	for r, g := range font {
		for _, row := range g {
			require.Len(t, row, 3, "%q", r)
		}
	}
}
