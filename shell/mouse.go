package shell

import (
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
)

// ClickMsg is a left click on a pane, sent by the shell once it focused the
// pane: X and Y are cells of the pane's body, and may fall outside it on the
// border. It is bezel.ClickMsg, which the components take as it comes.
type ClickMsg = bezel.ClickMsg

// WheelMsg is a wheel turn over a pane, moved into the cells of its body. The
// focus stays where it was. It is bezel.WheelMsg, which the components take.
type WheelMsg = bezel.WheelMsg

// PaneAt is the pane drawn at the screen cell (x, y), border included, and
// false over the chrome, a gap or an open dialog, which covers the panes.
func (s Shell) PaneAt(x, y int) (string, bool) {
	s = s.ready()
	if s.overlays.Len() > 0 {
		return "", false
	}
	// Last drawn wins, as on screen: paneCanvas draws in declaration order.
	order := paneOrder(s.cfg.Layout)
	for i := len(order) - 1; i >= 0; i-- {
		r, ok := s.rects[order[i]]
		if ok && x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
			return order[i], true
		}
	}
	return "", false
}

// Mouse finds the pane under the pointer of msg and hands back msg moved
// into that pane's body, for a drag or a release; a click comes as a ClickMsg
// and a wheel turn as a WheelMsg. false when no pane is under the pointer.
func (s Shell) Mouse(msg tea.MouseMsg) (string, tea.MouseMsg, bool) {
	if msg == nil {
		return "", nil, false // no pointer, no pane
	}
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
	if msg == nil {
		return nil
	}
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

// mouse is the shell's part of the mouse: an edge it drags, a left click, a
// wheel turn over a pane. false leaves the message to the app as it came.
func (s Shell) mouse(msg tea.MouseMsg) (Shell, bool, tea.Cmd) {
	if w, ok := msg.(tea.MouseWheelMsg); ok {
		return s.wheel(w)
	}
	if next, cmd, ok := s.mouseResize(msg); ok {
		return next, true, cmd
	}
	if c, ok := msg.(tea.MouseClickMsg); ok && c.Button == tea.MouseLeft {
		return s.click(c)
	}
	return s, false, nil
}

// wheel hands a turn over a pane to the app as a WheelMsg. Over the chrome or
// an open dialog it is the app's as it came.
func (s Shell) wheel(w tea.MouseWheelMsg) (Shell, bool, tea.Cmd) {
	pane, ok := s.PaneAt(w.X, w.Y)
	if !ok {
		return s, false, nil
	}
	turn := s.MouseIn(pane, w).(tea.MouseWheelMsg)
	return s, true, bezel.Emit(WheelMsg{Pane: pane, MouseWheelMsg: turn})
}

// click is the shell's part of a left click: an open dialog's, a tab's, or
// a pane's, which it focuses and hands to the app as a ClickMsg. A click on
// the chrome is the app's as it came.
func (s Shell) click(c tea.MouseClickMsg) (Shell, bool, tea.Cmd) {
	if s.overlays.Len() > 0 {
		var r bezel.Result
		s.overlays, r = s.overlays.Click(c.X, c.Y, s.Body())
		return s, true, r.Cmd
	}
	if len(s.cfg.Tabs) > 0 && c.Y == s.headerRows() {
		if i := draw.TabAt(s.width, s.tabNames(), s.tab, s.cfg.Theme.Chrome, c.X); i >= 0 {
			return s.SetTab(i), true, nil
		}
		return s, true, nil
	}
	pane, ok := s.PaneAt(c.X, c.Y)
	if !ok {
		return s, false, nil
	}
	clicked := ClickMsg{Pane: pane, MouseClickMsg: s.MouseIn(pane, c).(tea.MouseClickMsg)}
	send := bezel.Emit(clicked)
	if pane == s.Focused() || !slices.Contains(focusRing(s.cfg.Layout), pane) {
		return s, true, send
	}
	next, focus := s.focused(s.Focused(), s.focusOn(pane).relayout())
	return next, true, tea.Batch(focus, send)
}
