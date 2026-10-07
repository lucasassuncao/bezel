package draw

import (
	"time"

	"charm.land/bubbles/v2/textinput"
)

// BlinkSpeed is how often the caret of every input bezel builds flips. A
// variable so a host's tests can shrink it: each keystroke arms one blink.
var BlinkSpeed = 530 * time.Millisecond

// WithCaret gives in the BlinkSpeed caret.
func WithCaret(in textinput.Model) textinput.Model {
	s := in.Styles()
	s.Cursor.BlinkSpeed = BlinkSpeed
	in.SetStyles(s)
	return in
}
