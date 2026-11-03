package form

import (
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/tags"
	"github.com/lucasassuncao/bezel/theme"
)

// Tags is a field whose text becomes chips: enter adds a tag, backspace on an
// empty field takes the last one off.
type Tags struct {
	label string
	tg    tags.Model
	st    *Styles
}

// NewTags is an empty field 30 cells wide.
func NewTags(label string) Tags {
	return Tags{label: label, tg: tags.New(theme.Resolved{}).SetWidth(30)}
}

// WithTags is the field holding list, the way a form opens on saved values.
func (t Tags) WithTags(list []string) Tags {
	t.tg = t.tg.SetTags(list)
	return t
}

// WithSuggestions offers to complete a tag from list; tab takes the one shown.
func (t Tags) WithSuggestions(list []string) Tags {
	t.tg = t.tg.SetSuggestions(list)
	return t
}

// WithPlaceholder shows s, dimmed, while there is neither a tag nor text.
func (t Tags) WithPlaceholder(s string) Tags {
	t.tg = t.tg.WithPlaceholder(s)
	return t
}

// WithWidth gives the chips and the text width cells.
func (t Tags) WithWidth(width int) Tags {
	t.tg = t.tg.SetWidth(width)
	return t
}

// Tags is a copy of the tags, in the order they were added.
func (t Tags) Tags() []string { return t.tg.Tags() }
func (t Tags) Label() string  { return t.label }

// Init starts the caret blinking when the form opens focused on the field.
func (t Tags) Init() tea.Cmd { return t.tg.Blink() }

func (t Tags) retheme(th *theme.Resolved, st *Styles) Field {
	t.st, t.tg = st, t.tg.SetTheme(*th)
	return t
}

// Update hands msg to the tags and reports what they reported.
func (t Tags) Update(msg tea.Msg) (Field, bezel.Result) {
	var r bezel.Result
	t.tg, r = t.tg.Update(msg)
	return t, r
}

func (t Tags) Focus() (Field, tea.Cmd) {
	var cmd tea.Cmd
	t.tg, cmd = t.tg.Focus()
	return t, cmd
}

func (t Tags) Blur() Field {
	t.tg = t.tg.Blur()
	return t
}

func (t Tags) View(focused bool, labelW, width int) string {
	edge := styles(t.st).Edge
	if focused {
		edge = styles(t.st).FocusedEdge
	}
	return row(styles(t.st), focused, t.label, labelW, width, edge.Render("[")+t.tg.View()+edge.Render("]"))
}
