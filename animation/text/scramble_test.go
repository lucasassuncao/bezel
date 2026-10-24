package text

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestScrambleSettlesFromTheLeft(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewScramble("ab-cd", 100*ms)
	assert.Equal(t, 5, ansi.StringWidth(s.View()))
	assert.Equal(t, "-", string([]rune(s.View())[2]), "punctuation is never scrambled")

	s, _ = s.Update(s.FrameAt(start))
	s, _ = s.Update(s.FrameAt(start.Add(60 * ms)))
	assert.True(t, strings.HasPrefix(s.View(), "ab-"), "three of five have settled")
	assert.Equal(t, s.View(), s.View(), "a frame draws the same every time")

	s, cmd := s.Update(s.FrameAt(start.Add(time.Second)))
	assert.Nil(t, cmd)
	assert.Equal(t, "ab-cd", s.View())
	assert.False(t, s.Active())
}

// The unsettled letters change from frame to frame, drawn from the charset.
func TestScrambleChurnsItsUnsettledLetters(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewScramble("abcdefgh", time.Second).WithCharset("01")
	s, _ = s.Update(s.FrameAt(start))
	seen := map[string]bool{}
	for i := 1; i <= 5; i++ {
		s, _ = s.Update(s.FrameAt(start.Add(time.Duration(i) * 16 * ms)))
		seen[s.View()] = true
		assert.Equal(t, "", strings.Trim(s.View(), "01"), "only the charset shows")
	}
	assert.Greater(t, len(seen), 1)
}
