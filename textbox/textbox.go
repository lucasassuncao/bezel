// Package textbox is the bubbles text area set up for a multi-line value: no
// length or line cap, the numbered gutter NumberLines draws, and the theme's
// colours. Height stays the caller's: a panel sizes it, a form grows it.
package textbox

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// New is an empty text area themed with th. The widget's defaults cap a value
// at 400 characters and 99 lines, and drop the rest without a word.
func New(th theme.Resolved) textarea.Model {
	ta := textarea.New()
	ta.CharLimit = 0
	ta.MaxHeight = 0
	ta.ShowLineNumbers = false
	ta.EndOfBufferCharacter = '~'
	return Restyle(ta, th)
}

// Restyle recolours ta, for when the terminal reports its background late.
// The line under the cursor is left unpainted: highlighted, it reads as a
// selection. Blurred text is dimmed so the focused widget is the bright one.
func Restyle(ta textarea.Model, th theme.Resolved) textarea.Model {
	ta.SetPromptFunc(draw.GutterWidth, func(info textarea.PromptInfo) string {
		return draw.Gutter(info.LineNumber+1, th.Muted)
	})
	plain := lipgloss.NewStyle()
	s := ta.Styles()
	for _, st := range []*textarea.StyleState{&s.Focused, &s.Blurred} {
		st.Base, st.CursorLine, st.Text = plain, plain, plain
		st.EndOfBuffer, st.Placeholder = th.Muted, th.Muted
	}
	s.Blurred.Text = th.Dim
	ta.SetStyles(s)
	return ta
}

// SetText loads s. The widget breaks a line on \r and on \n alike, so a
// CRLF value would come back with every line doubled.
func SetText(ta *textarea.Model, s string) {
	ta.SetValue(foldCRLF(s))
}

// Update is ta.Update with a paste's CRLF folded first, for the same reason.
func Update(ta textarea.Model, msg tea.Msg) (textarea.Model, tea.Cmd) {
	if p, ok := msg.(tea.PasteMsg); ok {
		p.Content = foldCRLF(p.Content)
		msg = p
	}
	return ta.Update(msg)
}

func foldCRLF(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
