package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestARaceStartsPlacedLargestFirst(t *testing.T) {
	r := NewRace([]string{"a", "b", "c"}, []float64{1, 3, 2}, 100*ms)
	assert.False(t, r.Active())
	got := r.Racers()
	assert.Equal(t, []Racer{{"a", 1, 2}, {"b", 3, 0}, {"c", 2, 1}}, got)
}

func TestARacerOvertakesAnother(t *testing.T) {
	moving(t)
	r := NewRace([]string{"a", "b"}, []float64{5, 1}, 100*ms).WithEase(Linear)
	r, cmd := r.Set([]float64{5, 9})
	require.NotNil(t, cmd)
	start := time.Unix(10, 0)
	r, _ = r.Advance(start)
	mid, _ := r.Advance(start.Add(50 * ms))
	got := mid.Racers()
	assert.InDelta(t, 0.5, got[0].Place, 1e-9, "a on its way down")
	assert.InDelta(t, 0.5, got[1].Place, 1e-9, "b on its way up")
	assert.InDelta(t, 5, got[1].Value, 1e-9)
	end, done := r.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, []Racer{{"a", 5, 1}, {"b", 9, 0}}, end.Racers())
}

func TestASetMidRaceStartsFromWhereTheBarsAre(t *testing.T) {
	moving(t)
	r, _ := NewRace([]string{"a"}, []float64{0}, 100*ms).WithEase(Linear).Set([]float64{10})
	start := time.Unix(10, 0)
	r, _ = r.Advance(start)
	r, _ = r.Advance(start.Add(50 * ms))
	r, _ = r.Set([]float64{20})
	assert.InDelta(t, 5, r.Racers()[0].Value, 1e-9)
}

func TestAStillRaceJumps(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	r, cmd := NewRace([]string{"a", "b"}, []float64{1, 2}, time.Second).Set([]float64{3, 0})
	assert.Nil(t, cmd)
	assert.Equal(t, []Racer{{"a", 3, 0}, {"b", 0, 1}}, r.Racers())
}

func TestARaceTakesOnlyItsOwnFrames(t *testing.T) {
	moving(t)
	r, _ := NewRace([]string{"a"}, []float64{0}, 100*ms).Set([]float64{1})
	_, res := r.Update(NewFloat(0, 1, ms).FrameAt(time.Unix(1, 0)))
	assert.Nil(t, res.Cmd)
	_, res = r.Update(r.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, res.Cmd)
}
