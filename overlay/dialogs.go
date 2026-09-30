package overlay

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

// Text is a titled block of lines any key dismisses: a help screen, a notice.
type Text struct {
	title string
	lines []string
	style lipgloss.Style
	hint  legend.Style
}

func NewText(title string, lines []string, style lipgloss.Style, hint legend.Style) Text {
	return Text{title: title, lines: lines, style: style, hint: hint}
}

func (t Text) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return t, Close()
	}
	return t, nil
}

func (t Text) Legend() []legend.Entry { return []legend.Entry{legend.New("any key", "close")} }

func (t Text) View(body layout.Rect) string {
	return dialog(t.title, t.lines, okButton(t.lines, t.style), t.style, body)
}

// Option is one thing a Choice offers: the key that picks it, what it says
// in the hint, and what it does. A nil Do only closes the dialog.
type Option struct {
	Key   key.Binding
	Label string
	Do    tea.Cmd
}

// Opt builds an Option from a key label and what it does.
func Opt(keys, label string, do tea.Cmd) Option {
	return Option{Key: key.NewBinding(key.WithKeys(strings.Split(keys, "/")...), key.WithHelp(keys, label)), Label: label, Do: do}
}

// Choice is a dialog with a body and keyed options: apply / dry-run / cancel.
// Picking one runs its command and closes the dialog. Esc closes it too.
type Choice struct {
	title   string
	lines   []string
	options []Option
	style   lipgloss.Style
	hint    legend.Style
}

func NewChoice(title string, lines []string, options []Option, style lipgloss.Style, hint legend.Style) Choice {
	return Choice{title: title, lines: lines, options: options, style: style, hint: hint}
}

func (c Choice) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, nil
	}
	for _, o := range c.options {
		if key.Matches(k, o.Key) {
			if o.Do == nil {
				return c, Close()
			}
			return c, tea.Batch(o.Do, Close())
		}
	}
	return c, nil
}

// Legend is the options, plus esc unless an option already names it: esc
// always closes, the stack sees to that.
func (c Choice) Legend() []legend.Entry {
	out := make([]legend.Entry, 0, len(c.options)+1)
	hasEsc := false
	for _, o := range c.options {
		out = append(out, legend.Entry{Binding: o.Key})
		for _, k := range o.Key.Keys() {
			hasEsc = hasEsc || k == "esc"
		}
	}
	if !hasEsc {
		out = append(out, legend.New("esc", "cancel"))
	}
	return out
}

// View draws each option as a button naming its keys. The one enter picks is
// filled in the dialog's colour, since that is what enter will do.
func (c Choice) View(body layout.Rect) string {
	ink := accent(c.style)
	buttons := make([]string, 0, len(c.options))
	for _, o := range c.options {
		buttons = append(buttons, button(o.Label+" ("+o.Key.Help().Key+")", ink, slices.Contains(o.Key.Keys(), "enter")))
	}
	return dialog(c.title, c.lines, strings.Join(buttons, "  "), c.style, body)
}

// Pick is a dialog that chooses one row of a list with the arrows: which
// source manages this, which preset to apply. Enter runs OnPick with the
// cursor; esc closes with nothing.
type Pick struct {
	title  string
	rows   []string
	notes  []string // drawn under the rows
	cursor int
	onPick func(i int) tea.Cmd
	style  lipgloss.Style
	hint   legend.Style
	keys   Keys
}

// Keys are the bindings a Pick answers to besides esc.
type Keys struct {
	Up, Down, Enter key.Binding
}

func NewPick(title string, rows, notes []string, onPick func(i int) tea.Cmd, style lipgloss.Style, hint legend.Style) Pick {
	return Pick{title: title, rows: rows, notes: notes, onPick: onPick, style: style, hint: hint, keys: Keys{
		Up:    key.NewBinding(key.WithKeys("up")),
		Down:  key.NewBinding(key.WithKeys("down")),
		Enter: key.NewBinding(key.WithKeys("enter")),
	}}
}

func (p Pick) Cursor() int { return p.cursor }

func (p Pick) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	switch {
	case key.Matches(k, p.keys.Up):
		p.cursor = max(0, p.cursor-1)
	case key.Matches(k, p.keys.Down):
		p.cursor = max(0, min(len(p.rows)-1, p.cursor+1))
	case key.Matches(k, p.keys.Enter):
		if p.onPick == nil || len(p.rows) == 0 {
			return p, Close()
		}
		return p, tea.Batch(p.onPick(p.cursor), Close())
	}
	return p, nil
}

func (p Pick) Legend() []legend.Entry {
	return []legend.Entry{legend.New("↑/↓", "choose"), legend.New("enter", "pick"), legend.New("esc", "cancel")}
}

func (p Pick) View(body layout.Rect) string {
	var lines []string
	for i, r := range p.rows {
		if i == p.cursor {
			lines = append(lines, accent(p.style).Bold(true).Render("▸ "+r))
		} else {
			lines = append(lines, "  "+r)
		}
	}
	if len(p.notes) > 0 {
		lines = append(lines, "")
		lines = append(lines, p.notes...)
	}
	// Its keys are drawn in the box: the legend under it stays the screen's.
	lines = append(lines, "", legend.HintLine(p.Legend(), p.hint))
	return dialog(p.title, lines, "", p.style, body)
}
