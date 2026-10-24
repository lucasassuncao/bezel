package text

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestBigDrawsThreeEvenRows(t *testing.T) {
	rows := strings.Split(Big("12:30"), "\n")
	assert.Len(t, rows, 3)
	for _, r := range rows {
		assert.Equal(t, 5*3+4, ansi.StringWidth(r), "%q", r)
	}
}

func TestBigDrawsAGlyphInHalfBlocks(t *testing.T) {
	assert.Equal(t, "█▀█\n█ █\n▀▀▀", Big("0"))
	assert.Equal(t, "▀█▀\n █ \n ▀ ", Big("t"), "letters are drawn in capitals")
}

func TestBigLeavesWhatItLacksBlank(t *testing.T) {
	assert.Equal(t, "   \n   \n   ", Big("é"))
	assert.Equal(t, "\n\n", Big(""))
}

func TestEveryGlyphIsThreeWideAndFiveTall(t *testing.T) {
	for r, g := range font {
		for _, row := range g {
			assert.Len(t, row, 3, "%q", r)
		}
	}
}
