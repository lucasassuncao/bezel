package progress

import (
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestGaugeRulesATrackAndPointsTheNeedle(t *testing.T) {
	lines := strings.Split(ansi.Strip(Gauge(50, 100, 9, th())), "\n")
	assert.Equal(t, "├─┼─┼─┼─┤", lines[0])
	assert.Equal(t, "    ▲    ", lines[1])
}

func TestGaugeKeepsTheNeedleOnTheTrack(t *testing.T) {
	assert.Equal(t, "        ▲", strings.Split(ansi.Strip(Gauge(500, 100, 9, th())), "\n")[1])
	assert.Equal(t, "▲        ", strings.Split(ansi.Strip(Gauge(-5, 100, 9, th())), "\n")[1])
	assert.Equal(t, "▲        ", strings.Split(ansi.Strip(Gauge(5, 0, 9, th())), "\n")[1], "no total reads as nothing")
}

// NaN survived min and max and gave strings.Repeat a negative count.
func TestGaugeReadsNaNAsNothing(t *testing.T) {
	assert.Equal(t, "▲        ", strings.Split(ansi.Strip(Gauge(math.NaN(), 100, 9, th())), "\n")[1])
	assert.Equal(t, "▲        ", strings.Split(ansi.Strip(Gauge(math.Inf(1), math.Inf(1), 9, th())), "\n")[1])
}
