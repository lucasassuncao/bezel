package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGrowRaisesEveryValueFromZero(t *testing.T) {
	start := time.Unix(1000, 0)
	s := Grow([]float64{10, 20}, 100*ms)
	require.Equal(t, []float64{0, 0}, s.Values())
	s, _ = s.Advance(start)
	s, done := s.Advance(start.Add(time.Second))
	require.True(t, done)
	require.Equal(t, []float64{10, 20}, s.Values())
}

func TestDrawRevealsFromTheLeft(t *testing.T) {
	start := time.Unix(1000, 0)
	s := Draw([]float64{4, 4, 4, 4}, 100*ms)
	s.run.Ease = Linear
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(50 * ms))
	require.Equal(t, []float64{4, 4, 0, 0}, s.Values())
}

func TestPushRisesAtTheRightAndKeepsAWindow(t *testing.T) {
	start := time.Unix(1000, 0)
	s := Grow([]float64{1, 2, 3}, 100*ms)
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(time.Second))
	s, cmd := s.Push(9, 3)
	require.NotNil(t, cmd)
	require.Equal(t, []float64{2, 3, 0}, s.Values(), "the oldest dropped, the new one at zero")
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(time.Second))
	require.Equal(t, []float64{2, 3, 9}, s.Values())
}

func TestSeriesWithDurationTimesTheNextPush(t *testing.T) {
	start := time.Unix(1000, 0)
	s, _ := Grow(nil, 0).WithDuration(100*ms).WithEase(Linear).Push(10, 0)
	require.True(t, s.Active())
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(25 * ms))
	require.Equal(t, []float64{2.5}, s.Values(), "linear, as asked, past the push")
}
