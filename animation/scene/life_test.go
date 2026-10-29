package scene

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLifeFollowsConwaysRules(t *testing.T) {
	blinker := [][]bool{
		{false, false, false, false, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
		{false, false, false, false, false},
	}
	next := step(blinker, 5, 5)
	assert.Equal(t, []bool{false, true, true, true, false}, next[2], "a vertical blinker turns horizontal")
	assert.Equal(t, blinker, step(next, 5, 5))
}

func TestLifeIsTheSameBoardForTheSameTime(t *testing.T) {
	a := playedFor(Life(white), 3).View(20, 8)
	assert.Equal(t, a, playedFor(Life(white), 3).View(20, 8))
	assert.NotEqual(t, a, playedFor(Life(white), 3.5).View(20, 8))
}

func TestASeededBoardIsAboutAThirdAlive(t *testing.T) {
	alive := 0
	for _, row := range seedBoard(1, 50, 20) {
		for _, c := range row {
			if c {
				alive++
			}
		}
	}
	assert.InDelta(t, 333, alive, 80)
}
