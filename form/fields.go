package form

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/icon"
	"github.com/lucasassuncao/bezel/theme"
)

// is reports whether msg is a key press any of bs answers to.
func is(msg tea.Msg, bs ...key.Binding) bool {
	k, ok := msg.(tea.KeyPressMsg)
	return ok && key.Matches(k, bs...)
}

// optionAt is options[i], or "" when there are none to choose from.
func optionAt(options []string, i int) string {
	if i < 0 || i >= len(options) {
		return ""
	}
	return options[i]
}

// Input is a one-line text field.
type Input struct {
	label string
	in    textinput.Model
	th    theme.Resolved
}

// NewInput is an empty field width cells wide, placeholder shown until typed in.
func NewInput(label, placeholder string, width int, th theme.Resolved) Input {
	in := draw.WithCaret(textinput.New(), th.CaretBlink)
	in.Prompt = ""
	in.Placeholder = placeholder
	in.SetWidth(width)
	return Input{label: label, in: in, th: th}
}

// WithValue is the field holding s, the way a form opens on saved values.
func (i Input) WithValue(s string) Input {
	i.in.SetValue(s)
	return i
}

// Value is the text in the field.
func (i Input) Value() string { return i.in.Value() }
func (i Input) Label() string { return i.label }

func (i Input) Update(msg tea.Msg) (Field, tea.Cmd) {
	var cmd tea.Cmd
	i.in, cmd = i.in.Update(msg)
	return i, cmd
}

func (i Input) Focus() (Field, tea.Cmd) {
	cmd := i.in.Focus()
	return i, cmd
}

func (i Input) Blur() Field {
	i.in.Blur()
	return i
}

func (i Input) View(focused bool, labelW, width int) string {
	edge := i.th.Muted
	if focused {
		edge = i.th.Accent
	}
	return row(i.th, focused, i.label, labelW, width, edge.Render("[")+i.in.View()+edge.Render("]"))
}

// Checkbox is a yes/no choice; space or enter flips it.
type Checkbox struct {
	label   string
	checked bool
	th      theme.Resolved
	keys    Keys
}

// NewCheckbox is a box labelled label, ticked when checked.
func NewCheckbox(label string, checked bool, th theme.Resolved) Checkbox {
	return Checkbox{label: label, checked: checked, th: th, keys: DefaultKeys()}
}

// WithKeys rebinds the box: Press flips it.
func (c Checkbox) WithKeys(k Keys) Checkbox {
	c.keys = k
	return c
}

func (c Checkbox) rebind(k Keys) Field { return c.WithKeys(k) }

// Checked reports whether the box is ticked.
func (c Checkbox) Checked() bool { return c.checked }

// Label is empty on purpose: the label is drawn after the box, in the value
// column, so it must not widen the label column.
func (c Checkbox) Label() string           { return "" }
func (c Checkbox) Focus() (Field, tea.Cmd) { return c, nil }
func (c Checkbox) Blur() Field             { return c }

func (c Checkbox) Update(msg tea.Msg) (Field, tea.Cmd) {
	if is(msg, c.keys.Press) {
		c.checked = !c.checked
	}
	return c, nil
}

// View draws the box before the label, the way a list of options reads; the
// label sits in the value column so boxes line up under the other controls.
func (c Checkbox) View(focused bool, labelW, width int) string {
	box := c.th.Muted.Render("[ ]")
	if c.checked {
		box = c.th.Accent.Render("[x]")
	}
	name := c.label
	if focused {
		name = c.th.Bold.Render(name)
	}
	return row(c.th, focused, "", labelW, width, box+" "+name)
}

// Radio is one choice out of a few, all shown on the row; left and right move it.
type Radio struct {
	label    string
	options  []string
	selected int
	th       theme.Resolved
	set      icon.Set
	keys     Keys
}

