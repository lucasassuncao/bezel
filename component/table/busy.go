package table

import (
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
// busy: call it after New or SetItems.
func (m Model[T]) Tick() tea.Cmd { return m.spin.Tick() }

// spun turns the spinner while a row is busy and stops it when none is. It
// restarts on every change, so the Tick asked for after SetItems is the only
// one turning it.
func (m Model[T]) spun() Model[T] {
	m.spin = m.spin.Stop()
	for _, it := range m.items {
		if it.Busy {
			m.spin, _ = m.spin.Start()
			break
		}
	}
	return m
}

// busyMark is the spinner's frame in the Busy style, in the marker's width.
func (m Model[T]) busyMark(st Styles) string {
	f := m.spin.WithStyles(func(spinner.Styles) spinner.Styles {
		return spinner.Styles{Frame: st.Busy}
	}).View(theme.Resolved{}) // the edit replaces every style the theme would give
	return draw.Pad(f, draw.Width(marker))
}
