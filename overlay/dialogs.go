package overlay

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
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

// NewText shows text under title until any key.
func NewText(title, text string, th theme.Resolved) Text {
	return Text{title: title, lines: textLines(text), look: lookOf(th)}
}

// textLines is text split into lines, CRLF read as LF and trailing blank
// lines dropped; none for "".
func textLines(text string) []string {
	text = strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func (t Text) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return t, Close()
	}
	return t, nil
}

func (t Text) Legend() []legend.Entry { return []legend.Entry{legend.New("any key", "close")} }

func (t Text) View(body layout.Rect) string {
	return dialog(t.title, t.lines, t.okButton(t.lines), t.style(), body)
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
	return Option{Key: key.New(strings.Split(keys, "/")...).WithHelp(keys, label), Label: label, Do: do}
}

// Choice is a dialog with a body and keyed options: apply / dry-run / cancel.
// Picking one runs its command and closes the dialog. Esc closes it too.
type Choice struct {
	look
	title   string
	lines   []string
	options []Option
}

// NewChoice shows text under title with options as its buttons.
func NewChoice(title, text string, options []Option, th theme.Resolved) Choice {
	return Choice{title: title, lines: textLines(text), options: slices.Clone(options), look: lookOf(th)}
}

func (c Choice) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, nil
	}
	for _, o := range c.options {
		if key.Matches(k, o.Key) {
			return c, tea.Batch(o.Do, Close()) // Batch drops a nil Do
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
	return dialog(c.title, c.lines, strings.Join(c.buttons(), "  "), c.style(), body)
}

// buttons are the options as View draws them.
func (c Choice) buttons() []string {
	buttons := make([]string, 0, len(c.options))
	for _, o := range c.options {
		buttons = append(buttons, c.button(o.Label+" ("+o.Key.Help().Key+")", slices.Contains(o.Key.Keys(), "enter")))
	}
	return buttons
}

// Click runs the option whose button it lands on, as its key would.
func (c Choice) Click(x, y int, body layout.Rect) (Overlay, tea.Cmd) {
	i := buttonAt(c.View(body), "", dialogFrame(c.style()), c.buttons(), x, y)
	if i < 0 {
		return c, nil
	}
	return c, tea.Batch(c.options[i].Do, Close())
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
	keys   PickKeys
}

// PickKeys are the bindings a Pick answers to besides esc. Down has no help
// of its own: Up's "↑/↓" names both.
type PickKeys struct {
	Up, Down, Enter key.Binding
}

// DefaultPickKeys are the arrows and enter.
func DefaultPickKeys() PickKeys {
	return PickKeys{
		Up:    key.New("up").WithHelp("↑/↓", "choose"),
		Down:  key.New("down"),
		Enter: key.New("enter").WithHelp("enter", "pick"),
	}
}

// NewPick lists rows with the cursor on the first. onPick may be nil, which
// makes enter only close.
func NewPick(title string, rows []string, onPick func(i int) tea.Cmd, th theme.Resolved) Pick {
	return Pick{look: lookOf(th), title: title, rows: slices.Clone(rows), onPick: onPick, keys: DefaultPickKeys()}
}

// WithKeys rebinds the pick; the keys drawn in its box follow the bindings' help.
func (p Pick) WithKeys(k PickKeys) Pick {
	p.keys = k
	return p
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
	keys := legend.FromBindings([]key.Binding{p.keys.Up, p.keys.Down, p.keys.Enter})
	return append(keys, legend.New("esc", "cancel"))
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
		lines = append(lines, draw.More(below, 2, p.hint().Text))
	}
	if len(p.notes) > 0 {
		lines = append(lines, "")
		lines = append(lines, p.notes...)
	}
	// Its keys are drawn in the box: the legend under it stays the screen's.
	lines = append(lines, "", legend.HintLine(p.Legend(), p.hint()))
	return dialog(p.title, lines, "", p.style(), body)
}
