package tree

import (
	"slices"
	"time"

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
// is busy: call it after New, SetNodes or SetNode. A second one lands spent.
func (m Model[N]) Tick() tea.Cmd { return m.spin.Tick() }

// Active reports whether the Busy spinner turns.
func (m Model[N]) Active() bool { return m.spin.Active() }

// FrameAt is the spinner's frame for at, for a test that delivers frames itself.
func (m Model[N]) FrameAt(at time.Time) tea.Msg { return m.spin.FrameAt(at) }

// Advance is the tree at now, for a host that keeps its own clock; it
// settles only once no node is busy.
func (m Model[N]) Advance(now time.Time) (Model[N], bool) {
	var done bool
	m.spin, done = m.spin.Advance(now)
	return m, done
}

// BusyMark is the spinner's frame in the Busy style, padded to its widest.
// The row func draws it in place of a Busy node's icon: the rows, icons
// included, are the app's to draw.
func (m Model[N]) BusyMark(th theme.Resolved) string {
	st := m.styles(th)
	return m.spin.WithStyles(func(spinner.Styles) spinner.Styles {
		return spinner.Styles{Frame: st.Busy}
	}).View(th)
}

// spun turns the spinner while a node is busy and holds it when none is.
func (m Model[N]) spun() Model[N] {
	m.spin = m.spin.SetRunning(slices.ContainsFunc(m.nodes, func(n N) bool { return n.TreeInfo().Busy }))
	return m
}
