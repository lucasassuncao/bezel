package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOdometerRollsEveryDigitToTheNewValue(t *testing.T) {
	start := time.Unix(1000, 0)
	o := NewOdometer(0, 123, 300*ms)
	assert.Equal(t, "000", o.View(), "padded to the wider of the two")

	o, _ = o.Update(o.FrameAt(start))
	left, right := map[byte]bool{}, map[byte]bool{}
	for d := 16 * ms; d < 300*ms; d += 16 * ms {
		o, _ = o.Update(o.FrameAt(start.Add(d)))
		v := o.View()
		assert.Len(t, v, 3)
		left[v[0]], right[v[2]] = true, true
	}
	assert.Greater(t, len(right), len(left), "the right-hand digit spins more")

	o, cmd := o.Update(o.FrameAt(start.Add(time.Second)))
	assert.Nil(t, cmd)
	assert.Equal(t, "123", o.View())
	assert.Equal(t, 123, o.Value())
}

func TestOdometerRollsOnToTheNextValue(t *testing.T) {
	start := time.Unix(1000, 0)
	o, cmd := NewOdometer(5, 5, 300*ms).To(1284)
	assert.NotNil(t, cmd)
	assert.Equal(t, "0005", o.View())
	o, _ = o.Update(o.FrameAt(start))
	o, _ = o.Update(o.FrameAt(start.Add(time.Hour)))
	assert.Equal(t, "1284", o.View())

	assert.Equal(t, "0", NewOdometer(-5, -1, time.Second).View(), "below zero reads as zero")
}
