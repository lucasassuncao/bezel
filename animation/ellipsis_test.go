package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEllipsisGrowsAndKeepsItsWidth(t *testing.T) {
	start := time.Unix(1000, 0)
	e := NewEllipsis(3, time.Second)
	e, _ = e.Advance(start)
	assert.Equal(t, "   ", e.View())
	e, _ = e.Advance(start.Add(550 * time.Millisecond))
	assert.Equal(t, ".. ", e.View())
	e, _ = e.Advance(start.Add(990 * time.Millisecond))
	assert.Equal(t, "...", e.View())
}
