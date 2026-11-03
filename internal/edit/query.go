package edit

import (
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// Typed is the one character k types at a / prompt, "" for a key that types
// none or several.
func Typed(k tea.KeyPressMsg) string {
	if t := k.Key().Text; utf8.RuneCountInString(t) == 1 {
		return t
	}
	return ""
}

// Erase is text without its last character, "" left as it is.
func Erase(text string) string {
	_, size := utf8.DecodeLastRuneInString(text)
	return text[:len(text)-size]
}
