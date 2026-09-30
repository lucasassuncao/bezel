package theme

import (
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
)

// Resolved is a Theme after the whole cascade has been applied: ready-to-use
// lipgloss styles plus the colors they were derived from. A consumer builds one
// per instance and only reads it afterwards.
type Resolved struct {
	Colors Colors

	// Text styles for content that says what kind of thing it is.
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Danger      lipgloss.Style
	Info        lipgloss.Style
	Accent      lipgloss.Style
	Dim         lipgloss.Style // secondary text
	Muted       lipgloss.Style // hints and gutters, quieter than Dim
	Key         lipgloss.Style // a key name in running text
	Bold        lipgloss.Style
	Section     lipgloss.Style // a heading inside a list
	TableHeader lipgloss.Style
	Cursor      lipgloss.Style // the row under the cursor
	Badge       lipgloss.Style // a filled label; recolour its background per kind

	// Chrome styles the shell draws with. Derived from Colors like the rest.
	Panel        draw.PanelStyle
	PanelFocused draw.PanelStyle
	Chrome       draw.ChromeStyle
	Legend       legend.Style
	StatusOK     lipgloss.Style
	StatusErr    lipgloss.Style
	Modal        lipgloss.Style
}

// Resolve turns t into styles for a dark or light terminal. Steps: the
// colour cascade (ResolveColors), the derived styles, then t.Styles overrides.
func Resolve(t Theme, dark bool) Resolved {
	c := ResolveColors(t, dark)

	rt := buildDerivedStyles(c)
	rt.Colors = c

	if t.Styles.Danger != nil {
		rt.Danger = *t.Styles.Danger
	}
	if t.Styles.Muted != nil {
		rt.Muted = *t.Styles.Muted
	}
	if t.Styles.Cursor != nil {
		rt.Cursor = *t.Styles.Cursor
	}
	return rt
}

// ModalFor is the modal box coloured by what it reports: the accent for
// information, the palette's success/warning/danger inks otherwise.
func (r Resolved) ModalFor(kind overlay.Kind) lipgloss.Style {
	switch kind {
	case overlay.Success:
		return r.Modal.BorderForeground(lipgloss.Color(r.Colors.Success))
	case overlay.Warning:
		return r.Modal.BorderForeground(lipgloss.Color(r.Colors.Warning))
	case overlay.Danger:
		return r.Modal.BorderForeground(lipgloss.Color(r.Colors.Danger))
	}
	return r.Modal
}

// buildDerivedStyles creates the lipgloss styles from the resolved palette.
func buildDerivedStyles(c Colors) Resolved {
	accent := lipgloss.Color(c.Accent)
	selection := lipgloss.Color(c.Selection)
	border := lipgloss.Color(c.Border)
	dim := lipgloss.Color(c.Dim)
	success := lipgloss.Color(c.Success)
	warning := lipgloss.Color(c.Warning)
	danger := lipgloss.Color(c.Danger)
	info := lipgloss.Color(c.Info)
	onAccent := lipgloss.Color(c.OnAccent)

	r := Resolved{
		Success:     lipgloss.NewStyle().Foreground(success),
		Warning:     lipgloss.NewStyle().Foreground(warning),
		Danger:      lipgloss.NewStyle().Foreground(danger),
		Info:        lipgloss.NewStyle().Foreground(info),
		Accent:      lipgloss.NewStyle().Foreground(accent),
		Dim:         lipgloss.NewStyle().Foreground(dim),
		Muted:       lipgloss.NewStyle().Foreground(border),
		Key:         lipgloss.NewStyle().Bold(true).Foreground(selection),
		Bold:        lipgloss.NewStyle().Bold(true),
		Section:     lipgloss.NewStyle().Bold(true).Foreground(accent).PaddingLeft(1),
		TableHeader: lipgloss.NewStyle().Bold(true).Foreground(border),
		Cursor:      lipgloss.NewStyle().Bold(true).Foreground(selection),
		Badge:       lipgloss.NewStyle().Bold(true).Foreground(onAccent).Background(accent),
	}
	r.Panel = draw.PanelStyle{
		Border: lipgloss.NewStyle().Foreground(border),
		Title:  lipgloss.NewStyle().Foreground(dim),
	}
	// Focus rides on the thick edge and the filled title as well as the colour,
	// so it survives a monochrome terminal.
	r.PanelFocused = draw.PanelStyle{
		Border: lipgloss.NewStyle().Foreground(accent),
		Title:  lipgloss.NewStyle().Bold(true).Foreground(onAccent).Background(selection),
		Edge:   lipgloss.ThickBorder(),
	}
	r.Chrome = draw.ChromeStyle{
		Title:     lipgloss.NewStyle().Bold(true).Foreground(selection).PaddingLeft(1),
		Info:      lipgloss.NewStyle().Foreground(dim),
		Tab:       lipgloss.NewStyle().Foreground(border),
		TabActive: lipgloss.NewStyle().Bold(true).Foreground(selection).Underline(true),
		Status:    lipgloss.NewStyle().Foreground(border),
		Dim:       lipgloss.NewStyle().Foreground(border),
	}
	r.Legend = legend.Style{
		Key:  lipgloss.NewStyle().Bold(true).Foreground(accent),
		Text: lipgloss.NewStyle().Foreground(border),
	}
	r.StatusOK = lipgloss.NewStyle().Foreground(success)
	r.StatusErr = lipgloss.NewStyle().Bold(true).Foreground(danger)
	r.Modal = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(0, 1)
	return r
}
