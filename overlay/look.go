package overlay

import (
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Styles are what every dialog but Alert paints with. A dialog given a Kind
// still draws its border, title and focused button in that Kind's fixed
// colour, whatever the theme or WithStyles says.
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
	kind draw.Kind
	// tinted is a dialog given a Kind: Info's zero value alone would paint
	// every plain dialog blue.
	tinted bool
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

// style is the theme's modal box, its border in the Kind's fixed colour when
// the dialog was given one.
func (l look) style() lipgloss.Style {
	if !l.tinted {
		return l.own().Box
	}
	return l.own().Box.BorderForeground(l.kind.Color())
}

// withKind is l tinted with k.
func (l look) withKind(k draw.Kind) look {
	l.kind, l.tinted = k, true
	return l
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
func (c Confirm) WithKind(k draw.Kind) Confirm { c.look = c.withKind(k); return c }

// WithKind tints the text for what it reports.
func (t Text) WithKind(k draw.Kind) Text { t.look = t.withKind(k); return t }

// WithKind tints the choice for what it decides.
func (c Choice) WithKind(k draw.Kind) Choice { c.look = c.withKind(k); return c }

// WithKind tints the pick for what it decides.
func (p Pick) WithKind(k draw.Kind) Pick { p.look = p.withKind(k); return p }

// WithKind tints the help.
func (h Help) WithKind(k draw.Kind) Help { h.look = h.withKind(k); return h }

// WithKind tints the pager for what its text reports, such as a failed run.
func (p Pager) WithKind(k draw.Kind) Pager { p.look = p.withKind(k); return p }

// WithKind tints the prompt, such as Danger for a typed confirmation.
func (p Prompt) WithKind(k draw.Kind) Prompt { p.look = p.withKind(k); return p }

// WithKind tints the picker, its entries in the new colour.
func (p FilePicker) WithKind(k draw.Kind) FilePicker { p.look = p.withKind(k); return p }
