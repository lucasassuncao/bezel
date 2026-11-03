package list

import (
	"slices"

	"github.com/lucasassuncao/bezel/component/spinner"
	"github.com/lucasassuncao/bezel/theme"
)

// WithSpinner turns s on the Busy rows instead of spinner.MiniDot.
func (m Model[T]) WithSpinner(s spinner.Spinner) Model[T] {
	m.spin = spinner.New(s)
	return m.spun()
}

// spun turns the spinner while a row is busy and holds it when none is.
func (m Model[T]) spun() Model[T] {
	m.spin = m.spin.SetRunning(slices.ContainsFunc(m.rows, func(r Row[T]) bool { return r.Busy && !r.Section }))
	return m
}

// busyMark is the spinner's frame in the Busy style, padded to its widest.
func (m Model[T]) busyMark(st *Styles) string {
	return m.spin.WithStyles(func(spinner.Styles) spinner.Styles {
		return spinner.Styles{Frame: st.Busy}
	}).View(theme.Resolved{}) // the edit replaces every style the theme would give
}
