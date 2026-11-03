// Package breadcrumb is a path in one row, kv › app › prod, with a cursor on
// a part: left and right move it and enter chooses the path up to it. A path
// wider than its row gives way from the middle, as an ellipsis.
package breadcrumb

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// ellipsis stands for the parts a narrow row leaves out.
const ellipsis = "…"

// Keys are the bindings that move the cursor and choose. Each pair's second
// has no help of its own: the first names both.
type Keys struct {
	Left, Right key.Binding
	Home, End   key.Binding
	Enter       key.Binding
}

// DefaultKeys are the arrows for a part, home/end for the first and last,
// and enter to choose.
func DefaultKeys() Keys {
	return Keys{
		Left:  key.New("left").WithHelp("←/→", "part"),
		Right: key.New("right"),
		Home:  key.New("home").WithHelp("home/end", "first/last"),
		End:   key.New("end"),
		Enter: key.New("enter").WithHelp("enter", "choose"),
	}
}

// VimKeys are DefaultKeys with vim's letters added: h/l for a part and g/G
// for the first and last.
func VimKeys() Keys {
	k := DefaultKeys()
	k.Left, k.Right = key.New("left", "h").WithHelp("←/→ h/l", "part"), key.New("right", "l")
	k.Home, k.End = key.New("home", "g").WithHelp("g/G", "first/last"), key.New("end", "G")
	return k
}

// Styles are how a breadcrumb paints its parts.
type Styles struct {
	Part      lipgloss.Style // a part before the last
	Last      lipgloss.Style // where the path ends
	Cursor    lipgloss.Style // the part under the cursor, while focused
	Separator lipgloss.Style
	More      lipgloss.Style // the ellipsis for the parts left out
}

// DefaultStyles mute the parts and separators before the last, bold the last
// and mark the cursor as the caret.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Part: th.Muted, Last: th.Bold, Cursor: th.Caret, Separator: th.Muted, More: th.Muted}
}

var _ bezel.Model[Model] = Model{}

// Model is a value: every method returns the new breadcrumb.
type Model struct {
	parts  []string
	cursor int
	sep    string
	width  int // 0 for unbounded
	blur   bool
	keys   Keys
	edit   func(Styles) Styles
}

// New is the path of parts with the cursor on the last.
func New(parts []string) Model {
	return Model{sep: " › ", keys: DefaultKeys()}.SetParts(parts)
}

// WithSeparator draws sep between the parts instead of " › ".
func (m Model) WithSeparator(sep string) Model {
	m.sep = sep
	return m
}

// WithKeys rebinds the breadcrumb.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// WithStyles changes the styles derived from the theme through edit, under
// whatever theme View gets. It replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

func (m Model) styles(th theme.Resolved) Styles {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	return st
}

// SetParts replaces the path and puts the cursor on its last part.
func (m Model) SetParts(parts []string) Model {
	m.parts = slices.Clone(parts)
	m.cursor = max(len(m.parts)-1, 0)
	return m
}

// SetCursor puts the cursor on part i, held within the path.
func (m Model) SetCursor(i int) Model {
	m.cursor = min(max(i, 0), max(len(m.parts)-1, 0))
	return m
}

// SetWidth bounds the row View draws; 0 is unbounded.
func (m Model) SetWidth(w int) Model {
	m.width = max(w, 0)
	return m
}

// Focus shows the cursor and takes the keys again.
func (m Model) Focus() Model {
	m.blur = false
	return m
}

// Blur hides the cursor and leaves the keys to the app.
func (m Model) Blur() Model {
	m.blur = true
	return m
}

// Focused reports whether the breadcrumb shows its cursor and takes keys.
func (m Model) Focused() bool { return !m.blur }

// Parts is the whole path.
func (m Model) Parts() []string { return slices.Clone(m.parts) }

// Cursor is the index of the part under the cursor.
func (m Model) Cursor() int { return m.cursor }

// Selected is the part under the cursor, false for an empty path.
func (m Model) Selected() (string, bool) {
	if len(m.parts) == 0 {
		return "", false
	}
	return m.parts[m.cursor], true
}

// Path is the path up to the cursor's part, that part included.
func (m Model) Path() []string {
	if len(m.parts) == 0 {
		return nil
	}
	return slices.Clone(m.parts[:m.cursor+1])
}

