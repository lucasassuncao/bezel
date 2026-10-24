package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGrowRaisesEveryValueFromZero(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewGrow([]float64{10, 20}, 100*ms)
	assert.Equal(t, []float64{0, 0}, s.Values())
	s, _ = s.Advance(start)
	s, done := s.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, []float64{10, 20}, s.Values())
}

func TestDrawRevealsFromTheLeft(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewDraw([]float64{4, 4, 4, 4}, 100*ms)
	s.run.ease = Linear
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(50 * ms))
	assert.Equal(t, []float64{4, 4, 0, 0}, s.Values())
}

func TestPushRisesAtTheRightAndKeepsAWindow(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewGrow([]float64{1, 2, 3}, 100*ms)
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(time.Second))
	s, cmd := s.Push(9, 3)
	assert.NotNil(t, cmd)
	assert.Equal(t, []float64{2, 3, 0}, s.Values(), "the oldest dropped, the new one at zero")
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(time.Second))
	assert.Equal(t, []float64{2, 3, 9}, s.Values())
}

func TestSeriesWithDurationTimesTheNextPush(t *testing.T) {
	start := time.Unix(1000, 0)
	s, _ := NewGrow(nil, 0).WithDuration(100*ms).WithEase(Linear).Push(10, 0)
	assert.True(t, s.Active())
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(25 * ms))
	assert.Equal(t, []float64{2.5}, s.Values(), "linear, as asked, past the push")
}
