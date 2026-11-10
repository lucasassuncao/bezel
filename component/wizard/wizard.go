// Package wizard is a form split into pages the app defines: the steps drawn
// along the top, one page's form under them, and Back and Next underneath.
// A page can check its form before the wizard leaves it.
package wizard

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/form"
	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Page is one step: its title in the steps line, its form, and an optional
// check the form must pass before the wizard moves on.
type Page struct {
	title string
	form  form.Form
	check func(form.Form) error
}

// NewPage is a page titled title over f.
func NewPage(title string, f form.Form) Page { return Page{title: title, form: f} }

// WithCheck runs check on the page's form when Next is pressed; an error keeps
// the wizard on the page and is drawn above the buttons.
func (p Page) WithCheck(check func(form.Form) error) Page {
	p.check = check
	return p
}

// Title is the page's name in the steps line.
func (p Page) Title() string { return p.title }

// Form is the page's form as it is now.
func (p Page) Form() form.Form { return p.form }

// Stepped reports that the wizard moved to the page at Step.
type Stepped struct{ Step int }

// Done reports that Next was pressed on the last page and its check passed.
type Done struct{}

// Keys are the bindings that turn the page. tab and shift+tab stay the form's.
type Keys struct {
	Next, Back key.Binding
}

// DefaultKeys are pgdown and pgup.
func DefaultKeys() Keys {
	return Keys{
		Next: key.New("pgdown").WithHelp("pgdn", "next page"),
		Back: key.New("pgup").WithHelp("pgup", "previous page"),
	}
}

// Styles are how the steps, the error and the buttons are painted; each
// page's form keeps its own.
type Styles struct {
	Done    lipgloss.Style // a step behind the current one
	Current lipgloss.Style
	Ahead   lipgloss.Style // a step not reached yet
	Line    lipgloss.Style // the rule between steps
	Error   lipgloss.Style // a check that failed
	Button  draw.ButtonStyle
}

// DefaultStyles mark the steps behind in success and the current one in the
// accent, mute the rest and paint a failed check in danger.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{
		Done: th.Success, Current: th.Accent.Bold(true), Ahead: th.Muted, Line: th.Muted,
		Error: th.Danger, Button: th.Button,
	}
}

var _ bezel.Model[Model] = Model{}

// Model is a value: every method returns the new wizard.
type Model struct {
	pages              []Page
	at                 int
	err                string // the current page's failed check, until a key changes the form
	width              int
	keys               Keys
	set                icon.Set
	back, next, finish string
	th                 theme.Resolved
	edit               func(Styles) Styles
	st                 Styles
}

// New is a wizard over pages, painted in th, on the first page. Each page's
// form is repainted in th too.
func New(th theme.Resolved, pages ...Page) Model {
	m := Model{
		pages: slices.Clone(pages), keys: DefaultKeys(), set: icon.FromEnv(),
		back: "Back", next: "Next", finish: "Finish",
	}
	return m.SetTheme(th)
}

// WithKeys rebinds the page turns; each form keeps its own keys.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// WithLabels names the buttons: back, next, and finish for Next on the last page.
func (m Model) WithLabels(back, next, finish string) Model {
	m.back, m.next, m.finish = back, next, finish
	return m
}

// WithIcons draws the done and failed marks from set instead of icon.FromEnv's.
func (m Model) WithIcons(set icon.Set) Model {
	m.set = set
	return m
}

// WithStyles changes the styles derived from the theme through edit, now and
// after every SetTheme. It replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m.SetTheme(m.th)
}

// SetTheme repaints the wizard and every page's form.
func (m Model) SetTheme(th theme.Resolved) Model {
	m.th, m.st = th, DefaultStyles(th)
	if m.edit != nil {
		m.st = m.edit(m.st)
	}
	m.pages = slices.Clone(m.pages)
	for i := range m.pages {
		m.pages[i].form = m.pages[i].form.SetTheme(th)
	}
	return m
}

// SetWidth changes the cells every row of the wizard takes.
func (m Model) SetWidth(w int) Model {
	m.width = w
	return m
}

// Step is the index of the page shown.
func (m Model) Step() int { return m.at }

// Page is the i-th page as it is now, the zero Page when i is out of range;
// read its values through form.Get on its Form.
func (m Model) Page(i int) Page {
	if i < 0 || i >= len(m.pages) {
		return Page{}
	}
	return m.pages[i]
}

// Err is the current page's failed check, "" when there is none.
func (m Model) Err() string { return m.err }

// Init is the shown form's command, such as its input's caret blink.
func (m Model) Init() tea.Cmd {
	if len(m.pages) == 0 {
		return nil
	}
	return m.pages[m.at].form.Init()
}

// Legend is the shown form's keys, then the page turns unless a field holds them.
func (m Model) Legend() []legend.Entry {
	if len(m.pages) == 0 {
		return nil
	}
	out := m.pages[m.at].form.Legend()
	if m.holding() {
		return out
	}
	return append(out, legend.FromBindings([]key.Binding{m.keys.Next, m.keys.Back})...)
}

// holding reports whether the focused field keeps every key, as an open select does.
func (m Model) holding() bool {
	f := m.pages[m.at].form
	h, ok := f.Field(f.Focused()).(form.Holder)
	return ok && h.Holding()
}

