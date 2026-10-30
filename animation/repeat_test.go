package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// at drives f by its own frames: the first at start, then one at each offset.
func at(f Float, start time.Time, offsets ...time.Duration) Float {
	f, _ = f.Update(f.FrameAt(start))
	for _, d := range offsets {
		f, _ = f.Update(f.FrameAt(start.Add(d)))
	}
	return f
}

const ms = time.Millisecond

func TestADelayHoldsTheStartValue(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	f := NewFloat(0, 1, 100*ms).WithEase(Linear).WithDelay(100 * ms)
	f = at(f, start, 50*ms)
	assert.Equal(t, 0.0, f.cur)
	assert.True(t, f.Active(), "waiting is still moving")
	f = at(f, start, 50*ms, 150*ms)
	assert.InDelta(t, 0.5, f.cur, 1e-9)
}

func TestRepeatRunsAgainThenLands(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	f := NewFloat(0, 1, 100*ms).WithEase(Linear).WithRepeat(1)
	assert.InDelta(t, 0.5, at(f, start, 150*ms).cur, 1e-9, "the second run is half way")
	f = at(f, start, 250*ms)
	assert.False(t, f.Active())
	assert.Equal(t, 1.0, f.cur)
}

// A yoyo runs back on every other run, so an even count of runs lands on From.
func TestYoyoComesBack(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	f := NewFloat(0, 1, 100*ms).WithEase(Linear).WithYoyo(true).WithRepeat(1)
	assert.InDelta(t, 0.75, at(f, start, 25*ms, 125*ms).cur, 1e-9, "on the way back")
	f = at(f, start, 250*ms)
	assert.False(t, f.Active())
	assert.Equal(t, 0.0, f.cur)
	f, done := f.Advance(start.Add(time.Hour))
	assert.True(t, done)
	assert.Equal(t, 0.0, f.cur, "a landed animation stays where it landed")
}

func TestForeverNeverLands(t *testing.T) {
	moving(t)
	f := at(NewFloat(0, 1, 100*ms).WithYoyo(true).WithRepeat(Forever), time.Unix(1000, 0), time.Hour)
	assert.True(t, f.Active())
	assert.NotNil(t, f.Tick())
}

// Tween, the int one, takes the same three; a yoyo once round lands where it began.
func TestTweenTakesDelayRepeatAndYoyo(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	tw := NewTween(0, 10, 100*ms).WithEase(Linear).WithDelay(100 * ms).WithYoyo(true).WithRepeat(1)
	tw, _ = tw.Update(tw.FrameAt(start))
	tw, _ = tw.Update(tw.FrameAt(start.Add(150 * ms)))
	assert.Equal(t, 5, tw.cur)
	tw, _ = tw.Update(tw.FrameAt(start.Add(250 * ms)))
	assert.Equal(t, 5, tw.cur, "half way back")
	tw, res := tw.Update(tw.FrameAt(start.Add(time.Second)))
	assert.Nil(t, res.Cmd)
	assert.False(t, tw.Active())
	assert.Equal(t, 0, tw.cur)
}

// A fade passes the options on too: it pulses.
func TestAFadeRepeatsToo(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	fd := NewFade(black, white, 100*ms).WithEase(Linear).WithYoyo(true).WithRepeat(Forever)
	fd, _ = fd.Update(fd.FrameAt(start))
	fd, _ = fd.Update(fd.FrameAt(start.Add(150 * ms)))
	assert.True(t, fd.Active())
}
