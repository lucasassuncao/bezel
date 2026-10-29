// Package themebrowser provides an inline (not full-screen) terminal table
// listing bezel's built-in themes next to their category and named colours.
// Non-themable: it shows every theme, so it is painted by none.
package themebrowser

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"

	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/component/input"
	"github.com/lucasassuncao/bezel/component/table"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/clip"
	"github.com/lucasassuncao/bezel/internal/fuzzy"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// role is one colour of the palette subtable: its column titles, long and
// short, and where it sits in theme.Colors.
type role struct {
	name, short string
	hex         func(theme.Colors) string
}

var roles = []role{
	{"accent", "acc", func(c theme.Colors) string { return c.Accent }},
	{"select", "sel", func(c theme.Colors) string { return c.Selection }},
	{"border", "bor", func(c theme.Colors) string { return c.Border }},
	{"dim", "dim", func(c theme.Colors) string { return c.Dim }},
	{"text", "txt", func(c theme.Colors) string { return c.Text }},
	{"success", "ok", func(c theme.Colors) string { return c.Success }},
	{"warning", "warn", func(c theme.Colors) string { return c.Warning }},
	{"danger", "err", func(c theme.Colors) string { return c.Danger }},
	{"info", "info", func(c theme.Colors) string { return c.Info }},
	{"bg", "bg", func(c theme.Colors) string { return c.Background }},
}

// narrowRoles are the colours that tell presets apart, all a narrow
// terminal gets: accent, selection, border, dim, success and danger.
var narrowRoles = []int{0, 1, 2, 3, 5, 7}

// layouts are the palette subtables from widest to narrowest; the first
// that fits the terminal is drawn.
var layouts = []struct {
	roles []int
	short bool
}{
	{[]int{9, 0, 1, 2, 3, 4, 5, 6, 7, 8}, false},
	{[]int{9, 0, 1, 2, 3, 4, 5, 6, 7, 8}, true},
	{narrowRoles, true},
}

// The size drawn before the terminal reports its own.
const defaultWidth, defaultHeight = 80, 24

// chrome is the lines around the table: the panel's two edges, the status
// line, the key line and one left for the shell's prompt.
const chrome = 5

// marker is the width table.Model gives its cursor mark.
const marker = 2

// maxRows is the most themes shown at once; the rest scroll.
const maxRows = 20

// margin is the room the panel leaves after its widest row.
const margin = 2

// bindings are the keys the browser answers to besides the table's.
type bindings struct {
	Copy, Filter, Clear, Quit key.Binding
}

// defaultBindings are enter to copy, / to filter, esc to clear it and q to quit.
func defaultBindings() bindings {
	return bindings{
		Copy:   key.New("enter").WithHelp("enter", "copy name"),
		Filter: key.New("/").WithHelp("/", "filter"),
		Clear:  key.New("esc").WithHelp("esc", "clear filter"),
		Quit:   key.New("q", "ctrl+c").WithHelp("q", "quit"),
	}
}

// copiedMsg is how a copy went.
type copiedMsg struct {
	name string
	err  error
}

// entry is one theme before it is drawn: what its row's cells are built from.
type entry struct {
	id, name, category string
	colors             theme.Colors
}

type browserModel struct {
	entries []entry
	tbl     table.Model[string]
	layout  int // into layouts; -1 before the first build
	filter  input.Model
	typing  bool
	keys    bindings
	th      theme.Resolved
	icons   icon.Set
	w, h    int
	boxW    int // the panel's width: its widest row and a margin
	status  string
	write   func(string) error // the clipboard; nil is the system's
}

// newBrowserModel lists the built-in themes in th's colours. dark picks the
// variant of the adaptive default, for the table and for each palette.
func newBrowserModel(th theme.Theme, dark bool) *browserModel {
	all := theme.All()
	var entries []entry
	for _, cat := range theme.Categories() {
		for _, name := range cat.Themes {
			id := strconv.Itoa(len(entries) + 1)
			entries = append(entries, entry{id, name, cat.Name, theme.ResolveColors(all[name], dark)})
		}
	}
	res := theme.Resolve(th, dark)
	m := &browserModel{
		entries: entries,
		layout:  -1,
		keys:    defaultBindings(),
		th:      res,
		icons:   icon.FromEnv(),
		w:       defaultWidth,
		h:       defaultHeight,
	}
	m.filter = m.newFilter()
	m.build()
	return m
}

// newFilter is the empty filter field, drawn on the status line.
func (m *browserModel) newFilter() input.Model {
	return input.New(m.th).WithPrompt("/ ").WithPlaceholder("theme name")
}

