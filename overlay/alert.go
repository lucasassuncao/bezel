package overlay

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Alert is a message box with an OK button; any key dismisses it.
// Non-themable: its Kind sets its colours, so it has no WithStyles.
type Alert struct {
	look
	title   string
	message string
}

// NewAlert is an alert of kind, its border in that kind's colour.
func NewAlert(kind Kind, title, message string, th theme.Resolved) Alert {
	l := lookOf(th)
	l.kind = kind
	return Alert{look: l, title: title, message: message}
}

// Kind is what the alert reports.
func (a Alert) Kind() Kind { return a.kind }

func (a Alert) Update(msg tea.Msg) (Overlay, bezel.Result) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return a, bezel.Result{Handled: true, Cmd: Close()}
	}
	return a, bezel.Result{}
}

func (a Alert) Legend() []legend.Entry { return []legend.Entry{legend.New("any key", "close")} }

func (a Alert) View(body layout.Rect) string {
	lines := []string{a.message}
	return dialog(a.title, lines, a.okButton(lines), a.style(), body)
}

// Click closes the alert on a click on OK.
func (a Alert) Click(x, y int, body layout.Rect) (Overlay, bezel.Result) {
	row := a.okButton([]string{a.message})
	if buttonAt(a.View(body), row, dialogFrame(a.style()), []string{a.button("OK", true)}, x, y) == 0 {
		return a, bezel.Result{Handled: true, Cmd: Close()}
	}
	return a, bezel.Result{Handled: true}
}

// accent is the colour a dialog is drawn in: its border's.
func accent(style lipgloss.Style) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(style.GetBorderTopForeground())
}

// button is the theme's button with the focused one filled in the dialog's
// colour: its border's, which the dialog's kind tints.
func (l look) button(label string, focused bool) string {
	st := l.own().Button
	st.Focused = st.Focused.Background(l.style().GetBorderTopForeground())
	return draw.Button(label, st, focused)
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
func (l look) okButton(content []string) string {
	w := lipgloss.Width(strings.Join(content, "\n"))
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, l.button("OK", true))
}
