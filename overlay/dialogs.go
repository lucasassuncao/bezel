package overlay

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Text is a titled block of lines any key dismisses: a help screen, a notice.
type Text struct {
	look
	title string
	lines []string
}

// NewText shows lines under title until any key.
func NewText(title string, lines []string, th theme.Resolved) Text {
	return Text{title: title, lines: slices.Clone(lines), look: look{th: th}}
}

func (t Text) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return t, Close()
	}
	return t, nil
}

func (t Text) Legend() []legend.Entry { return []legend.Entry{legend.New("any key", "close")} }

func (t Text) View(body layout.Rect) string {
	return dialog(t.title, t.lines, okButton(t.lines, t.style()), t.style(), body)
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
	look
	title   string
	lines   []string
	options []Option
}

// NewChoice shows lines under title with options as its buttons.
func NewChoice(title string, lines []string, options []Option, th theme.Resolved) Choice {
	return Choice{title: title, lines: slices.Clone(lines), options: slices.Clone(options), look: look{th: th}}
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
	ink := accent(c.style())
	buttons := make([]string, 0, len(c.options))
	for _, o := range c.options {
		buttons = append(buttons, button(o.Label+" ("+o.Key.Help().Key+")", ink, slices.Contains(o.Key.Keys(), "enter")))
	}
	return dialog(c.title, c.lines, strings.Join(buttons, "  "), c.style(), body)
}

// Pick is a dialog that chooses one row of a list with the arrows: which
// source manages this, which preset to apply. Enter runs OnPick with the
// cursor; esc closes with nothing. A list taller than the body scrolls.
type Pick struct {
	look
	title  string
	rows   []string
	notes  []string      // drawn under the rows
	win    layout.Window // its Height is what fit at the last Resize
	onPick func(i int) tea.Cmd
	keys   Keys
}

// Keys are the bindings a Pick answers to besides esc.
type Keys struct {
	Up, Down, Enter key.Binding
}

// NewPick lists rows with the cursor on the first. onPick may be nil, which
// makes enter only close.
func NewPick(title string, rows []string, onPick func(i int) tea.Cmd, th theme.Resolved) Pick {
	return Pick{look: look{th: th}, title: title, rows: slices.Clone(rows), onPick: onPick, keys: Keys{
		Up:    key.NewBinding(key.WithKeys("up")),
		Down:  key.NewBinding(key.WithKeys("down")),
		Enter: key.NewBinding(key.WithKeys("enter")),
	}}
}

// WithNotes are lines drawn under the rows, such as what the picked one does.
func (p Pick) WithNotes(notes ...string) Pick {
	p.notes = slices.Clone(notes)
	return p
}

// WithCursor starts the cursor on row i, such as the value already chosen;
// an index outside the rows is clamped to them.
func (p Pick) WithCursor(i int) Pick {
	p.win.Cursor = i
	p.win = p.win.Clamp(len(p.rows))
	return p
}

// Cursor is the index of the row under the cursor.
func (p Pick) Cursor() int { return p.win.Cursor }

// Resize shows as many rows as fit in body, scrolled to keep the cursor's.
func (p Pick) Resize(body layout.Rect) Overlay {
	p.win.Height = p.fit(body)
	p.win = p.win.Clamp(len(p.rows))
	return p
}

// fit is how many rows body leaves after the frame, the title, the notes and
// the keys, at least one.
func (p Pick) fit(body layout.Rect) int {
	chrome := 4 // the title and the keys, each with its gap
	if len(p.notes) > 0 {
		chrome += 1 + strings.Count(strings.Join(p.notes, "\n"), "\n") + 1
	}
	return max(1, body.H-dialogFrame(p.style()).GetVerticalFrameSize()-chrome)
}

func (p Pick) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	switch {
	case key.Matches(k, p.keys.Up):
		p.win = p.win.Move(-1, len(p.rows), nil)
	case key.Matches(k, p.keys.Down):
		p.win = p.win.Move(1, len(p.rows), nil)
	case key.Matches(k, p.keys.Enter):
		if p.onPick == nil || len(p.rows) == 0 {
			return p, Close()
		}
		return p, tea.Batch(p.onPick(p.win.Cursor), Close())
	}
	return p, nil
}

func (p Pick) Legend() []legend.Entry {
	return []legend.Entry{legend.New("↑/↓", "choose"), legend.New("enter", "pick"), legend.New("esc", "cancel")}
}

func (p Pick) View(body layout.Rect) string {
	// Drawn for the body it is given, which a Resize may not have seen yet.
	w := p.win
	w.Height = p.fit(body)
	w = w.Clamp(len(p.rows))
	start, end, below := w.Bounds(len(p.rows))
	var lines []string
	for i := start; i < end; i++ {
		if i == w.Cursor {
			lines = append(lines, accent(p.style()).Bold(true).Render("▸ "+p.rows[i]))
		} else {
			lines = append(lines, "  "+p.rows[i])
		}
	}
	if below > 0 {
		lines = append(lines, p.hint().Text.Render(fmt.Sprintf("  ↓ %d more", below)))
	}
	if len(p.notes) > 0 {
		lines = append(lines, "")
		lines = append(lines, p.notes...)
	}
	// Its keys are drawn in the box: the legend under it stays the screen's.
	lines = append(lines, "", legend.HintLine(p.Legend(), p.hint()))
	return dialog(p.title, lines, "", p.style(), body)
}
