package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPendulumSwingsBothWays(t *testing.T) {
	start := time.Unix(1000, 0)
	p := NewPendulum(4, time.Second)
	p, _ = p.Advance(start)
	q, _ := p.Advance(start.Add(250 * time.Millisecond))
	assert.InDelta(t, 4, q.Cur, 1e-9)
	q, _ = p.Advance(start.Add(750 * time.Millisecond))
	assert.InDelta(t, -4, q.Cur, 1e-9)
	_, done := p.Advance(start.Add(time.Hour))
	assert.False(t, done, "without decay it swings for ever")
}

func TestPendulumWithDecayComesToRest(t *testing.T) {
	start := time.Unix(1000, 0)
	p := NewPendulum(4, time.Second).WithDecay(200 * time.Millisecond)
	p, _ = p.Advance(start)
	p, done := p.Advance(start.Add(10 * time.Second))
	assert.True(t, done)
	assert.Equal(t, 0.0, p.Cur)
}
