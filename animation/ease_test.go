package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEasesRunFromZeroToOneWithoutTurningBack(t *testing.T) {
	for name, e := range map[string]Ease{"linear": Linear, "in": EaseIn, "out": EaseOut, "inout": EaseInOut} {
		assert.InDelta(t, 0, e(0), 1e-9, name)
		assert.InDelta(t, 1, e(1), 1e-9, name)
		prev := 0.0
		for i := 1; i <= 20; i++ {
			v := e(float64(i) / 20)
			assert.GreaterOrEqual(t, v, prev, "%s turned back at %d", name, i)
			prev = v
		}
	}
	assert.Less(t, EaseIn(0.25), 0.25, "in starts slow")
	assert.Greater(t, EaseOut(0.25), 0.25, "out starts fast")
}

func TestATweenFollowsItsEase(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := Tween{From: 0, To: 100, Start: start, Dur: 100 * time.Millisecond}
	v, _ := tw.At(start.Add(25 * time.Millisecond))
	assert.Equal(t, 13, v, "the default is EaseInOut")
	v, _ = tw.WithEase(Linear).At(start.Add(25 * time.Millisecond))
	assert.Equal(t, 25, v)
}
