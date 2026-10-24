package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAStreamRevealsAtItsPace(t *testing.T) {
	s, cmd := NewStream(10 * ms).Push("hello")
	require.NotNil(t, cmd)
	start := time.Unix(10, 0)
	s, _ = s.Advance(start)
	assert.Equal(t, "", s.View())
	s, _ = s.Advance(start.Add(30 * ms))
	assert.Equal(t, "hel", s.View())
	s, done := s.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, "hello", s.View())
	assert.Nil(t, s.Tick(), "caught up, it asks for no frame")
}

func TestAPushWhileRevealingJoinsTheQueue(t *testing.T) {
	s, _ := NewStream(10 * ms).Push("ab")
	s, cmd := s.Push("cd")
	assert.Nil(t, cmd, "already running: no second chain of frames")
	assert.Equal(t, 4, s.Behind())
	start := time.Unix(10, 0)
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(time.Second))
	assert.Equal(t, "abcd", s.View())
}

func TestAStreamFarBehindSpeedsUp(t *testing.T) {
	long := make([]byte, 600)
	for i := range long {
		long[i] = 'x'
	}
	s, _ := NewStream(10 * ms).Push(string(long))
	start := time.Unix(10, 0)
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(100 * ms))
	assert.Greater(t, len(s.View()), 10, "more than one a step: it is catching up")
}

func TestAStreamTakesOnlyItsOwnFrames(t *testing.T) {
	s, _ := NewStream(10 * ms).Push("abc")
	other, _ := NewStream(10 * ms).Push("xyz")
	_, cmd := s.Update(other.FrameAt(time.Unix(10, 0)))
	assert.Nil(t, cmd)
	s, cmd = s.Update(s.FrameAt(time.Unix(10, 0)))
	assert.NotNil(t, cmd)
	assert.True(t, s.Active())
}

func TestSkipShowsEverythingAndStillShowsAtOnce(t *testing.T) {
	s, _ := NewStream(10 * ms).Push("abc")
	s = s.Skip()
	assert.Equal(t, "abc", s.View())
	assert.False(t, s.Active())

	t.Setenv("BEZEL_NO_MOTION", "1")
	still, cmd := NewStream(0).Push("now")
	assert.Nil(t, cmd)
	assert.Equal(t, "now", still.View())
}
