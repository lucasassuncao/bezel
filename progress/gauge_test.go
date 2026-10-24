package progress

import (
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
