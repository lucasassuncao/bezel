package edit

import (
	"cmp"
	"time"
)

// defaultBlink is how often a caret flips when the theme says nothing.
const defaultBlink = 530 * time.Millisecond

// Blink is every, or the default when every is zero, so a zero theme never
// ticks in a tight loop.
func Blink(every time.Duration) time.Duration { return cmp.Or(every, defaultBlink) }
