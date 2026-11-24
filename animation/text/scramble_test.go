package text

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestScrambleSettlesFromTheLeft(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewScramble("ab-cd", 100*ms)
	require.Equal(t, 5, ansi.StringWidth(s.View()))
	require.Equal(t, "-", string([]rune(s.View())[2]), "punctuation is never scrambled")

	s, _ = s.Update(s.FrameAt(start))
	s, _ = s.Update(s.FrameAt(start.Add(60 * ms)))
	require.True(t, strings.HasPrefix(s.View(), "ab-"), "three of five have settled")
	require.Equal(t, s.View(), s.View(), "a frame draws the same every time")

	s, res := s.Update(s.FrameAt(start.Add(time.Second)))
	require.Nil(t, res.Cmd)
	require.Equal(t, "ab-cd", s.View())
	require.False(t, s.Active())
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
		require.Equal(t, "", strings.Trim(s.View(), "01"), "only the charset shows")
	}
	require.Greater(t, len(seen), 1)
}
