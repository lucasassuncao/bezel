// Package tags is a field that turns what is typed into chips: enter adds the
// text as a tag, backspace on an empty field takes the last one off. For
// filters, labels and recipients; a form takes it as form.NewTags.
package tags

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/input"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// minField is the cells the text field keeps beside the chips, however many.
const minField = 8

// Keys are the bindings that add and take off a tag; the field's own editing
// keys are the input's.
type Keys struct {
	Add, Remove key.Binding
}

// DefaultKeys are enter to add and backspace, on an empty field, to remove.
func DefaultKeys() Keys {
	return Keys{
		Add:    key.New("enter").WithHelp("enter", "add tag"),
		Remove: key.New("backspace", "ctrl+h").WithHelp("backspace", "remove tag"),
	}
}

// Styles are how the chips are painted.
type Styles struct {
	Chip lipgloss.Style // a tag
	More lipgloss.Style // the count of tags that do not fit
}

// DefaultStyles paint a tag as the theme's badge and the count dim.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Chip: th.Badge, More: th.Dim}
}

var _ bezel.Model[Model] = Model{}

// Model is a value: every method returns the new field.
type Model struct {
	in    input.Model
	tags  []string
	keys  Keys
	width int
	th    theme.Resolved
	edit  func(Styles) Styles
}

// New is an empty field drawing in th's colours, as wide as what it holds.
func New(th theme.Resolved) Model {
	return Model{in: input.New(th), keys: DefaultKeys(), th: th}
}

// WithPlaceholder shows s, dimmed, while there is neither a tag nor text.
func (m Model) WithPlaceholder(s string) Model {
	m.in = m.in.WithPlaceholder(s)
	return m
}

// WithKeys rebinds adding and removing a tag.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// WithStyles changes the styles derived from the theme through edit. It
// replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

func (m Model) styles() Styles {
	st := DefaultStyles(m.th)
	if m.edit != nil {
		st = m.edit(st)
	}
	return st
}

// SetTheme repaints the field and its chips.
func (m Model) SetTheme(th theme.Resolved) Model {
	m.th, m.in = th, m.in.SetTheme(th)
	return m
}

// SetSuggestions offers to complete the text from list, as the input does:
// tab takes the one shown.
func (m Model) SetSuggestions(list []string) Model {
	m.in = m.in.SetSuggestions(list)
	return m
}

// SetTags replaces the tags, each one once, blanks dropped.
func (m Model) SetTags(list []string) Model {
	m.tags = nil
	for _, t := range list {
		m, _ = m.add(t)
	}
	return m
}

// SetWidth is the cells the chips and the field share.
func (m Model) SetWidth(w int) Model {
	m.width = max(w, 0)
	return m
}

// Tags is a copy of the tags, in the order they were added.
func (m Model) Tags() []string { return slices.Clone(m.tags) }

// Value is the text typed and not yet a tag.
func (m Model) Value() string { return m.in.Value() }

// Focused reports whether the field takes keys.
func (m Model) Focused() bool { return m.in.Focused() }

// Focus gives the field the keys and starts its caret blinking.
func (m Model) Focus() (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.in, cmd = m.in.Focus()
	return m, cmd
}

// Blur takes the keys away.
func (m Model) Blur() Model {
	m.in = m.in.Blur()
	return m
}

// Blink is the command that keeps a focused caret blinking.
func (m Model) Blink() tea.Cmd { return m.in.Blink() }

// Legend lists adding and removing a tag, then the input's hints.
func (m Model) Legend() []legend.Entry {
	return append(legend.FromBindings([]key.Binding{m.keys.Add, m.keys.Remove}), m.in.Legend()...)
}

// Update adds the text as a tag on Add, reporting [Added], and takes the last
// tag off on Remove while the text is empty, reporting [Removed]. Everything
// else goes to the input. A tag already there is not added twice.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	if k, ok := msg.(tea.KeyPressMsg); ok && m.in.Focused() {
		switch {
		case key.Matches(k, m.keys.Add):
			var added bool
			text := strings.TrimSpace(m.in.Value())
			if m, added = m.add(text); added {
				m.in = m.in.SetValue("")
				return m, bezel.Result{Handled: true, Event: Added{Tag: text}}
			}
			return m, bezel.Result{Handled: true}
		case key.Matches(k, m.keys.Remove) && m.in.Value() == "" && len(m.tags) > 0:
			last := m.tags[len(m.tags)-1]
			m.tags = slices.Clone(m.tags[:len(m.tags)-1])
			return m, bezel.Result{Handled: true, Event: Removed{Tag: last}}
		}
	}
	var r bezel.Result
	m.in, r = m.in.Update(msg)
	return m, r
}

// add appends t when it is a new tag, reporting whether it did.
func (m Model) add(t string) (Model, bool) {
	t = strings.TrimSpace(t)
	if t == "" || slices.Contains(m.tags, t) {
		return m, false
	}
	m.tags = append(slices.Clip(m.tags), t)
	return m, true
}

// View is the chips, then the field in the cells they leave. Chips that do
// not fit give way from the oldest, counted as "+N".
func (m Model) View() string {
	st := m.styles()
	chips := make([]string, len(m.tags))
	for i, t := range m.tags {
		chips[i] = st.Chip.Render(" " + t + " ")
	}
	room := m.width - minField
	from, used := len(chips), 0
	for from > 0 && (m.width == 0 || used+draw.Width(chips[from-1])+1+moreWidth(from-1) <= room) {
		from--
		used += draw.Width(chips[from]) + 1
	}
	var b strings.Builder
	if from > 0 {
		b.WriteString(st.More.Render(more(from)) + " ")
	}
	for _, c := range chips[from:] {
		b.WriteString(c + " ")
	}
	in := m.in
	if m.width > 0 {
		in = in.SetWidth(max(m.width-draw.Width(b.String()), 1))
	}
	return b.String() + in.View()
}

// more is the count of n chips left out.
func more(n int) string { return "+" + strconv.Itoa(n) }

// moreWidth is the cells the count takes for n chips left out, 0 for none.
func moreWidth(n int) int {
	if n == 0 {
		return 0
	}
	return len(more(n)) + 1
}
