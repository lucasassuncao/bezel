package shell

import (
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/key"
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
		return s.Push(s.helpOverlay()), nil
	case kindCommands:
		return s.OpenCommands(), nil
	case kindChangeTab:
		return s.showTab(cycle(s.tab, len(s.cfg.Tabs), backward(ac, pressed)))
	case kindChangePane:
		from := s.Focused()
		return s.focused(from, s.nextPane(backward(ac, pressed)).relayout()) // so FocusMsg names where focus really lands
	}
	return s, a.Run(ActionContext{Context: s.Context(), Arg: arg})
}

// showTab moves to tab i, playing the Transition when there is one.
func (s Shell) showTab(i int) (Shell, tea.Cmd) {
	from := s.tab
	if s = s.SetTab(i); s.tab == from || s.cfg.Transition == nil {
		return s, nil
	}
	s.trans = s.cfg.Transition(from, s.tab)
	return s, s.trans.Tick()
}

// focused is s after focus moved from from: the glow and a FocusMsg when it
// landed somewhere else, nothing when it stayed.
func (s Shell) focused(from string, next Shell) (Shell, tea.Cmd) {
	to := next.Focused()
	if to == from {
		return next, nil
	}
	next.glow = newGlow(from, next.cfg.FocusFade)
	return next, tea.Batch(func() tea.Msg { return FocusMsg{From: from, To: to} }, next.glow.run.Tick())
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
