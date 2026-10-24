package text

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestStrikeCrossesOutFromTheLeft(t *testing.T) {
	start := time.Unix(1000, 0)
	s := NewStrike("buy milk", time.Second).WithInk(white)
	assert.Equal(t, "buy milk", s.View())
	s, _ = s.Advance(start)
	s, done := s.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, "buy milk", ansi.Strip(s.View()))
	assert.Contains(t, s.View(), "\x1b[", "the text is styled struck")
}
