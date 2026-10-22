// Package form is the controls that ask for a value: a text input, a select,
// checkboxes, radios, a toggle and a button, stacked in a Form that moves focus
// with tab and shift+tab. Each control is one row, its label in a shared column.
package form

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
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

// Keys are the bindings a form and its controls answer to. Each control reads
// the ones that concern it, and the Form only Next and Prev.
type Keys struct {
	Next, Prev  key.Binding // the form: move the focus
	Press       key.Binding // a checkbox, toggle or button; opens and picks in a select
	Left, Right key.Binding // a radio; either one flips a toggle
	Up, Down    key.Binding // an open select
	Close       key.Binding // an open select, leaving its value as it was
}

// DefaultKeys are tab and shift+tab, space or enter, the arrows with h/l and
// k/j, and esc.
func DefaultKeys() Keys {
	return Keys{
		Next:  key.NewBinding(key.WithKeys("tab")),
		Prev:  key.NewBinding(key.WithKeys("shift+tab")),
		Press: key.NewBinding(key.WithKeys("space", "enter")),
		Left:  key.NewBinding(key.WithKeys("left", "h")),
		Right: key.NewBinding(key.WithKeys("right", "l")),
		Up:    key.NewBinding(key.WithKeys("up", "k")),
		Down:  key.NewBinding(key.WithKeys("down", "j")),
		Close: key.NewBinding(key.WithKeys("esc")),
	}
}

// keyed is a bezel control, which Form.WithKeys rebinds along with the form.
type keyed interface{ rebind(k Keys) Field }

// Form is fields in a column. tab and shift+tab move the focus; every other key
// goes to the field in focus.
type Form struct {
	fields []Field
	focus  int
	keys   Keys
}

// New is a form focused on its first field.
func New(fields ...Field) (Form, tea.Cmd) {
	f := Form{fields: slices.Clone(fields), keys: DefaultKeys()}
	if len(fields) == 0 {
		return f, nil
	}
	var cmd tea.Cmd
	f.fields[0], cmd = f.fields[0].Focus()
	return f, cmd
}

// WithKeys rebinds the form and every bezel control in it; a Field of the
// app's own keeps its keys.
func (f Form) WithKeys(k Keys) Form {
	f.keys = k
	f.fields = slices.Clone(f.fields)
	for i, fl := range f.fields {
		if kf, ok := fl.(keyed); ok {
			f.fields[i] = kf.rebind(k)
		}
	}
	return f
}

// Field is the i-th field as it is now; assert it to read its value, or use Get.
func (f Form) Field(i int) Field { return f.fields[i] }

// Get is the i-th field as a T, and false when i is out of range or the field
// is of another type: a reordered form fails a check instead of a panic.
func Get[T Field](f Form, i int) (T, bool) {
	if i < 0 || i >= len(f.fields) {
		var zero T
		return zero, false
	}
	v, ok := f.fields[i].(T)
	return v, ok
}

// Focused is the index of the field in focus.
func (f Form) Focused() int { return f.focus }

// Update moves the focus on tab and shift+tab, unless the focused field holds
// them, and hands every other message to the focused field.
func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	if len(f.fields) == 0 {
		return f, nil
	}
	held := false
	if h, ok := f.fields[f.focus].(Holder); ok {
		held = h.Holding()
	}
	if k, ok := msg.(tea.KeyPressMsg); ok && !held {
		switch {
		case key.Matches(k, f.keys.Next):
			return f.move(1)
		case key.Matches(k, f.keys.Prev):
			return f.move(-1)
		}
	}
	f.fields = slices.Clone(f.fields)
	var cmd tea.Cmd
	f.fields[f.focus], cmd = f.fields[f.focus].Update(msg)
	return f, cmd
}

// move hands the focus by steps, wrapping at either end.
func (f Form) move(by int) (Form, tea.Cmd) {
	f.fields = slices.Clone(f.fields)
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
