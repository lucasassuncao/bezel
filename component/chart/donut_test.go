package chart

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/theme"
)

func TestDonutIsARingTwiceAsWideAsItIsTall(t *testing.T) {
	th := theme.Resolve(theme.Plain, true)
	rows := strings.Split(ansi.Strip(Donut(1, 4, 8, th)), "\n")
	assert.Len(t, rows, 8)
	for _, r := range rows {
		assert.Equal(t, 16, ansi.StringWidth(r), "%q", r)
	}
	assert.Contains(t, rows[4], "25%")
	assert.NotContains(t, rows[0][:2], "█", "the corners are outside the ring")
}

func TestDonutFillsClockwiseFromTheTop(t *testing.T) {
	th := theme.Resolve(theme.Plain, true)
	quarter := Donut(1, 4, 10, th)
	full := Donut(4, 4, 10, th)
	empty := Donut(0, 4, 10, th)
	assert.NotEqual(t, quarter, full)
	assert.NotEqual(t, quarter, empty)
	assert.Contains(t, ansi.Strip(Donut(5, 0, 6, th)), "0%", "no total is nothing filled")
	assert.Contains(t, ansi.Strip(Donut(9, 4, 6, th)), "100%", "past the total is full")
}

func TestATinyDonutStillDrawsARing(t *testing.T) {
	th := theme.Resolve(theme.Plain, true)
	rows := strings.Split(ansi.Strip(Donut(1, 2, 1, th)), "\n")
	assert.Len(t, rows, 3, "three rows at least")
}
