package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAStreamRevealsAtItsPace(t *testing.T) {
	s, cmd := NewStream(10 * ms).Push("hello")
	require.NotNil(t, cmd)
	start := time.Unix(10, 0)
	s, _ = s.Advance(start)
	require.Equal(t, "", s.View())
	s, _ = s.Advance(start.Add(30 * ms))
	require.Equal(t, "hel", s.View())
	s, done := s.Advance(start.Add(time.Second))
	require.True(t, done)
	require.Equal(t, "hello", s.View())
	require.Nil(t, s.Tick(), "caught up, it asks for no frame")
}

func TestAZeroStreamRevealsAtTheDefaultPace(t *testing.T) {
	var s Stream
	s, _ = s.Push("hello")
	start := time.Unix(10, 0)
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(30 * ms))
	require.Equal(t, "hel", s.View())
}

func TestAClockSteppingBackRevealsNothingMore(t *testing.T) {
	s, _ := NewStream(10 * ms).Push("hello")
	start := time.Unix(10, 0)
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(20 * ms))
	s, _ = s.Advance(start.Add(-time.Second))
	require.Equal(t, "he", s.View())
}

func TestAPushWhileRevealingJoinsTheQueue(t *testing.T) {
	s, _ := NewStream(10 * ms).Push("ab")
	s, cmd := s.Push("cd")
	require.Nil(t, cmd, "already running: no second chain of frames")
	require.Equal(t, 4, s.Behind())
	start := time.Unix(10, 0)
	s, _ = s.Advance(start)
	s, _ = s.Advance(start.Add(time.Second))
	require.Equal(t, "abcd", s.View())
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
	require.Greater(t, len(s.View()), 10, "more than one a step: it is catching up")
}

func TestAStreamTakesOnlyItsOwnFrames(t *testing.T) {
	s, _ := NewStream(10 * ms).Push("abc")
	other, _ := NewStream(10 * ms).Push("xyz")
	_, res := s.Update(other.FrameAt(time.Unix(10, 0)))
	require.Nil(t, res.Cmd)
	s, res = s.Update(s.FrameAt(time.Unix(10, 0)))
	require.NotNil(t, res.Cmd)
	require.True(t, s.Active())
}

func TestSkipShowsEverythingAndStillShowsAtOnce(t *testing.T) {
	s, _ := NewStream(10 * ms).Push("abc")
	s = s.Skip()
	require.Equal(t, "abc", s.View())
	require.False(t, s.Active())

	t.Setenv("BEZEL_NO_MOTION", "1")
	still, cmd := NewStream(0).Push("now")
	require.Nil(t, cmd)
	require.Equal(t, "now", still.View())
}
