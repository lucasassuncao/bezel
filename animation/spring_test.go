package animation

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
)

// settle drives s with 16ms frames until it rests, and the highest Pos seen.
func settle(t *testing.T, s Spring, cmd tea.Cmd) (Spring, float64) {
	t.Helper()
	res := bezel.Result{Cmd: cmd}
	at, peak := time.Unix(1000, 0), s.pos
	for i := 0; res.Cmd != nil; i++ {
		require.Less(t, i, 1000, "the spring never came to rest")
		s, res = s.Update(s.FrameAt(at))
		at = at.Add(Frame)
		peak = max(peak, s.pos)
	}
	return s, peak
}

func TestSpringComesToRestOnItsTarget(t *testing.T) {
	s, cmd := NewSpring(0).To(10)
	assert.True(t, s.Active())
	s, peak := settle(t, s, cmd)
	assert.Equal(t, 10.0, s.pos)
	assert.Equal(t, 0.0, s.vel)
	assert.False(t, s.Active())
	assert.LessOrEqual(t, peak, 10.0, "a critically damped spring does not overshoot")
	assert.Equal(t, 10, s.Int())
}

func TestALooseSpringOvershoots(t *testing.T) {
	s, cmd := NewSpring(0).WithDamping(0.3).To(10)
	s, peak := settle(t, s, cmd)
	assert.Greater(t, peak, 10.5)
	assert.Equal(t, 10.0, s.pos)
}

// Retargeting mid-flight keeps the motion: no second clock, no lost speed.
func TestRetargetingKeepsTheSpringMoving(t *testing.T) {
	at := time.Unix(1000, 0)
	s, _ := NewSpring(0).To(10)
	for range 5 {
		s, _ = s.Update(s.FrameAt(at))
		at = at.Add(Frame)
	}
	vel := s.vel
	require.Positive(t, vel)
	s, cmd := s.To(20)
	assert.Nil(t, cmd, "its frames are already coming")
	assert.Equal(t, vel, s.vel)
	s, _ = settle(t, s, s.Tick())
	assert.Equal(t, 20.0, s.pos)
}

func TestASpringAtRestAsksForNothing(t *testing.T) {
	s, cmd := NewSpring(5).To(5)
	assert.Nil(t, cmd)
	assert.False(t, s.Active())
	var zero Spring
	_, cmd = zero.To(3)
	assert.NotNil(t, cmd, "the zero Spring moves too")
}
