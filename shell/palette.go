package shell

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/shell/internal/palette"
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
// that decides itself when ":" means commands and when it is text. Its caret
// blinks from the first key on; the Commands action starts it at once.
func (s Shell) OpenCommands() Shell {
	s = s.ready()
	return s.Push(palette.New(s.paletteItems(), s.cfg.Theme).SetHeight(s.paletteHeight())).relayout()
}

// initTop is the command that starts the overlay just pushed, nil when it
// has none: Push takes no command, so whoever pushed runs it.
func (s Shell) initTop() tea.Cmd {
	if in, ok := s.overlays.Top().(overlay.Initer); ok {
		return in.Init()
	}
	return nil
}

// paletteHeight is the candidates the palette lists: the user's choice, else
// the Config's, else the palette's own default.
func (s Shell) paletteHeight() int {
	switch {
	case s.paletteLines > 0:
		return s.paletteLines
	case s.cfg.PaletteLines > 0:
		return s.cfg.PaletteLines
	}
	return palette.DefaultHeight
}

// CommandLine is what is typed on the ":" command line, and whether it is
// open on top of the overlays.
func (s Shell) CommandLine() (string, bool) {
	p, ok := s.TopOverlay().(palette.Model)
	if !ok {
		return "", false
	}
	return p.Value(), true
}
