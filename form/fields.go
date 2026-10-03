package form

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/icon"
	"github.com/lucasassuncao/bezel/theme"
)

// pressed reports whether msg is space or enter, the keys that act on a control.
func pressed(msg tea.Msg) bool {
	k, ok := msg.(tea.KeyPressMsg)
	return ok && (k.String() == "space" || k.String() == "enter")
}

// keyOf is msg's key name, or "" for anything that is not a key press.
func keyOf(msg tea.Msg) string {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		return k.String()
	}
	return ""
}

// Input is a one-line text field.
type Input struct {
	label string
	in    textinput.Model
	th    theme.Resolved
}

// NewInput is an empty field width cells wide, placeholder shown until typed in.
func NewInput(label, placeholder string, width int, th theme.Resolved) Input {
	in := textinput.New()
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
}

func NewCheckbox(label string, checked bool, th theme.Resolved) Checkbox {
	return Checkbox{label: label, checked: checked, th: th}
}

func (c Checkbox) Checked() bool           { return c.checked }
func (c Checkbox) Label() string           { return "" }
func (c Checkbox) Focus() (Field, tea.Cmd) { return c, nil }
func (c Checkbox) Blur() Field             { return c }

func (c Checkbox) Update(msg tea.Msg) (Field, tea.Cmd) {
	if pressed(msg) {
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
}

func NewRadio(label string, options []string, selected int, th theme.Resolved) Radio {
	return Radio{label: label, options: options, selected: min(max(selected, 0), max(0, len(options)-1)), th: th, set: icon.Default()}
}

func (r Radio) Selected() int           { return r.selected }
func (r Radio) Value() string           { return r.options[r.selected] }
func (r Radio) Label() string           { return r.label }
func (r Radio) Focus() (Field, tea.Cmd) { return r, nil }
func (r Radio) Blur() Field             { return r }

func (r Radio) Update(msg tea.Msg) (Field, tea.Cmd) {
	switch keyOf(msg) {
	case "left", "h":
		r.selected = max(0, r.selected-1)
	case "right", "l":
		r.selected = min(len(r.options)-1, r.selected+1)
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

// Toggle is a switch, on or off; space, enter, left or right flips it.
type Toggle struct {
	label string
	on    bool
	th    theme.Resolved
	set   icon.Set
}

func NewToggle(label string, on bool, th theme.Resolved) Toggle {
	return Toggle{label: label, on: on, th: th, set: icon.Default()}
}

func (t Toggle) On() bool                { return t.on }
func (t Toggle) Label() string           { return t.label }
func (t Toggle) Focus() (Field, tea.Cmd) { return t, nil }
func (t Toggle) Blur() Field             { return t }

func (t Toggle) Update(msg tea.Msg) (Field, tea.Cmd) {
	if pressed(msg) || keyOf(msg) == "left" || keyOf(msg) == "right" {
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
}

func NewSelect(label string, options []string, selected int, th theme.Resolved) Select {
	selected = min(max(selected, 0), max(0, len(options)-1))
	return Select{label: label, options: options, selected: selected, cursor: selected, th: th, set: icon.Default()}
}

func (s Select) Selected() int           { return s.selected }
func (s Select) Value() string           { return s.options[s.selected] }
func (s Select) Label() string           { return s.label }
func (s Select) Holding() bool           { return s.open }
func (s Select) Focus() (Field, tea.Cmd) { return s, nil }

func (s Select) Blur() Field {
	s.open = false
	return s
}

func (s Select) Update(msg tea.Msg) (Field, tea.Cmd) {
	k := keyOf(msg)
	if !s.open {
		if k == "enter" || k == "space" {
			s.open, s.cursor = true, s.selected
		}
		return s, nil
	}
	switch k {
	case "up", "k":
		s.cursor = max(0, s.cursor-1)
	case "down", "j":
		s.cursor = min(len(s.options)-1, s.cursor+1)
	case "enter", "space":
		s.selected, s.open = s.cursor, false
	case "esc":
		s.open = false
	}
	return s, nil
}

func (s Select) View(focused bool, labelW, width int) string {
	edge := s.th.Muted
	if focused {
		edge = s.th.Accent
	}
	closed := edge.Render("‹ ") + s.options[s.selected] + edge.Render(" ›")
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
}

func NewButton(label string, onPress tea.Cmd, th theme.Resolved) Button {
	return Button{label: label, onPress: onPress, th: th}
}

func (b Button) Label() string           { return "" }
func (b Button) Focus() (Field, tea.Cmd) { return b, nil }
func (b Button) Blur() Field             { return b }

func (b Button) Update(msg tea.Msg) (Field, tea.Cmd) {
	if pressed(msg) {
		return b, b.onPress
	}
	return b, nil
}

func (b Button) View(focused bool, labelW, width int) string {
	return row(b.th, focused, "", labelW, width, draw.Button(b.label, b.th.Accent, focused))
}
