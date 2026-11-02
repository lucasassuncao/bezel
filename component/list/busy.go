package list

import (
	"github.com/lucasassuncao/bezel/component/spinner"
	"github.com/lucasassuncao/bezel/theme"
)

// WithSpinner turns s on the Busy rows instead of spinner.MiniDot.
func (m Model[T]) WithSpinner(s spinner.Spinner) Model[T] {
	m.spin = spinner.New(s)
	return m.spun()
}

// spun turns the spinner while a row is busy and stops it when none is. It
// restarts on every change, so the Tick asked for after SetRows is the only
// one turning it.
func (m Model[T]) spun() Model[T] {
	m.spin = m.spin.Stop()
	for _, r := range m.rows {
		if r.Busy && !r.Section {
			m.spin, _ = m.spin.Start()
			break
		}
	}
	return m
}

// busyMark is the spinner's frame in the Busy style, padded to its widest.
func (m Model[T]) busyMark(st *Styles) string {
	return m.spin.WithStyles(func(spinner.Styles) spinner.Styles {
		return spinner.Styles{Frame: st.Busy}
	}).View(theme.Resolved{}) // the edit replaces every style the theme would give
}
