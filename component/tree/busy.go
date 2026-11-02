package tree

import (
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/component/spinner"
	"github.com/lucasassuncao/bezel/theme"
)

// WithSpinner turns s for the Busy nodes instead of spinner.MiniDot.
func (m Model[N]) WithSpinner(s spinner.Spinner) Model[N] {
	m.spin = spinner.New(s)
	return m.spun()
}

// Tick is the command for the Busy spinner's next frame, nil while no node
// is busy: call it after New, SetNodes or SetNode.
func (m Model[N]) Tick() tea.Cmd { return m.spin.Tick() }

// BusyMark is the spinner's frame in the Busy style, padded to its widest:
// a row func draws it in place of a Busy node's icon.
func (m Model[N]) BusyMark(th theme.Resolved) string {
	st := m.styles(th)
	return m.spin.WithStyles(func(spinner.Styles) spinner.Styles {
		return spinner.Styles{Frame: st.Busy}
	}).View(th)
}

// spun turns the spinner while a node is busy and stops it when none is. It
// restarts on every change, so the Tick asked for after it is the only one
// turning it.
func (m Model[N]) spun() Model[N] {
	if len(m.spin.Spinner().Frames) == 0 { // a zero tree has no spinner yet
		m.spin = spinner.New(spinner.MiniDot())
	}
	m.spin = m.spin.Stop()
	for _, n := range m.nodes {
		if n.TreeInfo().Busy {
			m.spin, _ = m.spin.Start()
			break
		}
	}
	return m
}
