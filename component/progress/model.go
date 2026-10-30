package progress

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// Styles are how a Model paints its bar and its percentage.
type Styles struct {
	Full    lipgloss.Style // the filled cells, unless a gradient colours them
	Empty   lipgloss.Style // the track
	Percent lipgloss.Style
}

// DefaultStyles fill the bar in the accent over a muted track, the
// percentage dim.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Full: th.Accent, Empty: th.Muted, Percent: th.Dim}
}

var _ bezel.Model[Model] = Model{}

// Model is a bar that moves to the share it is set to on a spring instead of
// jumping there. A value: every method returns the new bar.
type Model struct {
	spring      animation.Spring // Pos and Target are shares in [0, 1]
	width       int
	full, empty string
	gradient    []color.Color
	percent     bool
	format      string
	edit        func(Styles) Styles
}

// New is an empty bar width cells wide, the percentage included, with
// █ over ░ and the percentage shown.
func New(width int) Model {
	return Model{spring: animation.NewSpring(0), width: max(width, 1), full: "█", empty: "░", percent: true, format: " %3.0f%%"}
}

// WithChars draws the filled cells with full and the track with empty.
func (m Model) WithChars(full, empty rune) Model {
	m.full, m.empty = string(full), string(empty)
	return m
}

// WithGradient colours the filled cells along colors, blended from the
// bar's left end to its right; fewer than two is the Full style again.
func (m Model) WithGradient(colors ...color.Color) Model {
	m.gradient = nil
	if len(colors) > 1 {
		m.gradient = append([]color.Color(nil), colors...)
	}
	return m
}

// WithPercentage shows or hides the percentage after the bar.
func (m Model) WithPercentage(on bool) Model {
	m.percent = on
	return m
}

// WithFormat writes the percentage with format, a fmt verb for a float64
// from 0 to 100, " %3.0f%%" by default.
func (m Model) WithFormat(format string) Model {
	m.format = format
	return m
}

// WithSpring sets how stiff the motion is, in radians a second, and how
// damped: 1 arrives without overshoot.
func (m Model) WithSpring(freq, damping float64) Model {
	m.spring = m.spring.WithFreq(freq).WithDamping(damping)
	return m
}

// WithStyles changes the styles derived from the theme through edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

// SetWidth changes the bar's width, the percentage included.
func (m Model) SetWidth(w int) Model {
	m.width = max(w, 1)
	return m
}

// Width is the bar's width, the percentage included.
func (m Model) Width() int { return m.width }

// SetPercent aims the bar at p, kept within [0, 1]; the command moves it.
func (m Model) SetPercent(p float64) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spring, cmd = m.spring.To(min(max(p, 0), 1))
	return m, cmd
}

// Percent is the share the bar is aimed at.
func (m Model) Percent() float64 { return m.spring.Target }

// Shown is the share the bar draws now, on its way to Percent.
func (m Model) Shown() float64 { return m.spring.Pos }

// Active reports whether the bar is still moving.
func (m Model) Active() bool { return m.spring.Active() }

// Update moves the bar on its own frames until it settles.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	var r bezel.Result
	m.spring, r = m.spring.Update(msg)
	return m, r
}

// View is the bar where it is now, then the percentage it is aimed at.
func (m Model) View(th theme.Resolved) string { return m.ViewAs(m.spring.Pos, th) }

// ViewAs is the bar at share p with no motion, for a host that drives the
// share itself.
func (m Model) ViewAs(p float64, th theme.Resolved) string {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	p = min(max(p, 0), 1)
	pct := ""
	if m.percent {
		pct = fmt.Sprintf(m.format, p*100)
	}
	barW := max(m.width-ansi.StringWidth(pct), 0)
	filled := int(p*float64(barW) + 0.5)
	var b strings.Builder
	for i := range filled {
		fill := st.Full
		if m.gradient != nil {
			fill = fill.Foreground(m.colorAt(i, barW))
		}
		b.WriteString(fill.Render(m.full))
	}
	b.WriteString(st.Empty.Render(strings.Repeat(m.empty, barW-filled)))
	if pct != "" {
		b.WriteString(st.Percent.Render(pct))
	}
	return draw.Fit(b.String(), m.width)
}

// colorAt is the gradient's colour at cell i of a bar n cells wide.
func (m Model) colorAt(i, n int) color.Color {
	if n <= 1 {
		return m.gradient[0]
	}
	at := float64(i) / float64(n-1) * float64(len(m.gradient)-1)
	k := min(int(at), len(m.gradient)-2)
	return animation.Blend(m.gradient[k], m.gradient[k+1], at-float64(k))
}
