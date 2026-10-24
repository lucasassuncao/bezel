package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// playTo is t advanced from its first frame at start to start+d.
func playTo(t Timeline, start time.Time, d time.Duration) Timeline {
	t, _ = t.Advance(start)
	t, _ = t.Advance(start.Add(d))
	return t
}

func TestSeqPlaysStepsOneAfterAnother(t *testing.T) {
	moving(t)
	tl := NewTimeline(Seq(
		Step("a", 0, 10, 100*ms).WithEase(Linear),
		Step("b", 0, 10, 100*ms).WithEase(Linear),
	))
	start := time.Unix(10, 0)
	mid := playTo(tl, start, 50*ms)
	assert.InDelta(t, 5, mid.Value("a"), 1e-9)
	assert.InDelta(t, 0, mid.Value("b"), 1e-9, "b waits for a")
	late := playTo(tl, start, 150*ms)
	assert.InDelta(t, 10, late.Value("a"), 1e-9)
	assert.InDelta(t, 5, late.Value("b"), 1e-9)
}

func TestParAndStaggerOverlap(t *testing.T) {
	moving(t)
	tl := NewTimeline(Stagger(50*ms,
		Step("a", 0, 1, 100*ms).WithEase(Linear),
		Step("b", 0, 1, 100*ms).WithEase(Linear),
	))
	got := playTo(tl, time.Unix(10, 0), 75*ms)
	assert.InDelta(t, 0.75, got.Value("a"), 1e-9)
	assert.InDelta(t, 0.25, got.Value("b"), 1e-9)
	assert.Equal(t, 150*ms, tl.span.length)
	assert.Equal(t, 100*ms, Par(Step("a", 0, 1, 100*ms), Step("b", 0, 1, 40*ms)).length)
}

func TestALaterStepOfTheSameNameTakesOver(t *testing.T) {
	moving(t)
	tl := NewTimeline(Seq(
		Step("x", 0, 10, 100*ms).WithEase(Linear),
		Gap(50*ms),
		Step("x", 10, 0, 100*ms).WithEase(Linear),
	))
	start := time.Unix(10, 0)
	assert.InDelta(t, 10, playTo(tl, start, 120*ms).Value("x"), 1e-9, "held through the gap")
	assert.InDelta(t, 5, playTo(tl, start, 200*ms).Value("x"), 1e-9)
}

func TestATimelineEndsAtItsLastValues(t *testing.T) {
	moving(t)
	tl := NewTimeline(Step("x", 0, 4, 100*ms))
	start := time.Unix(10, 0)
	tl, _ = tl.Advance(start)
	tl, done := tl.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.False(t, tl.Active())
	assert.InDelta(t, 4, tl.Value("x"), 1e-9)
	assert.Equal(t, 1.0, tl.Progress("x"))
	assert.Zero(t, tl.Value("missing"))
}

func TestAStillTimelineIsBornAtItsEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	tl := NewTimeline(Seq(Step("x", 0, 3, 100*ms), Step("y", 1, 2, 100*ms)))
	assert.False(t, tl.Active())
	assert.Nil(t, tl.Tick())
	assert.InDelta(t, 3, tl.Value("x"), 1e-9)
	assert.InDelta(t, 2, tl.Value("y"), 1e-9)
}

func TestATimelineTakesOnlyItsOwnFrames(t *testing.T) {
	moving(t)
	tl := NewTimeline(Step("x", 0, 1, 100*ms))
	other := NewTimeline(Step("x", 0, 1, 100*ms))
	_, cmd := tl.Update(other.FrameAt(time.Unix(10, 0)))
	assert.Nil(t, cmd)
	next, cmd := tl.Update(tl.FrameAt(time.Unix(10, 0)))
	require.NotNil(t, cmd)
	assert.True(t, next.Active())
}

func TestATimelineRepeatsAndYoyos(t *testing.T) {
	moving(t)
	tl := NewTimeline(Step("x", 0, 10, 100*ms).WithEase(Linear)).WithYoyo().WithRepeat(Forever)
	got := playTo(tl, time.Unix(10, 0), 150*ms)
	assert.True(t, got.Active())
	assert.InDelta(t, 5, got.Value("x"), 1e-6, "on its way back")
}

func TestCascadeDelaysEachMore(t *testing.T) {
	moving(t)
	fs := Cascade(100*ms, NewFloat(0, 1, 100*ms), NewFloat(0, 1, 100*ms), NewFloat(0, 1, 100*ms))
	require.Len(t, fs, 3)
	start := time.Unix(10, 0)
	for i := range fs {
		fs[i], _ = fs[i].Advance(start)
		fs[i], _ = fs[i].Advance(start.Add(150 * ms))
	}
	assert.Equal(t, 1.0, fs[0].Cur)
	assert.Greater(t, fs[1].Cur, 0.0)
	assert.Equal(t, 0.0, fs[2].Cur, "the third has not started")
}