// inner is the width the table gets inside the panel's edges and padding.
func (m *browserModel) inner() int { return max(m.boxW-4, 1) }

// build lays the table out for the terminal's size and the filter, keeping
// the cursor on the theme it was on.
func (m *browserModel) build() {
	shown := m.shown()
	at := m.layoutFor()
	m.boxW = min(m.needs(at)+4+margin, m.w)
	height := min(len(shown), maxRows) + 1 // and the header
	if len(shown) > maxRows {
		height++ // the "more" row
	}
	height = max(min(height, m.h-chrome), 2)
	if at == m.layout {
		m.tbl = m.tbl.SetItems(m.items(shown, at)).SetHeight(height)
		return
	}
	prev := ""
	if it, ok := m.tbl.Selected(); ok {
		prev = it.Value
	}
	m.layout = at
	m.tbl = table.New(m.columns(at), m.items(shown, at)).SetHeight(height)
	for i, e := range shown {
		if e.name == prev {
			m.tbl = moveTo(m.tbl, i)
		}
	}
}

// moveTo puts the cursor of a freshly built table on row i, which is how a
// new set of columns keeps the theme the old one was on.
func moveTo(t table.Model[string], i int) table.Model[string] {
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	for t.Cursor() < i {
		next, _ := t.Update(down)
		if next.Cursor() == t.Cursor() {
			break
		}
		t = next
	}
	return t
}

// shown is the themes the filter keeps, best match first, or all of them.
func (m *browserModel) shown() []entry {
	q := strings.TrimSpace(m.filter.Value())
	if q == "" {
		return m.entries
	}
	names := make([]string, len(m.entries))
	for i, e := range m.entries {
		names[i] = e.name
	}
	var out []entry
	for _, i := range fuzzy.Ranked(q, names) {
		out = append(out, m.entries[i])
	}
	return out
}

// layoutFor is the widest palette subtable whose columns fit the terminal.
// It measures every theme, not the filtered ones, so a filter never resizes.
func (m *browserModel) layoutFor() int {
	for i := range layouts {
		if m.needs(i) <= m.w-4 {
			return i
		}
	}
	return len(layouts) - 1
}

// needs is the width the rows of every theme take in layout at.
func (m *browserModel) needs(at int) int {
	widths := table.Fit(m.columns(at), rows(m.items(m.entries, at)), 1<<16)
	total := marker + table.Gap*(len(widths)-1)
	for _, w := range widths {
		total += w
	}
	return total
}

func (m *browserModel) columns(at int) []table.Column {
	cols := []table.Column{{Title: "ID"}, {Title: "Theme"}, {Title: "Category", Min: 4, Flex: true}}
	w := colourWidth(at)
	for _, r := range layouts[at].roles {
		cols = append(cols, table.Column{Title: center(title(r, at), w)})
	}
	return cols
}

func title(r, at int) string {
	if layouts[at].short {
		return roles[r].short
	}
	return roles[r].name
}

// colourWidth is every colour column's width in layout at: its longest
// title's, so the blocks sit evenly spaced.
func colourWidth(at int) int {
	w := 0
	for _, r := range layouts[at].roles {
		w = max(w, len(title(r, at)))
	}
	return w
}

// center pads s with spaces to w cells, as much on the left as the right.
func center(s string, w int) string {
	pad := max(w-draw.Width(s), 0)
	return strings.Repeat(" ", pad/2) + s + strings.Repeat(" ", pad-pad/2)
}

func (m *browserModel) items(shown []entry, at int) []table.Item[string] {
	items := make([]table.Item[string], len(shown))
	for i, e := range shown {
		cells := []string{e.id, e.name, e.category}
		for _, r := range layouts[at].roles {
			cells = append(cells, center(swatch(roles[r].hex(e.colors)), colourWidth(at)))
		}
		items[i] = table.Item[string]{Cells: cells, Value: e.name}
	}
	return items
}

func rows(items []table.Item[string]) [][]string {
	out := make([][]string, len(items))
	for i, it := range items {
		out[i] = it.Cells
	}
	return out
}