// NewRadio offers options with the one at selected chosen, clamped in range.
func NewRadio(label string, options []string, selected int, th theme.Resolved) Radio {
	selected = min(max(selected, 0), max(0, len(options)-1))
	return Radio{label: label, options: slices.Clone(options), selected: selected, th: th, set: icon.FromEnv(), keys: DefaultKeys()}
}

// WithKeys rebinds the radio: Left and Right move the choice.
func (r Radio) WithKeys(k Keys) Radio {
	r.keys = k
	return r
}

func (r Radio) rebind(k Keys) Field { return r.WithKeys(k) }

// WithIcons draws the chosen mark from set instead of icon.FromEnv's.
func (r Radio) WithIcons(set icon.Set) Radio {
	r.set = set
	return r
}

// Selected is the index of the chosen option.
func (r Radio) Selected() int { return r.selected }

// Value is the chosen option, or "" when there are none.
func (r Radio) Value() string           { return optionAt(r.options, r.selected) }
func (r Radio) Label() string           { return r.label }
func (r Radio) Focus() (Field, tea.Cmd) { return r, nil }
func (r Radio) Blur() Field             { return r }

func (r Radio) Update(msg tea.Msg) (Field, tea.Cmd) {
	switch {
	case is(msg, r.keys.Left):
		r.selected = max(0, r.selected-1)
	case is(msg, r.keys.Right):
		r.selected = max(0, min(len(r.options)-1, r.selected+1))
	}
	return r, nil
}

func (r Radio) View(focused bool, labelW, width int) string {
	parts := make([]string, len(r.options))
	for i, o := range r.options {
		if i == r.selected {
			parts[i] = r.th.Accent.Render("("+r.set.Bullet+")") + " " + o
			continue
		}
		parts[i] = r.th.Muted.Render("( )") + " " + r.th.Dim.Render(o)
	}
	return row(r.th, focused, r.label, labelW, width, strings.Join(parts, "   "))
}

// Toggle is a switch, on or off; space, enter, or left and right with h/l flip it.
type Toggle struct {
	label string
	on    bool
	th    theme.Resolved
	set   icon.Set
	keys  Keys
}

// NewToggle is a switch labelled label, on when on.
func NewToggle(label string, on bool, th theme.Resolved) Toggle {
	return Toggle{label: label, on: on, th: th, set: icon.FromEnv(), keys: DefaultKeys()}
}

// WithKeys rebinds the switch: Press, Left and Right flip it.
func (t Toggle) WithKeys(k Keys) Toggle {
	t.keys = k
	return t
}

func (t Toggle) rebind(k Keys) Field { return t.WithKeys(k) }

// WithIcons draws the switch from set instead of icon.FromEnv's.
func (t Toggle) WithIcons(set icon.Set) Toggle {
	t.set = set
	return t
}

// On reports whether the switch is on.
func (t Toggle) On() bool                { return t.on }
func (t Toggle) Label() string           { return t.label }
func (t Toggle) Focus() (Field, tea.Cmd) { return t, nil }
func (t Toggle) Blur() Field             { return t }

func (t Toggle) Update(msg tea.Msg) (Field, tea.Cmd) {
	if is(msg, t.keys.Press, t.keys.Left, t.keys.Right) {
		t.on = !t.on
	}
	return t, nil
}

func (t Toggle) View(focused bool, labelW, width int) string {
	sw := t.th.Muted.Render(t.set.Pending+t.set.Line+t.set.Line) + " " + t.th.Dim.Render("off")
	if t.on {
		sw = t.th.Success.Render(t.set.Line+t.set.Line+t.set.Running) + " " + t.th.Success.Render("on")
	}
	return row(t.th, focused, t.label, labelW, width, sw)
}

// Select is one value out of a list that opens under it: enter opens it, the
// arrows move, enter picks and esc closes without changing it.
type Select struct {
	label    string
	options  []string
	selected int
	cursor   int
	open     bool
	th       theme.Resolved
	set      icon.Set
	keys     Keys
}

