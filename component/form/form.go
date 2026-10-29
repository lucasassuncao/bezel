// Package form is the controls that ask for a value: a text input, a select,
// checkboxes, radios, a toggle and a button, stacked in a Form that moves focus
// with tab and shift+tab. Each control is one row, its label in a shared column,
// painted in the theme the Form was built with.
package form

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
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

// Initer is a Field with a command to run when the form opens focused on it.
// Form.Init runs it.
type Initer interface{ Init() tea.Cmd }

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

// DefaultKeys are tab and shift+tab, space or enter, the arrows and esc.
func DefaultKeys() Keys {
	return Keys{
		Next:  key.New("tab").WithHelp("tab", "next"),
		Prev:  key.New("shift+tab").WithHelp("shift+tab", "previous"),
		Press: key.New("space", "enter").WithHelp("space", "press"),
		Left:  key.New("left").WithHelp("←/→", "choose"),
		Right: key.New("right"),
		Up:    key.New("up").WithHelp("↑/↓", "move"),
		Down:  key.New("down"),
		Close: key.New("esc").WithHelp("esc", "close"),
	}
}

// VimKeys are DefaultKeys with h/l added to a radio and a toggle and k/j to
// an open select. An input still takes them as text.
func VimKeys() Keys {
	k := DefaultKeys()
	k.Left, k.Right = key.New("left", "h").WithHelp("h/l", "choose"), key.New("right", "l")
	k.Up, k.Down = key.New("up", "k").WithHelp("k/j", "move"), key.New("down", "j")
	return k
}

// keyed is a bezel control, which Form.WithKeys rebinds along with the form.
type keyed interface{ rebind(k Keys) Field }

// themed is a bezel control, which paints in the form's theme and styles.
type themed interface {
	retheme(th *theme.Resolved, st *Styles) Field
}

// Styles are how a form paints its controls; one set serves every control.
type Styles struct {
	Marker       lipgloss.Style // the › before the focused row
	Label        lipgloss.Style
	FocusedLabel lipgloss.Style // the focused row's label, and the option under an open select's cursor
	Edge         lipgloss.Style // brackets and empty marks
	FocusedEdge  lipgloss.Style // the brackets of the focused input or select
	Chosen       lipgloss.Style // a ticked box, the chosen radio, an open select's arrow
	Option       lipgloss.Style // an option not chosen, and "off"
	On           lipgloss.Style // a switch that is on
	Button       draw.ButtonStyle
}

// DefaultStyles mark the focus and the chosen in the accent, mute the
// empty marks and paint a switch that is on in the success colour.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{
		Marker: th.Accent, Label: th.Dim, FocusedLabel: th.Bold, Edge: th.Muted, FocusedEdge: th.Accent,
		Chosen: th.Accent, Option: th.Dim, On: th.Success, Button: th.Button,
	}
}

// Form is fields in a column. tab and shift+tab move the focus; every other key
// goes to the field in focus.
type Form struct {
	fields []Field
	focus  int
	keys   Keys
	width  int
	th     *theme.Resolved // shared by every copy and never written
	edit   func(Styles) Styles
}

// New is a form over fields, painted in th and focused on the first; nil
// fields are left out. Init starts the focused field.
func New(th theme.Resolved, fields ...Field) Form {
	f := Form{fields: slices.DeleteFunc(slices.Clone(fields), func(fl Field) bool { return fl == nil }), keys: DefaultKeys()}
	f = f.SetTheme(th)
	if len(f.fields) > 0 {
		f.fields[0], _ = f.fields[0].Focus()
	}
	return f
}

// Init is the focused field's command when it is an Initer, such as the
// caret blink of the input the form opens on.
func (f Form) Init() tea.Cmd {
	if len(f.fields) == 0 {
		return nil
	}
	if in, ok := f.fields[f.focus].(Initer); ok {
		return in.Init()
	}
	return nil
}

// SetTheme repaints every bezel control in the form; a Field of the app's
// own keeps its look.
func (f Form) SetTheme(th theme.Resolved) Form {
	f.th = &th
	return f.restyle()
}

// WithStyles changes the styles derived from the theme through edit, now and
// after every SetTheme. It replaces an earlier edit.
func (f Form) WithStyles(edit func(Styles) Styles) Form {
	f.edit = edit
	return f.restyle()
}

