package draw

import (
	"cmp"
	"time"
)

// DefaultBlink is how often a caret flips when nothing says otherwise.
const DefaultBlink = 530 * time.Millisecond

// Blink is every, or DefaultBlink when every is zero.
func Blink(every time.Duration) time.Duration { return cmp.Or(every, DefaultBlink) }