// Update turns the page on Next and Back, presses the button a click lands on
// and hands every other message to the shown form. Next on the last page
// reports Done; turning reports Stepped.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	if len(m.pages) == 0 {
		return m, bezel.Result{}
	}
	if c, ok := msg.(bezel.ClickMsg); ok {
		return m.click(c)
	}
	if k, ok := msg.(tea.KeyPressMsg); ok && !m.holding() {
		switch {
		case key.Matches(k, m.keys.Next):
			return m.forward()
		case key.Matches(k, m.keys.Back):
			return m.backward()
		}
	}
	var r bezel.Result
	m.pages = slices.Clone(m.pages)
	m.pages[m.at].form, r = m.pages[m.at].form.Update(msg)
	if r.Handled {
		m.err = ""
	}
	return m, r
}

// forward runs the page's check, then turns to the next page or reports Done.
func (m Model) forward() (Model, bezel.Result) {
	p := m.pages[m.at]
	if p.check != nil {
		if err := p.check(p.form); err != nil {
			m.err = err.Error()
			return m, bezel.Result{Handled: true}
		}
	}
	m.err = ""
	if m.at == len(m.pages)-1 {
		return m, bezel.Result{Handled: true, Event: Done{}}
	}
	return m.turned(m.at + 1)
}

// backward turns to the page before, if there is one.
func (m Model) backward() (Model, bezel.Result) {
	if m.at == 0 {
		return m, bezel.Result{Handled: true}
	}
	m.err = ""
	return m.turned(m.at - 1)
}

// turned shows page i, starting its form.
func (m Model) turned(i int) (Model, bezel.Result) {
	m.at = i
	return m, bezel.Result{Handled: true, Event: Stepped{Step: i}, Cmd: m.pages[i].form.Init()}
}

// The rows above a page's form: the steps line and a blank one.
const formTop = 2

// click hands a click on the form to it, shifted to the form's rows, and turns
// the page on a click on Back or Next. Anywhere else it is not handled.
func (m Model) click(c bezel.ClickMsg) (Model, bezel.Result) {
	formH := m.formHeight()
	if c.Y >= formTop && c.Y < formTop+formH {
		c.Y -= formTop
		var r bezel.Result
		m.pages = slices.Clone(m.pages)
		m.pages[m.at].form, r = m.pages[m.at].form.Update(c)
		return m, r
	}
	if c.Y != m.footerRow() {
		return m, bezel.Result{}
	}
	backX, nextX, end := m.buttonCols()
	switch {
	case m.at > 0 && c.X >= backX && c.X < nextX-2:
		return m.backward()
	case c.X >= nextX && c.X < end:
		return m.forward()
	}
	return m, bezel.Result{}
}

// formHeight is the rows the shown form takes at the wizard's width.
func (m Model) formHeight() int {
	_, h := draw.BlockSize(m.pages[m.at].form.SetWidth(m.width).View())
	return h
}

// footerRow is the row Back and Next are drawn on.
func (m Model) footerRow() int {
	row := formTop + m.formHeight() + 1
	if m.err != "" {
		row++
	}
	return row
}

// buttonCols is where Back starts, where Next starts and where it ends; Back
// is not drawn on the first page and takes no cells there.
func (m Model) buttonCols() (backX, nextX, end int) {
	backX, nextX = 2, 2
	if m.at > 0 {
		nextX += draw.Width(draw.Button(m.back, m.st.Button, false)) + 2
	}
	return backX, nextX, nextX + draw.Width(draw.Button(m.nextLabel(), m.st.Button, true))
}

// nextLabel is Next's name, or Finish's on the last page.
func (m Model) nextLabel() string {
	if m.at == len(m.pages)-1 {
		return m.finish
	}
	return m.next
}

// View is the steps, the shown form, a failed check and the buttons, SetWidth
// cells wide; nothing until SetWidth sizes it or with no pages.
func (m Model) View() string {
	if m.width <= 0 || len(m.pages) == 0 {
		return ""
	}
	lines := []string{draw.Fit(m.steps(), m.width), ""}
	lines = append(lines, m.pages[m.at].form.SetWidth(m.width).View(), "")
	if m.err != "" {
		lines = append(lines, draw.Fit("  "+m.st.Error.Render(m.set.Fail+" "+draw.OneLine(m.err)), m.width))
	}
	buttons := "  "
	if m.at > 0 {
		buttons += draw.Button(m.back, m.st.Button, false) + "  "
	}
	buttons += draw.Button(m.nextLabel(), m.st.Button, true)
	return strings.Join(append(lines, draw.Fit(buttons, m.width)), "\n")
}

// steps is every page's title with its state, or "2/4 Title" when that does
// not fit the width.
func (m Model) steps() string {
	parts := make([]string, len(m.pages))
	for i, p := range m.pages {
		title := draw.OneLine(p.title)
		switch {
		case i < m.at:
			parts[i] = m.st.Done.Render(m.set.OK + " " + title)
		case i == m.at:
			parts[i] = m.st.Current.Render(fmt.Sprintf("%d %s", i+1, title))
		default:
			parts[i] = m.st.Ahead.Render(fmt.Sprintf("%d %s", i+1, title))
		}
	}
	line := "  " + strings.Join(parts, m.st.Line.Render(" "+m.set.Line+m.set.Line+" "))
	if draw.Width(line) <= m.width {
		return line
	}
	return "  " + m.st.Current.Render(fmt.Sprintf("%d/%d %s", m.at+1, len(m.pages), draw.OneLine(m.pages[m.at].title)))
}
