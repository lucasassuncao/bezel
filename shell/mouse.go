package shell

import (
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
)

// PaneAt is the pane drawn at the screen cell (x, y), border included, and
// false over the chrome, a gap or an open dialog, which covers the panes.
func (s Shell) PaneAt(x, y int) (string, bool) {
	s = s.ready()
	if s.overlays.Len() > 0 {
		return "", false
	}
	// Last drawn wins, as on screen: paneCanvas draws in declaration order.
	order := layout.Order(s.cfg.Layout)
	for i := len(order) - 1; i >= 0; i-- {
		r, ok := s.rects[order[i]]
		if ok && x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
			return order[i], true
		}
	}
	return "", false
}

// Mouse finds the pane under the pointer of msg and hands back msg moved
// into that pane's body, the rect its Body draws in, so a component in it
// reads the cell it was given. false when no pane is under the pointer.
func (s Shell) Mouse(msg tea.MouseMsg) (string, tea.MouseMsg, bool) {
	at := msg.Mouse()
	pane, ok := s.PaneAt(at.X, at.Y)
	if !ok {
		return "", msg, false
	}
	return pane, s.MouseIn(pane, msg), true
}

// MouseIn is msg moved into pane's body wherever the pointer is, for a drag
// that began in the pane and left it: the cell may then fall outside.
func (s Shell) MouseIn(pane string, msg tea.MouseMsg) tea.MouseMsg {
	inner := draw.InnerRect(s.Rect(pane))
	m := msg.Mouse()
	m.X, m.Y = m.X-inner.X, m.Y-inner.Y
	switch msg.(type) {
	case tea.MouseClickMsg:
		return tea.MouseClickMsg(m)
	case tea.MouseReleaseMsg:
		return tea.MouseReleaseMsg(m)
	case tea.MouseWheelMsg:
		return tea.MouseWheelMsg(m)
	case tea.MouseMotionMsg:
		return tea.MouseMotionMsg(m)
	}
	return msg
}

// click is the shell's part of a left click: an open dialog's, a tab's, or
// the focus of a pane out of focus, after which the click comes back to be
// the app's. A click on a focused pane or the chrome is the app's at once.
func (s Shell) click(c tea.MouseClickMsg) (Shell, bool, tea.Cmd) {
	if s.overlays.Len() > 0 {
		var cmd tea.Cmd
		s.overlays, cmd = s.overlays.Click(c.X, c.Y, s.Body())
		return s, true, cmd
	}
	if len(s.cfg.Tabs) > 0 && c.Y == s.headerRows() {
		if i := draw.TabAt(s.width, s.tabNames(), s.tab, s.cfg.Theme.Chrome, c.X); i >= 0 {
			next, cmd := s.showTab(i)
			return next, true, cmd
		}
		return s, true, nil
	}
	pane, ok := s.PaneAt(c.X, c.Y)
	if !ok || pane == s.Focus() || !slices.Contains(layout.Ring(s.cfg.Layout), pane) {
		return s, false, nil
	}
	next, cmd := s.focused(s.Focus(), s.focusOn(pane).relayout())
	again := func() tea.Msg { return c } // now on the focused pane: the app's
	return next, true, tea.Sequence(cmd, again)
}
