package text

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestShakeSwingsAndSettles(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewShake(3, 300*time.Millisecond)
	assert.Equal(t, 0, s.Offset())

	s, _ = s.Update(s.FrameAt(start))
	s, _ = s.Update(s.FrameAt(start.Add(25 * time.Millisecond))) // the first swing's peak
	assert.Equal(t, 3, s.Offset())
	s, res := s.Update(s.FrameAt(start.Add(time.Second)))
	assert.Nil(t, res.Cmd)
	assert.Equal(t, 0, s.Offset())
	assert.False(t, s.Active())
}

func TestShakeViewKeepsItsWidth(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewShake(2, 300*time.Millisecond)
	assert.Equal(t, "  ab  \n  cd  ", s.View("ab\ncd"), "at rest the block sits in the middle")
	s, _ = s.Update(s.FrameAt(start))
	for ms := 0; ms < 300; ms += 16 {
		s, _ = s.Update(s.FrameAt(start.Add(time.Duration(ms) * time.Millisecond)))
		for _, l := range strings.Split(s.View("ab\ncd"), "\n") {
			assert.Equal(t, 6, ansi.StringWidth(l))
		}
	}
}
