package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEllipsisGrowsAndKeepsItsWidth(t *testing.T) {
	start := time.Unix(1000, 0)
	e := NewEllipsis(3, time.Second)
	e, _ = e.Advance(start)
	require.Equal(t, "   ", e.View())
	e, _ = e.Advance(start.Add(550 * time.Millisecond))
	require.Equal(t, ".. ", e.View())
	e, _ = e.Advance(start.Add(990 * time.Millisecond))
	require.Equal(t, "...", e.View())
}