// restyle hands every bezel control the theme and the styles derived from it.
func (f Form) restyle() Form {
	th := &noTheme
	if f.th != nil {
		th = f.th
	}
	st := DefaultStyles(*th)
	if f.edit != nil {
		st = f.edit(st)
	}
	f.fields = slices.Clone(f.fields)
	for i, fl := range f.fields {
		if tf, ok := fl.(themed); ok {
			f.fields[i] = tf.retheme(th, &st)
		}
	}
	return f
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

// Field is the i-th field as it is now, nil when i is out of range; assert it
// to read its value, or use Get.
func (f Form) Field(i int) Field {
	if i < 0 || i >= len(f.fields) {
		return nil
	}
	return f.fields[i]
}

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

// Legend is the keys for where the focus is: an open select's own, or tab
// and shift+tab with what the focused control answers to. An app's own Field
// adds nothing.
func (f Form) Legend() []legend.Entry {
	if len(f.fields) == 0 {
		return nil
	}
	k := f.keys
	var own []key.Binding
	switch fl := f.fields[f.focus].(type) {
	case Select:
		if fl.Holding() {
			return legend.FromBindings([]key.Binding{k.Up, k.Press, k.Close})
		}
		own = []key.Binding{k.Press}
	case Checkbox, Toggle, Button:
		own = []key.Binding{k.Press}
	case Radio:
		own = []key.Binding{k.Left}
	case Input:
		return append(legend.FromBindings([]key.Binding{k.Next, k.Prev}), fl.in.Legend()...)
	}
	return legend.FromBindings(append([]key.Binding{k.Next, k.Prev}, own...))
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
// pressable is a field a click runs, as space would: a button.
type pressable interface{ press() tea.Cmd }

// flippable is a field a click turns over: a checkbox or a toggle.
type flippable interface{ flip() Field }

// SetWidth changes the cells every row of the form takes.
func (f Form) SetWidth(w int) Form {
	f.width = w
	return f
}

// Click focuses the field drawn on row y of View and presses it when it is a
// button, a checkbox or a toggle. A row past the fields does nothing.
func (f Form) Click(y int) (Form, tea.Cmd) {
	i := f.fieldAt(y)
	if i < 0 {
		return f, nil
	}
	var focus tea.Cmd
	if i != f.focus {
		f, focus = f.move(i - f.focus)
	}
	switch fl := f.fields[i].(type) {
	case pressable:
		return f, tea.Batch(focus, fl.press())
	case flippable:
		f.fields = slices.Clone(f.fields)
		f.fields[i] = fl.flip()
	}
	return f, focus
}

// fieldAt is the field drawn on row y, counting the rows an open select
// takes; -1 past them.
func (f Form) fieldAt(y int) int {
	labelW := 0
	for _, fl := range f.fields {
		labelW = max(labelW, draw.Width(fl.Label()))
	}
	top := 0
	for i, fl := range f.fields {
		_, h := draw.BlockSize(fl.View(i == f.focus, labelW, f.width))
		if y >= top && y < top+h {
			return i
		}
		top += h
	}
	return -1
}

func (f Form) move(by int) (Form, tea.Cmd) {
	f.fields = slices.Clone(f.fields)
	f.fields[f.focus] = f.fields[f.focus].Blur()
	f.focus = (f.focus + by + len(f.fields)) % len(f.fields)
	var cmd tea.Cmd
	f.fields[f.focus], cmd = f.fields[f.focus].Focus()
	return f, cmd
}

// View is every field, SetWidth cells wide, labels in one column; nothing
// until SetWidth sizes it.
func (f Form) View() string {
	if f.width <= 0 {
		return ""
	}
	labelW := 0
	for _, fl := range f.fields {
		labelW = max(labelW, draw.Width(fl.Label()))
	}
	rows := make([]string, len(f.fields))
	for i, fl := range f.fields {
		rows[i] = fl.View(i == f.focus, labelW, f.width)
	}
	return strings.Join(rows, "\n")
}

// noTheme and noStyles are what a field outside a form works with; never written.
var (
	noTheme  theme.Resolved
	noStyles Styles
)

// styles is a field's styles, held by pointer so a copy of the field costs a
// word; the zero ones for a field no form has painted.
func styles(st *Styles) *Styles {
	if st == nil {
		return &noStyles
	}
	return st
}

// row is a field's frame: the focus marker, the label in its column, the control.
func row(st *Styles, focused bool, label string, labelW, width int, control string) string {
	mark := "  "
	if focused {
		mark = st.Marker.Render("›") + " "
	}
	name := st.Label.Render(draw.Pad(label, labelW))
	if focused {
		name = st.FocusedLabel.Render(draw.Pad(label, labelW))
	}
	if labelW == 0 {
		return draw.Fit(mark+control, width)
	}
	return draw.Fit(mark+name+"  "+control, width)
}
