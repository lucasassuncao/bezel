// Package cursor is a text field's caret on its own: it blinks, stays lit or
// hides, over the character it sits on, at the pace the theme sets.
package cursor

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/theme"
)

// Mode is how a focused cursor shows.
type Mode int

const (
	Blinking Mode = iota // lit and dark in turn, at the theme's CaretBlink
	Steady               // lit while focused
	Hidden               // never lit; the character shows plain
)

var modeNames = [...]string{"blinking", "steady", "hidden"}

func (m Mode) String() string {
	if m < 0 || int(m) >= len(modeNames) {
		return "unknown"
	}
	return modeNames[m]
}

// BlinkMsg flips the cursor whose ID it carries. A host hands it on the way
// it hands on keys; a cursor ignores every other cursor's.
type BlinkMsg struct {
	ID  int64
	tag int64 // the blink loop that armed it; only the newest flips
}

// Styles are how the character is painted.
type Styles struct {
	Cursor lipgloss.Style // while the cursor shows
	Text   lipgloss.Style // while it is dark, hidden or blurred
}

// DefaultStyles paints a lit cursor in the theme's Caret and the rest plain.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Cursor: th.Caret, Text: lipgloss.NewStyle()}
}

var _ bezel.Model[Model] = Model{}

// Model is a value: every method returns the new cursor.
type Model struct {
	char    string
	mode    Mode
	focused bool
	on      bool // false for the dark half of a blink
	id, tag int64
	edit    func(Styles) Styles
	th      *theme.Resolved // shared by every copy and never written
	st      *Styles         // DefaultStyles of th, edited; rebuilt, never written
}

// New is a blinking cursor over a space, not focused.
func New(th theme.Resolved) Model {
	m := Model{char: " ", on: true, id: frames.NextID(), th: &th}
	return m.restyle()
}

// noTheme and noStyles are what a zero Model works with; never written.
var (
	noTheme  theme.Resolved
	noStyles Styles
)

func (m *Model) resolved() *theme.Resolved {
	if m.th == nil {
		return &noTheme
	}
	return m.th
}

func (m *Model) styles() *Styles {
	if m.st == nil {
		return &noStyles
	}
	return m.st
}

// restyle derives the styles from the theme and applies the app's edit.
func (m Model) restyle() Model {
	st := DefaultStyles(*m.resolved())
	if m.edit != nil {
		st = m.edit(st)
	}
	m.st = &st
	return m
}

// WithStyles changes the styles derived from the theme through edit, now and
// after every SetTheme. It replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m.restyle()
}

// SetTheme repaints the cursor and sets its pace after the theme changed.
func (m Model) SetTheme(th theme.Resolved) Model {
	m.th = &th
	return m.restyle()
}

// SetChar is the character under the cursor, as the caret moves over text.
func (m Model) SetChar(s string) Model {
	m.char = s
	return m
}

// SetMode changes how a live cursor shows, lit at once; turning Blinking on a
// focused cursor starts the blink.
func (m Model) SetMode(mode Mode) (Model, tea.Cmd) {
	m.mode = mode
	return m.Reset()
}

// Mode is how the cursor shows when focused.
func (m Model) Mode() Mode { return m.mode }

// Focused reports whether the cursor shows at all.
func (m Model) Focused() bool { return m.focused }

// Visible reports whether View paints the cursor now: focused, not Hidden,
// and not in the dark half of a blink.
func (m Model) Visible() bool { return m.focused && m.mode != Hidden && m.on }

// ID tells this cursor's BlinkMsg from another's.
func (m Model) ID() int64 { return m.id }

// Focus shows the cursor and, when Blinking, starts its blink.
func (m Model) Focus() (Model, tea.Cmd) {
	m.focused = true
	return m.Reset()
}

// Blur hides the cursor; a blink already on its way lands spent.
func (m Model) Blur() Model {
	m.focused, m.on = false, true
	m.tag++
	return m
}

// Reset lights the cursor for a full blink, as after an edit: a caret that
// went dark while typing would seem to lag.
func (m Model) Reset() (Model, tea.Cmd) {
	m.on = true
	m.tag++
	if !m.blinks() {
		return m, nil
	}
	return m, m.tick()
}

// Blink is the command that keeps a focused cursor blinking, for a host that
// focused it before it could hand on a command; nil unless it blinks.
func (m Model) Blink() tea.Cmd {
	if !m.blinks() {
		return nil
	}
	return m.tick()
}

func (m Model) blinks() bool { return m.focused && m.mode == Blinking }

func (m Model) tick() tea.Cmd {
	msg := BlinkMsg{ID: m.id, tag: m.tag}
	return tea.Tick(edit.Blink(m.resolved().CaretBlink), func(time.Time) tea.Msg { return msg })
}

// Update flips the cursor on its own BlinkMsg and arms the next one. Any
// other message leaves it as it was, with no command.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	b, ok := msg.(BlinkMsg)
	if !ok || b.ID != m.id {
		return m, bezel.Result{}
	}
	if !m.blinks() || b.tag != m.tag {
		return m, bezel.Result{Handled: true}
	}
	m.on = !m.on
	m.tag++
	return m, bezel.Result{Handled: true, Cmd: m.tick()}
}

// View is the character, painted as the cursor while Visible.
func (m Model) View() string {
	if m.Visible() {
		return m.styles().Cursor.Render(m.char)
	}
	return m.styles().Text.Render(m.char)
}
