package overlay

import (
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
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

// Styles are what every dialog but Alert paints with. A Kind other than Info
// still recolours the border, the title and the focused button: those
// colours report something and stay the theme's.
type Styles struct {
	Box    lipgloss.Style   // the modal box
	Hint   legend.Style     // the keys a dialog lists inside its box
	Button draw.ButtonStyle // the focused one is filled in the border's colour
}

// DefaultStyles are the theme's modal box, legend and buttons.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Box: th.Modal, Hint: th.Legend, Button: th.Button}
}

// look is what every dialog draws with: the theme, its styles and the Kind
// tinting them. Both are held by pointer, shared by every copy and never
// written, so copying a dialog costs a word, not the 29KB a theme takes.
type look struct {
	th   *theme.Resolved
	st   *Styles
	edit func(Styles) Styles
	kind Kind
}

// lookOf is a look at th, untinted.
func lookOf(th theme.Resolved) look { return look{th: &th}.restyle() }

// restyle derives the styles from the theme and applies the app's edit.
func (l look) restyle() look {
	st := DefaultStyles(*l.styles())
	if l.edit != nil {
		st = l.edit(st)
	}
	l.st = &st
	return l
}

func (l look) withTheme(th theme.Resolved) look {
	l.th = &th
	return l.restyle()
}

func (l look) withStyles(edit func(Styles) Styles) look {
	l.edit = edit
	return l.restyle()
}

// own is the look's styles, or the zero ones for a dialog never built.
func (l look) own() *Styles {
	if l.st == nil {
		return &noStyles
	}
	return l.st
}

// noTheme and noStyles are what a zero look draws with; never written.
var (
	noTheme  theme.Resolved
	noStyles Styles
)

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
		return l.own().Box
	}
	return l.own().Box.BorderForeground(lipgloss.Color(border))
}

// hint is how a dialog draws the keys it lists inside its box.
func (l look) hint() legend.Style { return l.own().Hint }

// WithStyles changes the styles derived from the theme through edit; it
// replaces an earlier edit.
func (c Confirm) WithStyles(edit func(Styles) Styles) Confirm { c.look = c.withStyles(edit); return c }

// WithStyles changes the styles derived from the theme through edit.
func (t Text) WithStyles(edit func(Styles) Styles) Text { t.look = t.withStyles(edit); return t }

// WithStyles changes the styles derived from the theme through edit.
func (c Choice) WithStyles(edit func(Styles) Styles) Choice { c.look = c.withStyles(edit); return c }

// WithStyles changes the styles derived from the theme through edit.
func (p Pick) WithStyles(edit func(Styles) Styles) Pick { p.look = p.withStyles(edit); return p }

// WithStyles changes the styles derived from the theme through edit.
func (h Help) WithStyles(edit func(Styles) Styles) Help { h.look = h.withStyles(edit); return h }

// WithStyles changes the styles derived from the theme through edit.
func (p Pager) WithStyles(edit func(Styles) Styles) Pager { p.look = p.withStyles(edit); return p }

// WithStyles changes the styles derived from the theme through edit.
func (p Prompt) WithStyles(edit func(Styles) Styles) Prompt { p.look = p.withStyles(edit); return p }

// WithStyles changes the styles derived from the theme through edit.
func (p FilePicker) WithStyles(edit func(Styles) Styles) FilePicker {
	p.look = p.withStyles(edit)
	return p
}

// SetTheme repaints the alert under th; its Kind keeps its colour.
func (a Alert) SetTheme(th theme.Resolved) Overlay { a.look = a.withTheme(th); return a }

// SetTheme repaints the confirm under th.
func (c Confirm) SetTheme(th theme.Resolved) Overlay { c.look = c.withTheme(th); return c }

// SetTheme repaints the text under th.
func (t Text) SetTheme(th theme.Resolved) Overlay { t.look = t.withTheme(th); return t }

// SetTheme repaints the choice under th.
func (c Choice) SetTheme(th theme.Resolved) Overlay { c.look = c.withTheme(th); return c }

// SetTheme repaints the pick under th.
func (p Pick) SetTheme(th theme.Resolved) Overlay { p.look = p.withTheme(th); return p }

// SetTheme repaints the help under th.
func (h Help) SetTheme(th theme.Resolved) Overlay { h.look = h.withTheme(th); return h }

// SetTheme repaints the pager under th.
func (p Pager) SetTheme(th theme.Resolved) Overlay { p.look = p.withTheme(th); return p }

// SetTheme repaints the file picker under th.
func (p FilePicker) SetTheme(th theme.Resolved) Overlay { p.look = p.withTheme(th); return p }

// SetTheme repaints the prompt and its field under th.
func (p Prompt) SetTheme(th theme.Resolved) Overlay {
	p.look, p.input = p.withTheme(th), p.input.SetTheme(th)
	return p
}

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
