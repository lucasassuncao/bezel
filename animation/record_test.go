package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecordKeepsEveryFrameToTheEnd(t *testing.T) {
	frames := Record(NewTypewriter("abc", 20*time.Millisecond), Typewriter.View, time.Second)
	assert.Equal(t, "", frames[0])
	assert.Equal(t, "abc", frames[len(frames)-1])
	assert.Less(t, len(frames), 10, "it stops when the animation ends")
}

func TestRecordStopsAtItsLimit(t *testing.T) {
	frames := Record(NewSnow(white), func(a Ambient) string { return a.View(4, 2) }, 5*Frame)
	assert.Len(t, frames, 6)
}
