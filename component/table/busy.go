package table

import (
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/component/spinner"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// WithSpinner turns s on the Busy rows instead of spinner.MiniDot.
func (m Model[T]) WithSpinner(s spinner.Spinner) Model[T] {
	m.spin = spinner.New(s)
	return m.spun()
}

// Tick is the command for the Busy spinner's next frame, nil while no row is
// busy: call it after New or SetItems. A second one lands spent.
func (m Model[T]) Tick() tea.Cmd { return m.spin.Tick() }

// Active reports whether the Busy spinner turns.
func (m Model[T]) Active() bool { return m.spin.Active() }

// FrameAt is the spinner's frame for at, for a test that delivers frames itself.
func (m Model[T]) FrameAt(at time.Time) tea.Msg { return m.spin.FrameAt(at) }

// Advance is the table at now, for a host that keeps its own clock; it
// settles only once no row is busy.
func (m Model[T]) Advance(now time.Time) (Model[T], bool) {
	var done bool
	m.spin, done = m.spin.Advance(now)
	return m, done
}

// spun turns the spinner while a row is busy and holds it when none is.
func (m Model[T]) spun() Model[T] {
	m.spin = m.spin.SetRunning(slices.ContainsFunc(m.items, func(it Item[T]) bool { return it.Busy }))
	return m
}

// busyMark is the spinner's frame in the Busy style, in the marker's width.
func (m Model[T]) busyMark(st Styles) string {
	f := m.spin.WithStyles(func(spinner.Styles) spinner.Styles {
		return spinner.Styles{Frame: st.Busy}
	}).View(theme.Resolved{}) // the edit replaces every style the theme would give
	return draw.Pad(f, draw.Width(marker))
}
