package draw

import (
	"cmp"
	"time"

	"charm.land/bubbles/v2/textinput"
)

// DefaultBlink is how often a caret flips when nothing says otherwise.
const DefaultBlink = 530 * time.Millisecond

// Blink is every, or DefaultBlink when every is zero.
func Blink(every time.Duration) time.Duration { return cmp.Or(every, DefaultBlink) }

// WithCaret gives in a caret that flips every Blink(every). Each keystroke
// arms one blink, so a host's tests pass a short one through the theme.
func WithCaret(in textinput.Model, every time.Duration) textinput.Model {
	s := in.Styles()
	s.Cursor.BlinkSpeed = Blink(every)
	in.SetStyles(s)
	return in
}
