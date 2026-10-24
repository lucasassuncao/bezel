package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRubberGivesPastItsLimitsAndSpringsBack(t *testing.T) {
	r := NewRubber(0, 10).WithGive(4)
	assert.InDelta(t, 5.0, r.Pull(5).Pos(), 0, "inside the limits it goes where pulled")
	far := r.Pull(1000)
	assert.Less(t, far.Pos(), 14.0, "it never gets more than give past a limit")
	assert.Greater(t, far.Pos(), 13.0)
	assert.InDelta(t, -2, r.Pull(-4).Pos(), 1e-9, "4 past the floor resists to 2 past it")

	start := time.Unix(1000, 0)
	back, cmd := far.Release()
	assert.NotNil(t, cmd)
	back, _ = back.Advance(start)
	back, done := back.Advance(start.Add(5 * time.Second))
	assert.True(t, done)
	assert.Equal(t, 10.0, back.Pos())
}
