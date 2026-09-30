// Package bezeltest builds what a terminal would send, for tests that drive a
// Bubble Tea model. It depends on nothing else in bezel.
package bezeltest

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

var named = map[string]rune{
	"enter": tea.KeyEnter, "esc": tea.KeyEscape, "tab": tea.KeyTab,
	"space": tea.KeySpace, "backspace": tea.KeyBackspace, "delete": tea.KeyDelete,
	"insert": tea.KeyInsert, "up": tea.KeyUp, "down": tea.KeyDown,
	"left": tea.KeyLeft, "right": tea.KeyRight, "home": tea.KeyHome,
	"end": tea.KeyEnd, "pgup": tea.KeyPgUp, "pgdown": tea.KeyPgDown,
	"f1": tea.KeyF1, "f2": tea.KeyF2, "f3": tea.KeyF3, "f4": tea.KeyF4,
	"f5": tea.KeyF5, "f6": tea.KeyF6, "f7": tea.KeyF7, "f8": tea.KeyF8,
	"f9": tea.KeyF9, "f10": tea.KeyF10, "f11": tea.KeyF11, "f12": tea.KeyF12,
}

var modifiers = map[string]tea.KeyMod{
	"ctrl": tea.ModCtrl, "alt": tea.ModAlt, "shift": tea.ModShift,
	"meta": tea.ModMeta, "hyper": tea.ModHyper, "super": tea.ModSuper,
}

// Key is the key press a terminal sends for s, in tea's own notation:
// "enter", "ctrl+s", "shift+tab", "a", "?". Unknown names panic, so a
// typo in a test fails instead of sending some other key.
func Key(s string) tea.KeyPressMsg {
	mods, name := split(s)
	var msg tea.KeyPressMsg
	for _, m := range mods {
		mod, ok := modifiers[m]
		if !ok {
			panic(fmt.Sprintf("bezeltest.Key(%q): unknown modifier %q", s, m))
		}
		msg.Mod |= mod
	}
	if code, ok := named[name]; ok {
		msg.Code = code
	} else if utf8.RuneCountInString(name) == 1 {
		msg.Code, _ = utf8.DecodeRuneInString(name)
	} else {
		panic(fmt.Sprintf("bezeltest.Key(%q): unknown key %q", s, name))
	}
	// tea reads Text back as the key's name, so only a bare character has it.
	if msg.Mod == 0 && (msg.Code == tea.KeySpace || !isNamed(name)) {
		msg.Text = string(msg.Code)
	}
	return msg
}

// split cuts s into its modifiers and the key, where "+" alone or at the
// end ("ctrl++") is the plus key, not a separator.
func split(s string) (mods []string, name string) {
	switch {
	case s == "+":
		return nil, "+"
	case strings.HasSuffix(s, "++"):
		return strings.Split(s[:len(s)-2], "+"), "+"
	}
	i := strings.LastIndex(s, "+")
	if i < 0 {
		return nil, s
	}
	return strings.Split(s[:i], "+"), s[i+1:]
}

func isNamed(name string) bool {
	_, ok := named[name]
	return ok
}
