package shell

import "github.com/lucasassuncao/bezel/legend"

// declared is every action the current screen names, live or not: the
// global ones, the active tab's, then SetActions's.
func (s Shell) declared() []Action {
	out := append([]Action(nil), s.cfg.Actions...)
	if len(s.cfg.Tabs) > 0 {
		out = append(out, s.cfg.Tabs[s.tab].Actions(s.Context())...)
	}
	return append(out, s.actions...)
}

// available reports a live action: its capability allowed, its When true,
// and ChangeTab only with two tabs or more.
func (s Shell) available(a Action) bool {
	if ac, ok := a.(*action); ok {
		if ac.kind == kindChangeTab && len(s.cfg.Tabs) < 2 {
			return false
		}
		if ac.when != nil && !ac.when(s.Context()) {
			return false
		}
	}
	needs := a.Entry().Needs
	return needs == "" || s.cfg.Can == nil || s.cfg.Can(needs)
}

// active is the live declared actions, in the order they were declared: the
// legend prints them that way, so the app decides where help and quit sit.
func (s Shell) active() []Action {
	var out []Action
	for _, a := range s.declared() {
		if s.available(a) {
			out = append(out, a)
		}
	}
	return out
}

// actionEntries is the legend half of the active actions. Palette-only
// commands have no key and stay out; Help only when withHelp.
func (s Shell) actionEntries(withHelp bool) []legend.Entry {
	var out []legend.Entry
	for _, a := range s.active() {
		if e := a.Entry(); len(e.Keys()) > 0 && (withHelp || kindOf(a) != kindHelp) {
			out = append(out, e)
		}
	}
	return out
}

// helpAction is the active Help, or nil.
func (s Shell) helpAction() Action {
	for _, a := range s.active() {
		if kindOf(a) == kindHelp {
			return a
		}
	}
	return nil
}
