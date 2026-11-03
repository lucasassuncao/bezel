package edit

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestTypedIsOneCharacterOrNone(t *testing.T) {
	cases := map[string]tea.KeyPressMsg{
		"q": {Code: 'q', Text: "q"},
		"é": {Code: 'é', Text: "é"},
		"":  {Code: tea.KeyEnter},
	}
	for want, k := range cases {
		if got := Typed(k); got != want {
			t.Errorf("Typed(%v) = %q, want %q", k, got, want)
		}
	}
	if got := Typed(tea.KeyPressMsg{Text: "ab"}); got != "" {
		t.Errorf("a paste of two types %q, want none", got)
	}
}

func TestEraseDropsTheLastCharacter(t *testing.T) {
	for in, want := range map[string]string{"café": "caf", "a": "", "": ""} {
		if got := Erase(in); got != want {
			t.Errorf("Erase(%q) = %q, want %q", in, got, want)
		}
	}
}
