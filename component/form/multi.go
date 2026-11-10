package form

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/internal/fuzzy"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// MultiSelect is any number of values out of a list that opens under it:
// enter opens it, typing filters it fuzzily, space checks the option under the
// cursor, enter keeps the checks and esc puts back the ones it opened with.
type MultiSelect struct {
	label   string
	options []string
	checked []bool
	saved   []bool // the checks as the list opened, which Close puts back
	open    bool
	query   string
	shown   []int // the options the query matches, best first
	cursor  int   // into shown
	top     int   // the first of shown in the window
	height  int
	st      *Styles
	set     icon.Set
	keys    Keys
}

// NewMultiSelect offers options with none checked and six rows open.
func NewMultiSelect(label string, options []string) MultiSelect {
	m := MultiSelect{
		label: label, options: slices.Clone(options), checked: make([]bool, len(options)),
		height: 6, set: icon.FromEnv(), keys: DefaultKeys(),
	}
	return m.refilter()
}

// WithChecked checks the options at the indices given; one out of range is
// left out. It is the way a form opens on saved values.
func (m MultiSelect) WithChecked(indices ...int) MultiSelect {
	m.checked = make([]bool, len(m.options))
	for _, i := range indices {
		if i >= 0 && i < len(m.options) {
			m.checked[i] = true
		}
	}
	return m
}

// WithHeight shows at most rows options while open, at least one.
func (m MultiSelect) WithHeight(rows int) MultiSelect {
	m.height = max(rows, 1)
	return m.scrolled()
}

// WithKeys rebinds the field: Press opens it and keeps the checks, Up and Down
// move, Check and CheckAll check, Erase edits the filter, Close puts them back.
func (m MultiSelect) WithKeys(k Keys) MultiSelect {
	m.keys = k
	return m
}

func (m MultiSelect) rebind(k Keys) Field { return m.WithKeys(k) }

func (m MultiSelect) retheme(_ *theme.Resolved, st *Styles) Field {
	m.st = st
	return m
}

// WithIcons draws the list's marks from set instead of icon.FromEnv's.
func (m MultiSelect) WithIcons(set icon.Set) MultiSelect {
	m.set = set
	return m
}

// Checked is the indices of the checked options, in the options' order.
func (m MultiSelect) Checked() []int {
	var out []int
	for i, on := range m.checked {
		if on {
			out = append(out, i)
		}
	}
	return out
}

// Values is the checked options, in the options' order.
func (m MultiSelect) Values() []string {
	var out []string
	for _, i := range m.Checked() {
		out = append(out, m.options[i])
	}
	return out
}

// Filter is the text typed into the open list.
func (m MultiSelect) Filter() string { return m.query }
func (m MultiSelect) Label() string  { return m.label }

// Holding reports whether the list is open, keeping tab and esc from the form.
func (m MultiSelect) Holding() bool           { return m.open }
func (m MultiSelect) Focus() (Field, tea.Cmd) { return m, nil }

// Blur closes the list and keeps its checks, as enter would.
func (m MultiSelect) Blur() Field { return m.closed() }

// closed is the list shut with its filter cleared.
func (m MultiSelect) closed() MultiSelect {
	m.open, m.query, m.saved = false, "", nil
	return m.refilter()
}

// Update opens on Press. While open it takes Up, Down, Check, CheckAll,
// Press, Close, Erase and every key that types a character into the filter.
func (m MultiSelect) Update(msg tea.Msg) (Field, bezel.Result) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, bezel.Result{}
	}
	if !m.open {
		if key.Matches(k, m.keys.Press) {
			m.open, m.saved = true, slices.Clone(m.checked)
			return m.refilter(), bezel.Result{Handled: true}
		}
		return m, bezel.Result{}
	}
	// Check comes before Press, which shares space; a letter is filter text
	// before it is a vim key.
	typed := edit.Typed(k)
	switch {
	case key.Matches(k, m.keys.Check):
		m = m.check()
	case typed != "":
		m.query += typed
		m = m.refilter()
	case key.Matches(k, m.keys.Up):
		m.cursor = max(0, m.cursor-1)
	case key.Matches(k, m.keys.Down):
		m.cursor = max(0, min(len(m.shown)-1, m.cursor+1))
	case key.Matches(k, m.keys.CheckAll):
		m = m.checkAll()
	case key.Matches(k, m.keys.Press):
		m = m.closed()
	case key.Matches(k, m.keys.Close):
		m.checked = m.saved
		m = m.closed()
	case key.Matches(k, m.keys.Erase):
		m.query = edit.Erase(m.query)
		m = m.refilter()
	default:
		return m, bezel.Result{}
	}
	return m.scrolled(), bezel.Result{Handled: true}
}

