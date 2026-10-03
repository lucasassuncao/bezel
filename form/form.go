// Package form is the controls that ask for a value: a text input, a select,
// checkboxes, radios, a toggle and a button, stacked in a Form that moves focus
// with tab and shift+tab. Each control is one row, its label in a shared column.
package form

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// Field is one control. It gets keys only while focused, and draws its label
// padded to labelW so a form's values start in one column.
type Field interface {
	Update(msg tea.Msg) (Field, tea.Cmd)
	View(focused bool, labelW, width int) string
	Focus() (Field, tea.Cmd)
	Blur() Field
	Label() string
}

// Holder is a Field that wants the keys a Form would otherwise take, such as an
// open Select keeping tab and esc until it closes.
type Holder interface{ Holding() bool }

// Form is fields in a column. tab and shift+tab move the focus; every other key
// goes to the field in focus.
type Form struct {
	fields []Field
	focus  int
}

// New is a form focused on its first field.
func New(fields ...Field) (Form, tea.Cmd) {
	f := Form{fields: fields}
	if len(fields) == 0 {
		return f, nil
	}
	var cmd tea.Cmd
	f.fields[0], cmd = f.fields[0].Focus()
	return f, cmd
}

// Field is the i-th field as it is now; assert it to read its value.
func (f Form) Field(i int) Field { return f.fields[i] }

// Focused is the index of the field in focus.
func (f Form) Focused() int { return f.focus }

func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	if len(f.fields) == 0 {
		return f, nil
	}
	held := false
	if h, ok := f.fields[f.focus].(Holder); ok {
		held = h.Holding()
	}
	if k, ok := msg.(tea.KeyPressMsg); ok && !held {
		switch k.String() {
		case "tab":
			return f.move(1)
		case "shift+tab":
			return f.move(-1)
		}
	}
	var cmd tea.Cmd
	f.fields[f.focus], cmd = f.fields[f.focus].Update(msg)
	return f, cmd
}

// move hands the focus by steps, wrapping at either end.
func (f Form) move(by int) (Form, tea.Cmd) {
	f.fields = append([]Field(nil), f.fields...)
	f.fields[f.focus] = f.fields[f.focus].Blur()
	f.focus = (f.focus + by + len(f.fields)) % len(f.fields)
	var cmd tea.Cmd
	f.fields[f.focus], cmd = f.fields[f.focus].Focus()
	return f, cmd
}

// View is every field, width cells wide, labels in one column.
func (f Form) View(width int) string {
	labelW := 0
	for _, fl := range f.fields {
		labelW = max(labelW, draw.Width(fl.Label()))
	}
	rows := make([]string, len(f.fields))
	for i, fl := range f.fields {
		rows[i] = fl.View(i == f.focus, labelW, width)
	}
	return strings.Join(rows, "\n")
}

// row is a field's frame: the focus marker, the label in its column, the control.
func row(th theme.Resolved, focused bool, label string, labelW, width int, control string) string {
	mark := "  "
	if focused {
		mark = th.Accent.Render("›") + " "
	}
	name := th.Dim.Render(draw.Pad(label, labelW))
	if focused {
		name = th.Bold.Render(draw.Pad(label, labelW))
	}
	if labelW == 0 {
		return draw.Fit(mark+control, width)
	}
	return draw.Fit(mark+name+"  "+control, width)
}
