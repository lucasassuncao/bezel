package shell

import (
	"maps"
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell/internal/palette"
)

// The cells Resize moves an edge per press: columns are narrow, so two.
const (
	stepCols = 2
	stepRows = 1
)

// Resize grows the focused pane on its keys, in order: narrower, wider,
// shorter, taller, moving the innermost Resizable box's edge around it. With
// the ":" palette open, shorter and taller resize the palette instead.
func Resize(opts ...Option) Action {
	keys := []string{"ctrl+shift+left", "ctrl+shift+right", "ctrl+shift+up", "ctrl+shift+down"}
	return build(&action{kind: kindResize, keys: keys, label: "ctrl+shift+arrows", desc: "resize"}, opts)
}

// ResetSizes gives every Resizable box the sizes its layout declares again,
// and the palette its Config.PaletteLines.
func ResetSizes(opts ...Option) Action {
	return build(&action{kind: kindResetSizes, keys: []string{"="}, label: "=", desc: "reset sizes", name: "reset-sizes"}, opts)
}

// ResizedMsg is sent when an edge between panes moved, by the mouse, Resize
// or ResetSizes: the app sizes its models from Rect again, as on a resize.
type ResizedMsg struct{}

func resizedCmd() tea.Msg { return ResizedMsg{} }

// drag is an edge held by the mouse: which one, and how far from its first
// cell the pointer took it.
type drag struct {
	edge layout.Edge
	off  int
}

// edges are the edges of the Resizable boxes on screen.
func (s Shell) edges() []layout.Edge {
	return s.sizes.Edges(s.cfg.Layout, s.Body(), s.presentFunc())
}

// edgeAt is the edge whose border cells hold (x, y): the last column or row
// of the child before it, or the first of the one after.
func (s Shell) edgeAt(x, y int) (drag, bool) {
	if s.overlays.Len() > 0 {
		return drag{}, false
	}
	for _, e := range s.edges() {
		along, across := y, x
		if !e.Cols {
			along, across = x, y
		}
		if along >= e.From && along < e.To && (across == e.At-1 || across == e.At) {
			return drag{edge: e, off: across - e.At}, true
		}
	}
	return drag{}, false
}

// dragTo moves the held edge so it follows the pointer at (x, y).
func (s Shell) dragTo(x, y int) Shell {
	for _, e := range s.edges() {
		if e.Box != s.held.edge.Box || e.Index != s.held.edge.Index {
			continue
		}
		at := x
		if !e.Cols {
			at = y
		}
		s.sizes = s.sizes.Move(s.cfg.Layout, s.Body(), s.presentFunc(), e, at-s.held.off-e.At)
		return s
	}
	return s
}

// mouseResize is the shell's part of a press, motion or release on an edge,
// with a ResizedMsg when the edge moved.
func (s Shell) mouseResize(msg tea.MouseMsg) (Shell, tea.Cmd, bool) {
	if next, ok := s.paletteMouse(msg); ok {
		return next, nil, true
	}
	m := msg.Mouse()
	switch msg.(type) {
	case tea.MouseClickMsg:
		if d, ok := s.edgeAt(m.X, m.Y); ok && m.Button == tea.MouseLeft {
			s.held, s.dragging = d, true
			return s, nil, true
		}
	case tea.MouseMotionMsg, tea.MouseReleaseMsg:
		if !s.dragging {
			break
		}
		_, released := msg.(tea.MouseReleaseMsg)
		s.dragging = !released
		next := s.dragTo(m.X, m.Y)
		return next, next.resizedFrom(s), true
	}
	return s, nil, false
}

// resizedFrom is a ResizedMsg when s's panes sit elsewhere than before's.
func (s Shell) resizedFrom(before Shell) tea.Cmd {
	if maps.Equal(s.placed(), before.placed()) {
		return nil
	}
	return resizedCmd
}

// placed is where the panes land with s's sizes.
func (s Shell) placed() map[string]layout.Rect {
	rects, _ := s.sizes.Resolve(s.cfg.Layout, s.Body(), s.presentFunc())
	return rects
}

// resized is s after a Resize key: which way pressed says, by its place
// among the action's keys.
func (s Shell) resized(a *action, pressed string) Shell {
	dir := 0
	for i, k := range a.keys {
		if k == pressed {
			dir = i
		}
	}
	cols, delta := dir < 2, stepCols
	if !cols {
		delta = stepRows
	}
	if dir%2 == 0 {
		delta = -delta
	}
	s.sizes = s.sizes.Grow(s.cfg.Layout, s.Body(), s.presentFunc(), s.Focused(), cols, delta)
	return s
}

// paletteKey is a Resize key's up or down on the open palette: shorter or
// taller, as for a pane. Left and right stay the typed line's.
func (s Shell) paletteKey(msg tea.Msg) (Shell, bool) {
	p, open := s.overlays.Top().(palette.Model)
	k, isKey := msg.(tea.KeyPressMsg)
	if !open || !isKey {
		return s, false
	}
	for _, a := range s.active() {
		ac, ok := a.(*action)
		if !ok || ac.kind != kindResize {
			continue
		}
		switch i := slices.Index(ac.keys, k.String()); i {
		case 2, 3:
			return s.setPaletteHeight(p, min(p.Height(), s.Body().H-1)+2*i-5), true
		}
	}
	return s, false
}

// setPaletteHeight makes n, kept to what the body holds, the open palette's
// height and the one it opens with again.
func (s Shell) setPaletteHeight(p palette.Model, n int) Shell {
	s.paletteLines = max(1, min(n, s.Body().H-1))
	s.overlays = s.overlays.Pop().Push(p.SetHeight(s.paletteLines))
	return s
}

// paletteMouse drags the open palette's rule: a press on it holds it, and
// the rule follows the pointer's row until the release.
func (s Shell) paletteMouse(msg tea.MouseMsg) (Shell, bool) {
	p, ok := s.overlays.Top().(palette.Model)
	if !ok {
		s.paletteDrag = false
		return s, false
	}
	body, m := s.Body(), msg.Mouse()
	lines := lipgloss.Height(p.View(body)) - 1 // under the rule
	rule := body.Y + body.H - lines - 1
	switch msg.(type) {
	case tea.MouseClickMsg:
		if m.Button != tea.MouseLeft || m.Y != rule || m.X < body.X || m.X >= body.X+body.W {
			return s, false
		}
		s.paletteDrag = true
		return s, true
	case tea.MouseMotionMsg, tea.MouseReleaseMsg:
		if !s.paletteDrag {
			return s, false
		}
		_, released := msg.(tea.MouseReleaseMsg)
		s = s.setPaletteHeight(p, lines+rule-m.Y)
		s.paletteDrag = !released
		return s, true
	}
	return s, false
}
