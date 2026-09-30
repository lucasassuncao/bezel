package overlay

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

// HelpSection is one group of keys under a heading.
type HelpSection struct {
	Name    string
	Entries []legend.Entry
}

// Help lists every key by section: a heading in the dialog's colour, then one
// key per row with what it does. ?, esc, q or enter closes it.
type Help struct {
	title    string
	sections []HelpSection
	style    lipgloss.Style
	hint     legend.Style
}

func NewHelp(title string, sections []HelpSection, style lipgloss.Style, hint legend.Style) Help {
	return Help{title: title, sections: sections, style: style, hint: hint}
}

func (h Help) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && (k.String() == "?" || k.String() == "esc" || k.String() == "q" || k.String() == "enter") {
		return h, Close()
	}
	return h, nil
}

func (h Help) Legend() []legend.Entry { return []legend.Entry{legend.New("esc", "close")} }
func (h Help) View(body layout.Rect) string {
	var lines []string
	if h.title != "" {
		lines = append(lines, accent(h.style).Bold(true).Render(h.title), "")
	}
	for i, s := range h.sections {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, accent(h.style).Bold(true).Render(s.Name), "")
		for _, e := range s.Entries {
			lines = append(lines, "  "+draw.Pad(e.Help().Key+" ", 16)+" "+e.Help().Desc)
		}
	}
	lines = append(lines, "", h.hint.Text.Render("esc closes"))
	return Box(lines, body, h.style.Padding(1, 3))
}
