// Package wheel reads a mouse wheel turn as rows: the one rule every
// component that scrolls by the wheel shares.
package wheel

import (
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
)

// Unwrap is the turn a shell's WheelMsg carries, and any other message as it
// came: a component calls it first, so both reach the same case.
func Unwrap(msg tea.Msg) tea.Msg {
	if w, ok := msg.(bezel.WheelMsg); ok {
		return w.MouseWheelMsg
	}
	return msg
}

// Rows is -1 for a turn up and 1 for a turn down. A sideways turn, or one
// with shift held, pans instead, so it is false like any other message.
func Rows(msg tea.Msg) (int, bool) {
	w, ok := msg.(tea.MouseWheelMsg)
	if !ok || w.Mod.Contains(tea.ModShift) {
		return 0, false
	}
	switch w.Button {
	case tea.MouseWheelUp:
		return -1, true
	case tea.MouseWheelDown:
		return 1, true
	}
	return 0, false
}
