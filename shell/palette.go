package shell

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/palette"
)

// defaultStatusTTL is Config.StatusTTL when the app sets none.
const defaultStatusTTL = 5 * time.Second

// runMsg runs an action picked in the palette, after the palette has gone.
type runMsg struct {
	action Action
	arg    string
}

// paletteItems is every named action declared here, live or not: the palette
// lists the live ones and says "not available here" for the rest.
func (s Shell) paletteItems() []palette.Item {
	var out []palette.Item
	for _, a := range s.declared() {
		ac, ok := a.(*action)
		if !ok || ac.name == "" {
			continue
		}
		key := ac.hint
		if len(ac.keys) > 0 {
			key = ac.label
		}
		out = append(out, palette.Item{
			Name: ac.name, Arg: ac.arg, Key: key, Desc: ac.desc,
			Available: s.available(a),
			Run: func(arg string) tea.Cmd {
				return func() tea.Msg { return runMsg{action: a, arg: arg} }
			},
		})
	}
	return out
}

// OpenCommands opens the palette over the declared named actions, for an app
// that decides itself when ":" means commands and when it is text.
func (s Shell) OpenCommands() Shell {
	return s.Push(palette.New(s.paletteItems(), s.cfg.Theme)).relayout()
}
