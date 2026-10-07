package draw

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
)

// A host shrinks BlinkSpeed in its tests; an input that ignored it would make
// every keystroke wait out the 530ms default.
func TestWithCaretUsesBlinkSpeed(t *testing.T) {
	old := BlinkSpeed
	t.Cleanup(func() { BlinkSpeed = old })
	BlinkSpeed = time.Millisecond

	if got := WithCaret(textinput.New()).Styles().Cursor.BlinkSpeed; got != time.Millisecond {
		t.Errorf("caret blinks every %v, want %v", got, time.Millisecond)
	}
}