// NewSelect offers options with the one at selected chosen, clamped in range.
func NewSelect(label string, options []string, selected int, th theme.Resolved) Select {
	selected = min(max(selected, 0), max(0, len(options)-1))
	return Select{
		label: label, options: slices.Clone(options), selected: selected, cursor: selected,
		th: th, set: icon.FromEnv(), keys: DefaultKeys(),
	}
}

// WithKeys rebinds the select: Press opens and picks, Up and Down move, Close closes.
func (s Select) WithKeys(k Keys) Select {
	s.keys = k
	return s
}

func (s Select) rebind(k Keys) Field { return s.WithKeys(k) }

// WithIcons draws the list's marks from set instead of icon.FromEnv's.
func (s Select) WithIcons(set icon.Set) Select {
	s.set = set
	return s
}

// Selected is the index of the chosen option.
func (s Select) Selected() int { return s.selected }

// Value is the chosen option, or "" when there are none.
func (s Select) Value() string { return optionAt(s.options, s.selected) }
func (s Select) Label() string { return s.label }

// Holding reports whether the list is open, keeping tab and esc from the form.
func (s Select) Holding() bool           { return s.open }
func (s Select) Focus() (Field, tea.Cmd) { return s, nil }

func (s Select) Blur() Field {
	s.open = false
	return s
}

func (s Select) Update(msg tea.Msg) (Field, tea.Cmd) {
	if !s.open {
		if is(msg, s.keys.Press) {
			s.open, s.cursor = true, s.selected
		}
		return s, nil
	}
	switch {
	case is(msg, s.keys.Up):
		s.cursor = max(0, s.cursor-1)
	case is(msg, s.keys.Down):
		s.cursor = max(0, min(len(s.options)-1, s.cursor+1))
	case is(msg, s.keys.Press):
		s.open = false
		if len(s.options) > 0 {
			s.selected = s.cursor
		}
	case is(msg, s.keys.Close):
		s.open = false
	}
	return s, nil
}

func (s Select) View(focused bool, labelW, width int) string {
	edge := s.th.Muted
	if focused {
		edge = s.th.Accent
	}
	closed := edge.Render("‹ ") + s.Value() + edge.Render(" ›")
	head := row(s.th, focused, s.label, labelW, width, closed)
	if !s.open {
		return head
	}
	indent := strings.Repeat(" ", 2+labelW+2*min(labelW, 1))
	lines := []string{head}
	for i, o := range s.options {
		if i == s.cursor {
			lines = append(lines, draw.Fit(indent+s.th.Accent.Render(s.set.Arrow+" ")+s.th.Bold.Render(o), width))
			continue
		}
		lines = append(lines, draw.Fit(indent+"  "+s.th.Dim.Render(o), width))
	}
	return strings.Join(lines, "\n")
}

// Button runs onPress on space or enter; a form's submit is one.
type Button struct {
	label   string
	onPress tea.Cmd
	th      theme.Resolved
	keys    Keys
}

// NewButton is a button labelled label; onPress may be nil.
func NewButton(label string, onPress tea.Cmd, th theme.Resolved) Button {
	return Button{label: label, onPress: onPress, th: th, keys: DefaultKeys()}
}

// WithKeys rebinds the button: Press presses it.
func (b Button) WithKeys(k Keys) Button {
	b.keys = k
	return b
}

func (b Button) rebind(k Keys) Field { return b.WithKeys(k) }

// Label is empty on purpose: a button sits in the value column and must not
// widen the label column.
func (b Button) Label() string           { return "" }
func (b Button) Focus() (Field, tea.Cmd) { return b, nil }
func (b Button) Blur() Field             { return b }

func (b Button) Update(msg tea.Msg) (Field, tea.Cmd) {
	if is(msg, b.keys.Press) {
		return b, b.onPress
	}
	return b, nil
}

func (b Button) View(focused bool, labelW, width int) string {
	return row(b.th, focused, "", labelW, width, draw.Button(b.label, b.th.Button, focused))
}