// swatch is one colour as a block, with no colour when the theme has none.
func swatch(hex string) string {
	if hex == "" {
		return "  "
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render("██")
}

func (m *browserModel) Init() tea.Cmd { return nil }

// Update sizes the table to the terminal, copies on enter, filters on / and
// quits on q or ctrl+c; the table takes the keys that move its cursor.
func (m *browserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.build()
	case copiedMsg:
		m.status = m.copied(msg)
	case tea.KeyPressMsg:
		return m, m.key(msg)
	default:
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *browserModel) key(k tea.KeyPressMsg) tea.Cmd {
	switch {
	case k.String() == "ctrl+c", !m.typing && key.Matches(k, m.keys.Quit):
		return tea.Quit
	case key.Matches(k, m.keys.Copy):
		m.typing, m.filter = false, m.filter.Blur()
		return m.copy()
	case key.Matches(k, m.keys.Clear):
		if m.typing || m.filter.Value() != "" {
			m.typing, m.filter = false, m.newFilter()
			m.build()
		}
		return nil
	case !m.typing && key.Matches(k, m.keys.Filter):
		m.typing, m.status = true, ""
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Focus()
		return cmd
	}
	if next, act := m.tbl.Update(k); act != table.None || next.Cursor() != m.tbl.Cursor() {
		m.tbl = next
		return nil
	}
	if !m.typing {
		return nil
	}
	before := m.filter.Value()
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(k)
	if m.filter.Value() != before {
		m.build()
	}
	return cmd
}

// copy writes the name under the cursor to the clipboard.
func (m *browserModel) copy() tea.Cmd {
	it, ok := m.tbl.Selected()
	if !ok {
		return nil
	}
	name := it.Value
	return clip.Copy(m.write, name, func(err error) tea.Msg { return copiedMsg{name, err} })
}

func (m *browserModel) copied(msg copiedMsg) string {
	if msg.err != nil {
		return m.icons.Render(icon.Fail, m.th) + " " + m.th.StatusErr.Render(fmt.Sprintf("could not copy %q: %v", msg.name, msg.err))
	}
	return m.icons.Render(icon.OK, m.th) + " " + m.th.StatusOK.Render(fmt.Sprintf("copied %q to the clipboard", msg.name))
}

// View renders inline, without AltScreen: the panel as wide as its rows,
// then the status line and the keys.
func (m *browserModel) View() tea.View {
	body := m.tbl.SetWidth(m.inner()).View(m.th)
	h := strings.Count(body, "\n") + 3
	pos := ""
	if n := m.tbl.Len(); n > 0 {
		pos = fmt.Sprintf("%d/%d", m.tbl.Cursor()+1, n)
	}
	edges := draw.Edges{Title: "Themes", Right: pos, Scroll: m.tbl.Scroll()}
	panel := draw.FramedPanel(layout.Rect{W: m.boxW, H: h}, edges, body, m.th.Panel)
	lines := []string{panel, " " + draw.Fit(m.statusLine(), m.w-1)}
	keys := legend.Pack(m.legend(), m.w-1, 1, m.th.Legend)
	if len(keys) > 0 {
		lines = append(lines, " "+keys[0])
	}
	return tea.NewView(strings.Join(lines, "\n"))
}

func (m *browserModel) statusLine() string {
	switch {
	case m.typing:
		return m.filter.View()
	case m.status != "":
		return m.status
	case m.filter.Value() != "":
		return m.th.Dim.Render("filter: " + m.filter.Value())
	case m.tbl.Len() == 0:
		return m.th.Dim.Render("no theme matches")
	}
	return ""
}

func (m *browserModel) legend() []legend.Entry {
	tk := table.DefaultKeys()
	out := []legend.Entry{{Binding: tk.Up}, {Binding: tk.PageUp}}
	if m.typing {
		return append(out, legend.Entry{Binding: m.keys.Copy}, legend.Entry{Binding: m.keys.Clear})
	}
	out = append(out, legend.Entry{Binding: m.keys.Filter}, legend.Entry{Binding: m.keys.Copy})
	if m.filter.Value() != "" {
		out = append(out, legend.Entry{Binding: m.keys.Clear})
	}
	return append(out, legend.Entry{Binding: m.keys.Quit})
}

// List writes the built-in themes, one line per category: what --list-themes
// prints where there is no terminal to browse in, such as a pipe or a script.
func List(w io.Writer) error {
	for _, c := range theme.Categories() {
		names := slices.Sorted(slices.Values(c.Themes))
		if _, err := fmt.Fprintf(w, "  %-14s %s\n", c.Name, strings.Join(names, ", ")); err != nil {
			return err
		}
	}
	return nil
}

// BrowseInTerminal is an inline table of the built-in themes
// and their named colours (enter copies a name, / filters, q quits), in an
// optional theme.Theme's colours; without a terminal on stdout it Lists.
func BrowseInTerminal(t ...theme.Theme) error {
	if !term.IsTerminal(os.Stdout.Fd()) {
		return List(os.Stdout)
	}
	th := theme.Theme{}
	if len(t) > 0 {
		th = t[0]
	}
	m := newBrowserModel(th, theme.DarkTerminal())
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run theme browser: %w", err)
	}
	return nil
}
