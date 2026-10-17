package draw

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
)

// A host shrinks the blink in its tests; an input that ignored it would make
// every keystroke wait out the 530ms default.
func TestWithCaretUsesTheBlinkGiven(t *testing.T) {
	if got := WithCaret(textinput.New(), time.Millisecond).Styles().Cursor.BlinkSpeed; got != time.Millisecond {
		t.Errorf("caret blinks every %v, want %v", got, time.Millisecond)
	}
}

func TestWithCaretDefaultsAZeroBlink(t *testing.T) {
	if got := WithCaret(textinput.New(), 0).Styles().Cursor.BlinkSpeed; got != DefaultBlink {
		t.Errorf("caret blinks every %v, want %v", got, DefaultBlink)
	}
}