// check flips the option under the cursor.
func (m MultiSelect) check() MultiSelect {
	if len(m.shown) == 0 {
		return m
	}
	m.checked = slices.Clone(m.checked)
	i := m.shown[m.cursor]
	m.checked[i] = !m.checked[i]
	return m
}

// checkAll checks every option the filter shows, or unchecks them when they
// all are already.
func (m MultiSelect) checkAll() MultiSelect {
	all := true
	for _, i := range m.shown {
		all = all && m.checked[i]
	}
	m.checked = slices.Clone(m.checked)
	for _, i := range m.shown {
		m.checked[i] = !all
	}
	return m
}

// refilter ranks the options against the query and puts the cursor on the best.
func (m MultiSelect) refilter() MultiSelect {
	m.shown = fuzzy.Ranked(m.query, m.options)
	m.cursor, m.top = 0, 0
	return m
}

// window is the shown options drawn, top to end, and how many hide below;
// while some do, the last row says so instead of holding one.
func (m MultiSelect) window() (end, below int) {
	end = min(len(m.shown), m.top+m.height)
	if end < len(m.shown) && m.height > 1 {
		end--
	}
	return end, len(m.shown) - end
}

// scrolled moves the window so the cursor is in it.
func (m MultiSelect) scrolled() MultiSelect {
	m.top = min(m.top, m.cursor)
	for end, _ := m.window(); m.cursor >= end && m.top < m.cursor; end, _ = m.window() {
		m.top++
	}
	return m
}

// openLegend is the keys of the open list; Press is shown by the keys Check
// does not take, so the default reads "enter done".
func (m MultiSelect) openLegend() []legend.Entry {
	var done []string
	for _, k := range m.keys.Press.Keys() {
		if !slices.Contains(m.keys.Check.Keys(), k) {
			done = append(done, k)
		}
	}
	out := legend.FromBindings([]key.Binding{m.keys.Up, m.keys.Check, m.keys.CheckAll})
	if len(done) > 0 {
		out = append(out, legend.New(strings.Join(done, "/"), "done"))
	}
	return append(out, legend.FromBindings([]key.Binding{m.keys.Close})...)
}

func (m MultiSelect) View(focused bool, labelW, width int) string {
	st := styles(m.st)
	edge := st.Edge
	if focused {
		edge = st.FocusedEdge
	}
	value := st.Option.Render("none")
	if v := m.Values(); len(v) > 0 {
		value = strings.Join(v, ", ")
	}
	head := row(st, focused, m.label, labelW, width, edge.Render("‹ ")+value+edge.Render(" ›"))
	if !m.open {
		return head
	}
	indent := strings.Repeat(" ", 2+labelW+2*min(labelW, 1))
	lines := []string{head, draw.Fit(indent+st.FocusedEdge.Render("/ ")+m.query+st.Edge.Render("▏"), width)}
	if len(m.shown) == 0 {
		return strings.Join(append(lines, draw.Fit(indent+"  "+st.Description.Render("no match"), width)), "\n")
	}
	end, below := m.window()
	for n := m.top; n < end; n++ {
		lines = append(lines, draw.Fit(indent+m.optionRow(st, n), width))
	}
	if below > 0 {
		lines = append(lines, draw.Fit(indent+draw.More(below, 2, st.Description), width))
	}
	return strings.Join(lines, "\n")
}

// optionRow is the n-th shown option: the cursor's arrow, its box, its text
// with the letters the filter matched underlined.
func (m MultiSelect) optionRow(st *Styles, n int) string {
	i := m.shown[n]
	lead, text := "  ", st.Option
	if n == m.cursor {
		lead, text = st.Chosen.Render(m.set.Arrow+" "), st.FocusedLabel
	}
	box := st.Edge.Render("[ ]")
	if m.checked[i] {
		box = st.Chosen.Render("[x]")
	}
	_, hits, _ := fuzzy.Match(m.query, m.options[i])
	return lead + box + " " + underlined(m.options[i], hits, text)
}

// underlined is s in st, the runes at hits underlined.
func underlined(s string, hits []int, st lipgloss.Style) string {
	if len(hits) == 0 {
		return st.Render(s)
	}
	var b strings.Builder
	for i, r := range []rune(s) {
		if slices.Contains(hits, i) {
			b.WriteString(st.Underline(true).Render(string(r)))
		} else {
			b.WriteString(st.Render(string(r)))
		}
	}
	return b.String()
}
