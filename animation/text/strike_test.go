package text

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestStrikeCrossesOutFromTheLeft(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewStrike("buy milk", time.Second).WithInk(white)
	require.Equal(t, "buy milk", s.View())
	s, _ = s.Advance(start)
	s, done := s.Advance(start.Add(time.Second))
	require.True(t, done)
	require.Equal(t, "buy milk", ansi.Strip(s.View()))
	require.Contains(t, s.View(), "\x1b[", "the text is styled struck")
}
