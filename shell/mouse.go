package shell

import (
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
)

// ClickMsg is a left click on a pane, sent by the shell once it focused the
// pane: X and Y are cells of the pane's body, the rect its Body draws in,
// and may fall outside it on the border. The embedded message is the click
// moved there, for a component that takes the mouse itself.
type ClickMsg struct {
	Pane string
	tea.MouseClickMsg
}

// Screen is the shell's view as the full-screen tea.View a program returns,
// with mouse reporting on when Config.Mouse is.
func (s Shell) Screen(panes map[string]Pane) tea.View {
	s = s.ready()
	v := tea.NewView(s.View(panes))
	v.AltScreen = true
	if s.cfg.Mouse {
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}

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
// into that pane's body, for the wheel or a drag; a click comes as a
// ClickMsg instead. false when no pane is under the pointer.
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
// a pane's, which it focuses and hands to the app as a ClickMsg. A click on
// the chrome is the app's as it came.
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
	if !ok {
		return s, false, nil
	}
	clicked := ClickMsg{Pane: pane, MouseClickMsg: s.MouseIn(pane, c).(tea.MouseClickMsg)}
	send := func() tea.Msg { return clicked }
	if pane == s.Focus() || !slices.Contains(layout.Ring(s.cfg.Layout), pane) {
		return s, true, send
	}
	next, focus := s.focused(s.Focus(), s.focusOn(pane).relayout())
	return next, true, tea.Batch(focus, send)
}
