package overlay

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Alert is a message box with an OK button; any key dismisses it.
type Alert struct {
	look
	title   string
	message string
}

// NewAlert is an alert of kind, its border in that kind's colour.
func NewAlert(kind Kind, title, message string, th theme.Resolved) Alert {
	return Alert{look: look{th: th, kind: kind}, title: title, message: message}
}

// Kind is what the alert reports.
func (a Alert) Kind() Kind { return a.kind }

func (a Alert) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return a, Close()
	}
	return a, nil
}

func (a Alert) Legend() []legend.Entry { return []legend.Entry{legend.New("any key", "close")} }

func (a Alert) View(body layout.Rect) string {
	lines := []string{a.message}
	return dialog(a.title, lines, okButton(lines, a.style()), a.style(), body)
}

// accent is the colour a dialog is drawn in: its border's.
func accent(style lipgloss.Style) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(style.GetBorderTopForeground())
}

// button is draw.Button in the dialog's colour.
func button(label string, ink lipgloss.Style, focused bool) string {
	return draw.Button(label, ink, focused)
}

// dialog is the look every overlay shares: a title in the dialog's colour,
// the content, then a row of buttons when there are any, with room around.
func dialog(title string, content []string, buttons string, style lipgloss.Style, body layout.Rect) string {
	lines := append([]string{accent(style).Bold(true).Render(title), ""}, content...)
	if buttons != "" {
		lines = append(lines, "", buttons)
	}
	return Box(lines, body, dialogFrame(style))
}

// dialogFrame is the box a dialog draws in: style, padded clear of its border.
func dialogFrame(style lipgloss.Style) lipgloss.Style { return style.Padding(1, 3) }

// okButton is a focused OK centred under content.
func okButton(content []string, style lipgloss.Style) string {
	w := lipgloss.Width(strings.Join(content, "\n"))
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, button("OK", accent(style), true))
}
