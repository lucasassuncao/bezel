package draw

import (
	"testing"
	"time"
)

// A host shrinks the blink in its tests; an input that ignored it would make
// every keystroke wait out the 530ms default.
func TestBlinkUsesTheBlinkGiven(t *testing.T) {
	if got := Blink(time.Millisecond); got != time.Millisecond {
		t.Errorf("caret blinks every %v, want %v", got, time.Millisecond)
	}
}

func TestBlinkDefaultsAZeroBlink(t *testing.T) {
	if got := Blink(0); got != DefaultBlink {
		t.Errorf("caret blinks every %v, want %v", got, DefaultBlink)
	}
}