// Legend lists the keys the breadcrumb answers to.
func (m Model) Legend() []legend.Entry {
	k := m.keys
	return legend.FromBindings([]key.Binding{k.Left, k.Right, k.Home, k.End, k.Enter})
}

// Update moves the cursor on its keys unless blurred; enter, or a click on
// the part already under the cursor, reports a [Chosen]. A click on another
// part moves there. Anything else comes back unhandled.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	if len(m.parts) == 0 {
		return m, bezel.Result{}
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.blur {
			return m, bezel.Result{}
		}
		k := m.keys
		switch {
		case key.Matches(msg, k.Left):
			return m.SetCursor(m.cursor - 1), bezel.Result{Handled: true}
		case key.Matches(msg, k.Right):
			return m.SetCursor(m.cursor + 1), bezel.Result{Handled: true}
		case key.Matches(msg, k.Home):
			return m.SetCursor(0), bezel.Result{Handled: true}
		case key.Matches(msg, k.End):
			return m.SetCursor(len(m.parts) - 1), bezel.Result{Handled: true}
		case key.Matches(msg, k.Enter):
			return m, bezel.Result{Handled: true, Event: Chosen{Path: m.Path()}}
		}
	case bezel.ClickMsg:
		i := m.partAt(msg.X)
		switch {
		case msg.Y != 0 || i < 0:
			return m, bezel.Result{}
		case i == m.cursor:
			return m, bezel.Result{Handled: true, Event: Chosen{Path: m.Path()}}
		}
		return m.SetCursor(i), bezel.Result{Handled: true}
	}
	return m, bezel.Result{}
}

// shown is the parts View draws, -1 where the ellipsis goes: as many from
// the end as fit after the first, never leaving out the cursor's part.
func (m Model) shown() []int {
	all := make([]int, len(m.parts))
	for i := range all {
		all[i] = i
	}
	if m.width == 0 || m.measure(all) <= m.width {
		return all
	}
	keep := all
	for hide := 1; hide < len(m.parts)-1; hide++ {
		if keep = m.hiding(hide); m.measure(keep) <= m.width {
			break
		}
	}
	return keep
}

// hiding is the parts with 1 to hide left out, each run of them an ellipsis,
// and the cursor's part shown among them.
func (m Model) hiding(hide int) []int {
	keep := []int{0}
	if c := m.cursor; c >= 1 && c <= hide {
		if c > 1 {
			keep = append(keep, -1)
		}
		keep = append(keep, c)
		if c < hide {
			keep = append(keep, -1)
		}
	} else {
		keep = append(keep, -1)
	}
	for i := hide + 1; i < len(m.parts); i++ {
		keep = append(keep, i)
	}
	return keep
}

// measure is the cells a row of these parts takes.
func (m Model) measure(idx []int) int {
	w := 0
	for n, i := range idx {
		if n > 0 {
			w += draw.Width(m.sep)
		}
		w += draw.Width(m.label(i))
	}
	return w
}

func (m Model) label(i int) string {
	if i < 0 {
		return ellipsis
	}
	return m.parts[i]
}

// partAt is the part drawn at cell x of View, -1 for a separator, the
// ellipsis or past the end.
func (m Model) partAt(x int) int {
	for n, i := range m.shown() {
		if n > 0 {
			x -= draw.Width(m.sep)
		}
		w := draw.Width(m.label(i))
		if x >= 0 && x < w {
			return i
		}
		if x < 0 {
			return -1
		}
		x -= w
	}
	return -1
}

// View is the path in one row, cut at SetWidth when even the shortest form
// does not fit.
func (m Model) View(th theme.Resolved) string {
	st := m.styles(th)
	var b strings.Builder
	for n, i := range m.shown() {
		if n > 0 {
			b.WriteString(st.Separator.Render(m.sep))
		}
		b.WriteString(m.paint(i, st).Render(m.label(i)))
	}
	if m.width > 0 {
		return draw.Truncate(b.String(), m.width)
	}
	return b.String()
}

// paint is the style part i wears.
func (m Model) paint(i int, st Styles) lipgloss.Style {
	switch {
	case i < 0:
		return st.More
	case i == m.cursor && !m.blur:
		return st.Cursor
	case i == len(m.parts)-1:
		return st.Last
	}
	return st.Part
}
