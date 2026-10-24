package animation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func showFloat(f Float) string { return fmt.Sprintf("%.1f", f.Cur) }

func TestRecordKeepsEveryFrameToTheEnd(t *testing.T) {
	frames := Record(NewFloat(0, 1, 50*time.Millisecond).WithEase(Linear), showFloat, time.Second)
	assert.Equal(t, "0.0", frames[0])
	assert.Equal(t, "1.0", frames[len(frames)-1])
	assert.Less(t, len(frames), 10, "it stops when the animation ends")
}

func TestRecordStopsAtItsLimit(t *testing.T) {
	frames := Record(NewFloat(0, 1, 50*time.Millisecond).WithRepeat(Forever), showFloat, 5*Frame)
	assert.Len(t, frames, 6)
}
