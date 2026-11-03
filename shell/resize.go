package shell

import (
	"maps"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
)

// The cells Resize moves an edge per press: columns are narrow, so two.
const (
	stepCols = 2
	stepRows = 1
)

// Resize grows the focused pane on its keys, in order: narrower, wider,
// shorter, taller. It moves the edges of a layout.Box marked Resizable, the
// innermost of that axis around the pane; a pane in none stays as it is.
func Resize(opts ...Option) Action {
	keys := []string{"ctrl+shift+left", "ctrl+shift+right", "ctrl+shift+up", "ctrl+shift+down"}
	return build(&action{kind: kindResize, keys: keys, label: "ctrl+shift+arrows", desc: "resize"}, opts)
}

// ResetSizes gives every Resizable box the sizes its layout declares again.
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
