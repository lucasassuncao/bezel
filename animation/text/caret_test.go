package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCaretBlinks(t *testing.T) {
	start := time.Unix(1000, 0)
	c := NewCaret("▌", time.Second)
	require.Equal(t, "▌", c.View())
	c, _ = c.Advance(start)
	c, _ = c.Advance(start.Add(700 * time.Millisecond))
	require.Equal(t, " ", c.View(), "hidden in the second half, as wide as the glyph")
	c, _ = c.Advance(start.Add(1200 * time.Millisecond))
	require.Equal(t, "▌", c.View())
}
