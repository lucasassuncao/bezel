package overlay

import (
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Kind is what a dialog reports; it picks the colour of its border and title.
type Kind int

const (
	Info Kind = iota
	Success
	Warning
	Danger
)

// look is what every dialog draws with: the theme, and the Kind tinting it.
// The theme is held by pointer, shared by every copy and never written, so
// copying a dialog costs a word, not the 29KB a theme.Resolved takes.
type look struct {
	th   *theme.Resolved
	kind Kind
}

// lookOf is a look at th, untinted.
func lookOf(th theme.Resolved) look { return look{th: &th} }

// noTheme is what a zero look draws with; it is never written.
var noTheme theme.Resolved

// styles is the look's theme, or the zero one for a dialog never built.
func (l look) styles() *theme.Resolved {
	if l.th == nil {
		return &noTheme
	}
	return l.th
}

// style is the theme's modal box, its border in the kind's colour: the accent
// for information, the palette's success, warning or danger otherwise.
func (l look) style() lipgloss.Style {
	c := l.styles().Colors
	var border string
	switch l.kind {
	case Success:
		border = c.Success
	case Warning:
		border = c.Warning
	case Danger:
		border = c.Danger
	default:
		return l.styles().Modal
	}
	return l.styles().Modal.BorderForeground(lipgloss.Color(border))
}

// hint is how a dialog draws the keys it lists inside its box.
func (l look) hint() legend.Style { return l.styles().Legend }

// WithKind tints the confirm, such as Danger for one that deletes.
func (c Confirm) WithKind(k Kind) Confirm { c.kind = k; return c }

// WithKind tints the text for what it reports.
func (t Text) WithKind(k Kind) Text { t.kind = k; return t }

// WithKind tints the choice for what it decides.
func (c Choice) WithKind(k Kind) Choice { c.kind = k; return c }

// WithKind tints the pick for what it decides.
func (p Pick) WithKind(k Kind) Pick { p.kind = k; return p }

// WithKind tints the help.
func (h Help) WithKind(k Kind) Help { h.kind = k; return h }

// WithKind tints the pager for what its text reports, such as a failed run.
func (p Pager) WithKind(k Kind) Pager { p.kind = k; return p }

// WithKind tints the prompt, such as Danger for a typed confirmation.
func (p Prompt) WithKind(k Kind) Prompt { p.kind = k; return p }

// WithKind tints the picker, its entries in the new colour.
func (p FilePicker) WithKind(k Kind) FilePicker { p.kind = k; return p }
