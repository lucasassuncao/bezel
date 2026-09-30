package shell

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// runKey runs the first live action bound to k. Display-only actions never
// match, so the key reaches the focused component.
func (s Shell) runKey(k tea.KeyPressMsg) (Shell, tea.Cmd, bool) {
	for _, a := range s.active() {
		if ac, ok := a.(*action); ok && ac.displayOnly() {
			continue
		}
		if key.Matches(k, a.Entry().Binding) {
			next, cmd := s.run(a, "", k.String())
			return next, cmd, true
		}
	}
	return s, nil, false
}

// run performs a: the shell's own work for the prebuilt kinds, a.Run for the
// rest. pressed is the key that fired it, "" from the palette.
func (s Shell) run(a Action, arg, pressed string) (Shell, tea.Cmd) {
	ac, _ := a.(*action)
	if ac != nil && ac.override {
		return s, a.Run(ActionContext{Context: s.Context(), Arg: arg})
	}
	switch kindOf(a) {
	case kindHelp:
		return s.Push(s.helpOverlay()).relayout(), nil
	case kindCommands:
		return s.OpenCommands(), nil
	case kindChangeTab:
		return s.SetTab(cycle(s.tab, len(s.cfg.Tabs), backward(ac, pressed))), nil
	case kindChangePane:
		step := 1
		if backward(ac, pressed) {
			step = -1
		}
		from := s.Focus()
		s = s.nextPane(step).relayout()
		if to := s.Focus(); to != from {
			return s, func() tea.Msg { return FocusMsg{From: from, To: to} }
		}
		return s, nil
	}
	return s, a.Run(ActionContext{Context: s.Context(), Arg: arg})
}

// backward reports a cycle walking back: fired by any key after the first.
func backward(a *action, pressed string) bool {
	return a != nil && pressed != "" && len(a.keys) > 1 && pressed != a.keys[0]
}

func cycle(i, n int, back bool) int {
	if n == 0 {
		return i
	}
	if back {
		return (i - 1 + n) % n
	}
	return (i + 1) % n
}
