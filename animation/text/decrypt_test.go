package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecryptLandsLettersFromTheLeft(t *testing.T) {
	start := time.Unix(1000, 0)
	d := NewDecrypt("GATE 12", time.Second)
	d, _ = d.Advance(start)
	d, _ = d.Advance(start.Add(400 * time.Millisecond))
	v := d.View()
	assert.Equal(t, "G", v[:1], "the first letter has landed")
	assert.NotEqual(t, "GATE 12", v, "the last ones still turn")
	assert.Equal(t, len("GATE 12"), len(v))

	d, done := d.Advance(start.Add(time.Second))
	assert.True(t, done)
	assert.Equal(t, "GATE 12", d.View())
}

// A letter turns through the flaps just before its own, in order.
func TestDecryptRollsThroughTheFlapsInOrder(t *testing.T) {
	d := Decrypt{text: []string{"c"}, p: 0.3 - 2*flip}
	assert.Equal(t, "a", d.flapAt(0, "c"), "two flaps before c, kept lowercase")
	assert.Equal(t, "é", d.flapAt(0, "é"), "a letter off the board shows as it is")
}
