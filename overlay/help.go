package overlay

import (
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// HelpSection is one group of keys under a heading.
type HelpSection struct {
	Name    string
	Entries []legend.Entry
}

// Help lists every key by section: a heading in the dialog's colour, then one
// key per row with what it does. ?, esc, q or enter closes it.
type Help struct {
	look
	title    string
	sections []HelpSection
	keys     HelpKeys
}

// HelpKeys are the bindings a Help closes on, besides the stack's esc.
type HelpKeys struct {
	Close key.Binding
}

// DefaultHelpKeys close on ?, esc, q or enter.
func DefaultHelpKeys() HelpKeys {
	return HelpKeys{Close: key.New("?", "esc", "q", "enter").WithHelp("esc", "close")}
}

// NewHelp lists sections, with no title over them.
func NewHelp(sections []HelpSection, th theme.Resolved) Help {
	return Help{look: lookOf(th), sections: slices.Clone(sections), keys: DefaultHelpKeys()}
}

// WithKeys rebinds the help; the line under the keys follows Close's help.
func (h Help) WithKeys(k HelpKeys) Help {
	h.keys = k
	return h
}

// WithTitle draws title over the sections.
func (h Help) WithTitle(title string) Help {
	h.title = title
	return h
}

func (h Help) Update(msg tea.Msg) (Overlay, bezel.Result) {
	if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, h.keys.Close) {
		return h, bezel.Result{Handled: true, Cmd: Close()}
	}
	_, isKey := msg.(tea.KeyPressMsg)
	return h, bezel.Result{Handled: isKey}
}

func (h Help) Legend() []legend.Entry { return legend.FromBindings([]key.Binding{h.keys.Close}) }
func (h Help) View(body layout.Rect) string {
	var lines []string
	if h.title != "" {
		lines = append(lines, accent(h.style()).Bold(true).Render(h.title), "")
	}
	for i, s := range h.sections {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, accent(h.style()).Bold(true).Render(s.Name), "")
		for _, e := range s.Entries {
			lines = append(lines, "  "+draw.Pad(e.Help().Key+" ", 16)+" "+e.Help().Desc)
		}
	}
	lines = append(lines, "", h.hint().Text.Render(h.keys.Close.Help().Key+" closes"))
	return Box(lines, body, dialogFrame(h.style()))
}
