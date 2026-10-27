// Package textbox is a multi-line text field: its lines numbered in draw's
// gutter, a long one wrapped at the field's width, the caret moved by the
// readline keys, and no cap on what it holds.
package textbox

import (
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/component/cursor"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/clip"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

// Keys are the editing bindings. Any other key that types text inserts it.
type Keys struct {
	Left, Right, WordLeft, WordRight                key.Binding
	Up, Down, PageUp, PageDown                      key.Binding
	LineStart, LineEnd, Top, Bottom                 key.Binding
	Newline, DeleteBack, DeleteForward              key.Binding
	DeleteWordBack, DeleteWordForward               key.Binding
	KillToStart, KillToEnd                          key.Binding
	UpperWord, LowerWord, CapitalizeWord, Transpose key.Binding
	Paste                                           key.Binding // reads the clipboard

	// Each moves as its plain twin does and drags the selection along.
	SelectLeft, SelectRight, SelectWordLeft, SelectWordRight key.Binding
	SelectUp, SelectDown, SelectLineStart, SelectLineEnd     key.Binding
	SelectAll                                                key.Binding
	Copy, Cut                                                key.Binding // the selection, to the clipboard
}

// DefaultKeys are the arrows, pgup/pgdown, ctrl+home/end for the ends of the
// text, and the readline keys an input answers to, ctrl+v to paste; shift
// with a move selects, alt+a selects everything, alt+w copies it and alt+x
// cuts it; alt+u, alt+l and alt+c change a word's case, ctrl+t transposes.
func DefaultKeys() Keys {
	return Keys{
		Left:              key.NewBinding(key.WithKeys("left", "ctrl+b")),
		Right:             key.NewBinding(key.WithKeys("right", "ctrl+f")),
		WordLeft:          key.NewBinding(key.WithKeys("alt+left", "ctrl+left", "alt+b")),
		WordRight:         key.NewBinding(key.WithKeys("alt+right", "ctrl+right", "alt+f")),
		Up:                key.NewBinding(key.WithKeys("up", "ctrl+p")),
		Down:              key.NewBinding(key.WithKeys("down", "ctrl+n")),
		PageUp:            key.NewBinding(key.WithKeys("pgup")),
		PageDown:          key.NewBinding(key.WithKeys("pgdown")),
		LineStart:         key.NewBinding(key.WithKeys("home", "ctrl+a")),
		LineEnd:           key.NewBinding(key.WithKeys("end", "ctrl+e")),
		Top:               key.NewBinding(key.WithKeys("ctrl+home", "alt+<")),
		Bottom:            key.NewBinding(key.WithKeys("ctrl+end", "alt+>")),
		Newline:           key.NewBinding(key.WithKeys("enter", "ctrl+m")),
		DeleteBack:        key.NewBinding(key.WithKeys("backspace", "ctrl+h")),
		DeleteForward:     key.NewBinding(key.WithKeys("delete", "ctrl+d")),
		DeleteWordBack:    key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w", "ctrl+backspace")),
		DeleteWordForward: key.NewBinding(key.WithKeys("alt+delete", "alt+d", "ctrl+delete")),
		KillToStart:       key.NewBinding(key.WithKeys("ctrl+u")),
		KillToEnd:         key.NewBinding(key.WithKeys("ctrl+k")),
		Paste:             key.NewBinding(key.WithKeys("ctrl+v")),
		SelectLeft:        key.NewBinding(key.WithKeys("shift+left")),
		SelectRight:       key.NewBinding(key.WithKeys("shift+right")),
		SelectWordLeft:    key.NewBinding(key.WithKeys("ctrl+shift+left", "alt+shift+left")),
		SelectWordRight:   key.NewBinding(key.WithKeys("ctrl+shift+right", "alt+shift+right")),
		SelectUp:          key.NewBinding(key.WithKeys("shift+up")),
		SelectDown:        key.NewBinding(key.WithKeys("shift+down")),
		SelectLineStart:   key.NewBinding(key.WithKeys("shift+home")),
		SelectLineEnd:     key.NewBinding(key.WithKeys("shift+end")),
		SelectAll:         key.NewBinding(key.WithKeys("alt+a")),
		UpperWord:         key.NewBinding(key.WithKeys("alt+u")),
		LowerWord:         key.NewBinding(key.WithKeys("alt+l")),
		CapitalizeWord:    key.NewBinding(key.WithKeys("alt+c")),
		Transpose:         key.NewBinding(key.WithKeys("ctrl+t")),
		Copy:              key.NewBinding(key.WithKeys("alt+w")),
		Cut:               key.NewBinding(key.WithKeys("alt+x")),
	}
}

// CopiedMsg is how a copy or a cut to the clipboard went: Err is nil when
// Text is on it. The field ignores it; an app shows it on its status row.
type CopiedMsg struct {
	Text string
	Err  error
}

// Position is a place in the text: a line and the characters before it on
// that line, both from 0.
type Position struct{ Line, Column int }

// before reports whether p comes earlier in the text than q.
func (p Position) before(q Position) bool {
	return p.Line < q.Line || p.Line == q.Line && p.Column < q.Column
}

// BlinkMsg flips the caret of the field that armed it. A host hands it on
// the way it hands on keys; a field ignores every other field's.
type BlinkMsg struct{ id, tag int64 }

// tabWidth is the spaces a tab in a paste becomes: a tab has no width a
// terminal agrees on.
const tabWidth = 4

// line is a line of the text and where its wrapped rows start at the field's
// width, kept together so an edit rewraps only the line it changed.
type line struct {
	edit.Line
	starts []int
}

// Model is a value: every method returns the new field.
type Model struct {
	lines     []line // never empty: an empty field is one empty line
	row       int    // the line the caret is on
	goal      int    // the column up and down aim for; -1 is the caret's own
	top       int    // the first wrapped row shown
	width     int
	height    int
	focused   bool
	caretOn   bool
	mode      cursor.Mode
	term      bool // the terminal's own cursor stands in for the caret
	id, tag   int64
	keys      Keys
	clipboard func() (string, error) // nil is the system's
	write     func(string) error     // nil is the system's
	th        *theme.Resolved        // shared by every copy and never written
	highlight func(string) [][]Span
	marks     [][]Span // highlight's answer for painted
	painted   string
	selecting bool // the selection runs from anchor to the caret
	anchor    Position
	dragging  bool   // the left button went down in the field and is held
	charLimit int    // the most characters, line breaks included; 0 is no cap
	maxLines  int    // the most lines; 0 is no cap
	grow      [2]int // the least and most rows a growing field takes; zero is fixed
	noNumbers bool   // no gutter of line numbers
	prompt    func(line int) string
	promptW   int // the cells every prompt is fit to
	holder    string
}

// Span styles bytes [Start, End) of one line of the text.
type Span struct {
	Start, End int
	Style      lipgloss.Style
}

// New is an empty field, not focused, sized by its text until SetWidth and
// SetHeight say otherwise.
func New(th theme.Resolved) Model {
	return Model{lines: []line{{starts: []int{0}}}, goal: -1, id: frames.NextID(), keys: DefaultKeys(), th: &th, caretOn: true}
}

// ready is m, or for the zero Model the empty field New builds, so a field
// never built still edits and draws; every method that reads a line asks.
func (m Model) ready() Model {
	if len(m.lines) > 0 {
		return m
	}
	return New(*m.styles())
}

// noTheme is what a zero Model draws with; it is never written.
var noTheme theme.Resolved

// styles is the theme, held by pointer so a copy of the Model costs a word,
// not the 29KB a theme.Resolved takes; the zero one for a Model never built.
func (m *Model) styles() *theme.Resolved {
	if m.th == nil {
		return &noTheme
	}
	return m.th
}

// wrap is l with its rows worked out at the field's width. The helpers a loop
// calls take a pointer, which stays cheaper than a copy of the whole Model.
func (m *Model) wrap(l edit.Line) line { return line{Line: l, starts: segments(l, m.textWidth())} }

// WithClipboard reads the clipboard through read on a paste key, for a test
// that must not touch the machine's own; nil is the system's.
func (m Model) WithClipboard(read func() (string, error)) Model {
	m = m.ready()
	m.clipboard = read
	return m
}

// WithClipboardWrite writes the clipboard through write on a copy or a cut,
// for a test that must not touch the machine's own; nil is the system's.
func (m Model) WithClipboardWrite(write func(string) error) Model {
	m = m.ready()
	m.write = write
	return m
}

// Copy is the command that puts the selection on the clipboard, answering
// with a CopiedMsg; nil when nothing is selected.
func (m Model) Copy() tea.Cmd {
	if !m.HasSelection() {
		return nil
	}
	text := m.SelectedText()
	return clip.Copy(m.write, text, func(err error) tea.Msg { return CopiedMsg{Text: text, Err: err} })
}

// WithCharLimit caps the text at n characters, a line break counting as one:
// typing stops there and a paste is cut to what fits. Zero is no cap.
func (m Model) WithCharLimit(n int) Model {
	m = m.ready()
	m.charLimit = max(n, 0)
	return m.WithValue(m.Value())
}

// WithMaxLines caps the text at n lines: enter stops making new ones and a
// paste is cut at the last that fits. Zero is no cap.
func (m Model) WithMaxLines(n int) Model {
	m = m.ready()
	m.maxLines = max(n, 0)
	return m.WithValue(m.Value())
}

// CharLimit is the most characters the text holds; zero is no cap.
func (m Model) CharLimit() int { return m.charLimit }

// MaxLines is the most lines the text holds; zero is no cap.
func (m Model) MaxLines() int { return m.maxLines }

// fit is the cleaned s cut to what the limits leave room for beside the text
// already held.
func (m *Model) fit(s string) string {
	if m.charLimit == 0 && m.maxLines == 0 {
		return s
	}
	chars, lines := len(m.lines)-1, len(m.lines)
	for _, l := range m.lines {
		chars += l.Len()
	}
	var b strings.Builder
	for _, g := range edit.Graphemes(s) {
		if m.charLimit > 0 && chars >= m.charLimit || g == "\n" && m.maxLines > 0 && lines >= m.maxLines {
			break
		}
		if g == "\n" {
			lines++
		}
		chars++
		b.WriteString(g)
	}
	return b.String()
}

// WithKeys rebinds the editing keys.
func (m Model) WithKeys(k Keys) Model {
	m = m.ready()
	m.keys = k
	return m
}

// WithHighlight colours the text while focused: fn takes the whole value and
// gives each line its spans. It runs again only when the text changes.
func (m Model) WithHighlight(fn func(value string) [][]Span) Model {
	m = m.ready()
	m.highlight, m.painted = fn, ""
	return m.repaint(true)
}

// repaint asks the highlighter again when the text is no longer what it painted.
func (m Model) repaint(force bool) Model {
	if m.highlight == nil {
		m.marks = nil
		return m
	}
	if v := m.Value(); force || v != m.painted {
		m.marks, m.painted = m.highlight(v), v
	}
	return m
}

// SetTheme repaints the field after the terminal's background was learned.
func (m Model) SetTheme(th theme.Resolved) Model {
	m = m.ready()
	m.th = &th
	return m
}

// WithValue is the field holding s, the caret at its end. CRLF and lone CRs
// read as line breaks, so a Windows file comes back one line per line.
func (m Model) WithValue(s string) Model {
	m = m.ready()
	m.lines = []line{m.wrap(edit.NewLine(""))}
	parts := strings.Split(m.fit(clean(s)), "\n")
	m.lines = make([]line, len(parts))
	for i, p := range parts {
		m.lines[i] = m.wrap(edit.NewLine(p))
	}
	m.row, m.goal, m.selecting = len(m.lines)-1, -1, false
	return m.scroll().repaint(false)
}

// SetValue replaces the text after something happened, a reload say, with
// the caret at its end.
func (m Model) SetValue(s string) Model { return m.WithValue(s) }

// Value is the text, its lines joined by "\n".
func (m Model) Value() string {
	parts := make([]string, len(m.lines))
	for i, l := range m.lines {
		parts[i] = l.String()
	}
	return strings.Join(parts, "\n")
}

// LineCount is how many lines the text has; an empty one has one.
func (m Model) LineCount() int { return len(m.ready().lines) }

// Line is the line the caret is on, from 0.
func (m Model) Line() int { return m.row }

// Column is how many characters sit before the caret on its line.
func (m Model) Column() int { m = m.ready(); return m.lines[m.row].Cursor() }

// MoveTo puts the caret on line at column col, both kept within the text,
// and scrolls it into view.
func (m Model) MoveTo(line, col int) Model {
	m = m.ready()
	m.row = min(max(line, 0), len(m.lines)-1)
	m = m.setLine(m.lines[m.row].SetCursor(col))
	m.goal = -1
	return m.scroll()
}

// SetWidth is the field's width in cells, the gutter included; zero draws
// each line whole.
func (m Model) SetWidth(w int) Model {
	m = m.ready()
	if w = max(w, 0); w != m.width {
		m.width = w
		m = m.rewrap()
	}
	return m.scroll()
}

// rewrap works every line's rows out again, after the room for text changed.
func (m Model) rewrap() Model {
	lines := make([]line, len(m.lines))
	for i, l := range m.lines {
		lines[i] = m.wrap(l.Line)
	}
	m.lines = lines
	return m
}

// WithLineNumbers shows or hides the gutter of line numbers; it shows by
// default.
func (m Model) WithLineNumbers(on bool) Model {
	m = m.ready()
	m.noNumbers = !on
	return m.rewrap().scroll()
}

// WithPrompt draws p before every line, after the line numbers, in the
// accent, as "> ".
func (m Model) WithPrompt(p string) Model {
	return m.WithPromptFunc(ansi.StringWidth(p), func(int) string { return p })
}

// WithPromptFunc draws fn's answer for each line, from 0, before it, fit to
// width cells; a wrapped line's later rows get blanks. nil draws none.
func (m Model) WithPromptFunc(width int, fn func(line int) string) Model {
	m = m.ready()
	m.prompt, m.promptW = fn, max(width, 0)
	if fn == nil {
		m.promptW = 0
	}
	return m.rewrap().scroll()
}

// WithPlaceholder shows s, muted, while the field is empty.
func (m Model) WithPlaceholder(s string) Model {
	m = m.ready()
	m.holder = s
	return m
}

// Placeholder is what an empty field shows.
func (m Model) Placeholder() string { return m.holder }

// lead is the cells before the text on every row: the gutter and the prompt.
func (m *Model) lead() int {
	w := m.promptW
	if !m.noNumbers {
		w += draw.GutterWidth
	}
	return w
}

// margin is what goes before the text on wrapped row k of line i.
func (m *Model) margin(i, k int) string {
	var b strings.Builder
	switch {
	case m.noNumbers:
	case k == 0:
		b.WriteString(draw.Gutter(i+1, m.styles().Muted))
	default:
		b.WriteString(m.styles().Muted.Render(strings.Repeat(" ", draw.GutterWidth-2) + "│ "))
	}
	if m.prompt != nil {
		p := strings.Repeat(" ", m.promptW)
		if k == 0 {
			p = draw.Fit(m.prompt(i), m.promptW)
		}
		b.WriteString(m.styles().Accent.Render(p))
	}
	return b.String()
}

// SetHeight is how many rows the field shows; zero shows every one. It ends
// a dynamic height.
func (m Model) SetHeight(h int) Model {
	m = m.ready()
	m.height, m.grow = max(h, 0), [2]int{}
	return m.scroll()
}

// WithDynamicHeight makes the field as tall as its text, never under least
// rows nor over most, scrolling past that; the height then follows every
// edit. A least under 1 turns it off and SetHeight rules again.
func (m Model) WithDynamicHeight(least, most int) Model {
	m = m.ready()
	m.grow = [2]int{}
	if least >= 1 {
		m.grow = [2]int{least, max(most, least)}
	}
	return m.scroll()
}

// Width is the field's width in cells, the gutter included.
func (m Model) Width() int { return m.width }

// Height is how many rows the field shows.
func (m Model) Height() int { return m.height }

// Focused reports whether the field takes keys and shows its caret.
func (m Model) Focused() bool { return m.focused }

// Focus gives the field the keys and starts its caret blinking.
func (m Model) Focus() (Model, tea.Cmd) {
	m = m.ready()
	m.focused = true
	return m.lit()
}

// Blur takes the keys away, hides the caret and dims the text.
func (m Model) Blur() Model {
	m.focused = false
	return m
}

// Blink is the command that keeps a focused caret blinking, for a host's
// Init; Focus and every edit arm their own.
func (m Model) Blink() tea.Cmd {
	if !m.blinks() {
		return nil
	}
	return m.tick()
}

// blinks reports whether the field ticks its own caret.
func (m *Model) blinks() bool { return m.focused && m.mode == cursor.Blinking && !m.term }

// shows reports whether the caret is painted now.
func (m *Model) shows() bool { return m.focused && m.caretOn && m.mode != cursor.Hidden && !m.term }

// WithCursorMode sets how the caret shows while focused: blinking at the
// theme's pace, steady, or hidden.
func (m Model) WithCursorMode(mode cursor.Mode) Model {
	m = m.ready()
	m.mode = mode
	return m
}

// SetCursorMode changes how a live caret shows, lit at once; a focused field
// turned to Blinking starts its blink with the command returned.
func (m Model) SetCursorMode(mode cursor.Mode) (Model, tea.Cmd) {
	m = m.ready()
	m.mode = mode
	if !m.focused {
		return m, nil
	}
	return m.lit()
}

// CursorMode is how the caret shows while focused.
func (m Model) CursorMode() cursor.Mode { return m.mode }

// WithTerminalCursor leaves the caret to the terminal's own cursor: View
// paints none, and TerminalCursor says where to put the real one.
func (m Model) WithTerminalCursor() Model {
	m = m.ready()
	m.term = true
	return m
}

// TerminalCursor is the terminal's cursor for a field made WithTerminalCursor,
// placed within the field's view: add the field's origin before handing it
// to tea.View. nil while blurred, hidden, or painting its own caret.
func (m Model) TerminalCursor() *tea.Cursor {
	m = m.ready()
	if !m.term || !m.focused || m.mode == cursor.Hidden {
		return nil
	}
	l := m.lines[m.row]
	from := l.starts[segmentOf(l.starts, l.Cursor())]
	start, _ := m.window()
	c := tea.NewCursor(m.lead()+l.Width(from, l.Cursor()), m.caretRow()-start)
	c.Blink = m.mode == cursor.Blinking
	return c
}

// Update edits on a key or a paste while focused, and flips the caret on its
// own BlinkMsg. A key it has no use for leaves it as it was, with no command.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m = m.ready()
	switch msg := msg.(type) {
	case BlinkMsg:
		if !m.blinks() || msg.id != m.id || msg.tag != m.tag {
			return m, nil
		}
		m.caretOn = !m.caretOn
		return m.arm()
	case tea.PasteMsg:
		if m.focused {
			if m.HasSelection() {
				m = m.cut()
			}
			m.selecting = false
			return m.insert(msg.Content).edited()
		}
	case tea.KeyPressMsg:
		if m.focused {
			return m.key(msg)
		}
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg:
		if m.focused {
			return m.mouse(msg.(tea.MouseMsg))
		}
	}
	return m, nil
}

// mouse selects with the left button: a click puts the caret on the cell and
// anchors there, shift+click extends from the old anchor, a drag drags the
// caret and a release ends it. Cells are the field's own, from its top left.
func (m Model) mouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	at := msg.Mouse()
	switch msg.(type) {
	case tea.MouseClickMsg:
		if at.Button != tea.MouseLeft {
			return m, nil
		}
		keep := m.selecting && at.Mod&tea.ModShift != 0
		anchor := m.anchor
		if !keep {
			anchor = m.PositionAt(at.X, at.Y)
		}
		m = m.dragTo(at.X, at.Y)
		m.selecting, m.anchor, m.dragging = true, anchor, true
	case tea.MouseMotionMsg:
		if !m.dragging {
			return m, nil
		}
		m = m.dragTo(at.X, at.Y)
	case tea.MouseReleaseMsg:
		if !m.dragging {
			return m, nil
		}
		m = m.dragTo(at.X, at.Y)
		m.dragging = false
	}
	return m.scroll().lit()
}

// dragTo moves the caret to the cell, keeping the selection as it was.
func (m Model) dragTo(x, y int) Model {
	p, sel, anchor := m.PositionAt(x, y), m.selecting, m.anchor
	m = m.MoveTo(p.Line, p.Column)
	m.selecting, m.anchor = sel, anchor
	return m
}

// PositionAt is the place in the text drawn at cell (x, y) of the field's
// view, counted from its top left; a cell past the text lands on the
// nearest place in it.
func (m Model) PositionAt(x, y int) Position {
	m = m.ready()
	start, _ := m.window()
	row := max(start+y, 0)
	for i, l := range m.lines {
		if row < len(l.starts) {
			from, to := span(l.starts, row, l.Len())
			col := from + edit.NewLine(l.Slice(from, to)).IndexAt(max(x-m.lead(), 0))
			if row+1 < len(l.starts) {
				col = min(col, to-1) // the end of a wrapped row is the next row's start
			}
			return Position{i, col}
		}
		row -= len(l.starts)
	}
	last := len(m.lines) - 1
	return Position{last, m.lines[last].Len()}
}

// typingMods are the modifiers that make a key a command, never text.
const typingMods = tea.ModCtrl | tea.ModAlt | tea.ModMeta | tea.ModSuper | tea.ModHyper

func (m Model) key(k tea.KeyPressMsg) (Model, tea.Cmd) {
	if moved, ok := m.selectKey(k); ok {
		return moved.scroll().lit()
	}
	keys := m.keys
	switch {
	case key.Matches(k, keys.Copy):
		return m, m.Copy()
	case key.Matches(k, keys.Cut) && m.HasSelection():
		cmd := m.Copy()
		m, blink := m.cut().edited()
		return m, tea.Batch(cmd, blink)
	}
	if m.HasSelection() && editKey(k, keys) {
		m = m.cut()
		if !typingKey(k, keys) {
			m.goal = -1
			return m.edited()
		}
	}
	m.selecting = false
	l := m.lines[m.row]
	vertical := false
	switch {
	case key.Matches(k, keys.Up), key.Matches(k, keys.Down):
		m, vertical = m.vertical(key.Matches(k, keys.Down), 1), true
	case key.Matches(k, keys.PageUp), key.Matches(k, keys.PageDown):
		m, vertical = m.vertical(key.Matches(k, keys.PageDown), max(m.height-1, 1)), true
	case key.Matches(k, keys.Top):
		m = m.MoveTo(0, 0)
	case key.Matches(k, keys.Bottom):
		m = m.MoveTo(len(m.lines)-1, m.lines[len(m.lines)-1].Len())
	case key.Matches(k, keys.Left):
		m = m.left()
	case key.Matches(k, keys.Right):
		m = m.right()
	case key.Matches(k, keys.WordLeft):
		m = m.setLine(l.WordLeft())
	case key.Matches(k, keys.WordRight):
		m = m.setLine(l.WordRight())
	case key.Matches(k, keys.LineStart):
		m = m.setLine(l.SetCursor(0))
	case key.Matches(k, keys.LineEnd):
		m = m.setLine(l.SetCursor(l.Len()))
	case key.Matches(k, keys.Newline):
		m = m.insert("\n")
	case key.Matches(k, keys.DeleteBack):
		m = m.deleteBack(edit.Line.DeleteBack)
	case key.Matches(k, keys.DeleteWordBack):
		m = m.deleteBack(edit.Line.DeleteWordBack)
	case key.Matches(k, keys.DeleteForward):
		m = m.deleteForward(edit.Line.DeleteForward)
	case key.Matches(k, keys.DeleteWordForward):
		m = m.deleteForward(edit.Line.DeleteWordForward)
	case key.Matches(k, keys.KillToStart):
		m = m.setLine(l.KillToStart())
	case key.Matches(k, keys.KillToEnd):
		m = m.setLine(l.KillToEnd())
	case wordKey(k, keys) != nil:
		m = m.setLine(wordKey(k, keys)(l.Line))
	case key.Matches(k, keys.Paste):
		return m, clip.Paste(m.clipboard) // it comes back as a tea.PasteMsg
	default:
		text := edit.Printable(k.Key().Text)
		if text == "" || k.Mod&typingMods != 0 {
			return m, nil
		}
		m = m.insert(text)
	}
	if !vertical {
		m.goal = -1
	}
	return m.edited()
}

// selectKey moves the caret for a select key, anchoring the selection where
// the caret was when none is under way.
func (m Model) selectKey(k tea.KeyPressMsg) (Model, bool) {
	keys, l := m.keys, m.lines[m.row]
	if key.Matches(k, keys.SelectAll) {
		last := len(m.lines) - 1
		return m.Select(Position{}, Position{last, m.lines[last].Len()}), true
	}
	var moved Model
	vertical := false
	switch {
	case key.Matches(k, keys.SelectLeft):
		moved = m.left()
	case key.Matches(k, keys.SelectRight):
		moved = m.right()
	case key.Matches(k, keys.SelectWordLeft):
		moved = m.setLine(l.WordLeft())
	case key.Matches(k, keys.SelectWordRight):
		moved = m.setLine(l.WordRight())
	case key.Matches(k, keys.SelectUp), key.Matches(k, keys.SelectDown):
		moved, vertical = m.vertical(key.Matches(k, keys.SelectDown), 1), true
	case key.Matches(k, keys.SelectLineStart):
		moved = m.setLine(l.SetCursor(0))
	case key.Matches(k, keys.SelectLineEnd):
		moved = m.setLine(l.SetCursor(l.Len()))
	default:
		return m, false
	}
	if !m.selecting {
		moved.selecting, moved.anchor = true, m.caret()
	}
	if !vertical {
		moved.goal = -1
	}
	return moved, true
}

// editKey reports whether k changes the text, so a selection goes first.
func editKey(k tea.KeyPressMsg, keys Keys) bool {
	for _, b := range []key.Binding{keys.DeleteBack, keys.DeleteForward, keys.DeleteWordBack,
		keys.DeleteWordForward, keys.KillToStart, keys.KillToEnd} {
		if key.Matches(k, b) {
			return true
		}
	}
	return typingKey(k, keys)
}

// typingKey reports whether k types text, or a line break, in place of the
// selection rather than only deleting it.
func typingKey(k tea.KeyPressMsg, keys Keys) bool {
	if key.Matches(k, keys.Newline) {
		return true
	}
	for _, b := range []key.Binding{keys.Left, keys.Right, keys.WordLeft, keys.WordRight, keys.Up,
		keys.Down, keys.PageUp, keys.PageDown, keys.LineStart, keys.LineEnd, keys.Top, keys.Bottom,
		keys.DeleteBack, keys.DeleteForward, keys.DeleteWordBack, keys.DeleteWordForward,
		keys.KillToStart, keys.KillToEnd, keys.Paste} {
		if key.Matches(k, b) {
			return false
		}
	}
	return edit.Printable(k.Key().Text) != "" && k.Mod&typingMods == 0
}

// caret is where the caret is.
func (m Model) caret() Position { return Position{m.row, m.lines[m.row].Cursor()} }

// HasSelection reports whether some text is selected.
func (m Model) HasSelection() bool {
	return m.selecting && len(m.lines) > 0 && m.anchor != m.caret()
}

// Selection is the selected text's start and end, the end just past it; both
// are the caret when nothing is selected.
func (m Model) Selection() (start, end Position) {
	m = m.ready()
	c := m.caret()
	if !m.HasSelection() {
		return c, c
	}
	if m.anchor.before(c) {
		return m.anchor, c
	}
	return c, m.anchor
}

// SelectedText is the text the selection holds, its lines joined by "\n".
func (m Model) SelectedText() string {
	m = m.ready()
	start, end := m.Selection()
	if start.Line == end.Line {
		return m.lines[start.Line].Slice(start.Column, end.Column)
	}
	parts := []string{m.lines[start.Line].Slice(start.Column, m.lines[start.Line].Len())}
	for _, l := range m.lines[start.Line+1 : end.Line] {
		parts = append(parts, l.String())
	}
	parts = append(parts, m.lines[end.Line].Slice(0, end.Column))
	return strings.Join(parts, "\n")
}

// Select selects from one position to another, both kept within the text;
// the caret lands on to.
func (m Model) Select(from, to Position) Model {
	m = m.MoveTo(from.Line, from.Column)
	anchor := m.caret()
	m = m.MoveTo(to.Line, to.Column)
	m.selecting, m.anchor = true, anchor
	return m
}

// ClearSelection drops the selection, leaving the caret where it is.
func (m Model) ClearSelection() Model {
	m.selecting = false
	return m
}

// cut deletes the selection, leaving the caret where it began.
func (m Model) cut() Model {
	start, end := m.Selection()
	m.selecting = false
	head := m.lines[start.Line].Slice(0, start.Column)
	tail := m.lines[end.Line].Slice(end.Column, m.lines[end.Line].Len())
	joined := edit.NewLine(head + tail).SetCursor(start.Column)
	m.lines = slices.Concat(m.lines[:start.Line], []line{m.wrap(joined)}, m.lines[end.Line+1:])
	m.row = start.Line
	return m
}

// wordKey is the edit a case or transpose key makes, nil for any other key.
func wordKey(k tea.KeyPressMsg, keys Keys) func(edit.Line) edit.Line {
	switch {
	case key.Matches(k, keys.UpperWord):
		return func(l edit.Line) edit.Line { return l.MapWord(strings.ToUpper) }
	case key.Matches(k, keys.LowerWord):
		return func(l edit.Line) edit.Line { return l.MapWord(strings.ToLower) }
	case key.Matches(k, keys.CapitalizeWord):
		return func(l edit.Line) edit.Line { return l.MapWord(capitalize) }
	case key.Matches(k, keys.Transpose):
		return edit.Line.Transpose
	}
	return nil
}

// capitalize is w with its first letter upper case and the rest lower.
func capitalize(w string) string {
	r, n := utf8.DecodeRuneInString(w)
	return string(unicode.ToTitle(r)) + strings.ToLower(w[n:])
}

// InsertString puts s at the caret as typing would, in place of a selection
// and within the limits, and leaves the caret after it.
func (m Model) InsertString(s string) Model {
	m = m.ready()
	if m.HasSelection() {
		m = m.cut()
	}
	m.selecting, m.goal = false, -1
	return m.insert(s).scroll().repaint(false)
}

// InsertRune is InsertString for one character.
func (m Model) InsertRune(r rune) Model { return m.InsertString(string(r)) }

// Word is the word the caret is in or touches, "" between spaces.
func (m Model) Word() string {
	m = m.ready()
	return m.lines[m.row].WordAt()
}

// edited keeps the caret in view and lights it for a full blink: a caret
// that went dark while typing would seem to lag.
func (m Model) edited() (Model, tea.Cmd) { return m.scroll().repaint(false).lit() }

func (m Model) lit() (Model, tea.Cmd) {
	m.caretOn = true
	return m.arm()
}

// arm schedules the next flip under a new tag, so every tick already on its
// way, this one's twin included, lands spent.
func (m Model) arm() (Model, tea.Cmd) {
	m.tag++
	if !m.blinks() {
		return m, nil
	}
	return m, m.tick()
}

func (m Model) tick() tea.Cmd {
	msg := BlinkMsg{id: m.id, tag: m.tag}
	return tea.Tick(draw.Blink(m.styles().CaretBlink), func(time.Time) tea.Msg { return msg })
}

// breaks folds every line break to "\n" and turns tabs into spaces; a
// Replacer is safe to share.
var breaks = strings.NewReplacer("\r\n", "\n", "\r", "\n", "\t", strings.Repeat(" ", tabWidth))

// clean folds every line break to "\n", turns tabs into spaces and drops
// what would draw nothing, line by line.
func clean(s string) string {
	parts := strings.Split(breaks.Replace(s), "\n")
	for i, p := range parts {
		parts[i] = edit.Printable(p)
	}
	return strings.Join(parts, "\n")
}

// setLine replaces the caret's line with l, never touching the slice an
// older copy of the field still reads.
func (m Model) setLine(l edit.Line) Model {
	m.lines = slices.Clone(m.lines)
	m.lines[m.row] = m.wrap(l)
	return m
}

// insert puts s at the caret, its line breaks splitting the line, and leaves
// the caret after it.
func (m Model) insert(s string) Model {
	parts := strings.Split(m.fit(clean(s)), "\n")
	before, after := m.lines[m.row].Split()
	if len(parts) == 1 {
		return m.setLine(m.lines[m.row].Insert(parts[0]))
	}
	added := make([]line, 0, len(parts))
	added = append(added, m.wrap(before.Insert(parts[0])))
	for _, p := range parts[1 : len(parts)-1] {
		added = append(added, m.wrap(edit.NewLine(p)))
	}
	added = append(added, m.wrap(edit.NewLine(parts[len(parts)-1]).Join(after)))
	m.lines = slices.Concat(m.lines[:m.row], added, m.lines[m.row+1:])
	m.row += len(parts) - 1
	return m
}

// deleteBack runs del on the caret's line, or at its start joins it to the
// line above, the caret where they meet.
func (m Model) deleteBack(del func(edit.Line) edit.Line) Model {
	if l := m.lines[m.row]; l.Cursor() > 0 || m.row == 0 {
		return m.setLine(del(l.Line))
	}
	joined := m.lines[m.row-1].Join(m.lines[m.row].Line)
	m.lines = slices.Concat(m.lines[:m.row-1], []line{m.wrap(joined)}, m.lines[m.row+1:])
	m.row--
	return m
}

// deleteForward runs del on the caret's line, or at its end joins the line
// below to it.
func (m Model) deleteForward(del func(edit.Line) edit.Line) Model {
	if l := m.lines[m.row]; l.Cursor() < l.Len() || m.row == len(m.lines)-1 {
		return m.setLine(del(l.Line))
	}
	joined := m.lines[m.row].Join(m.lines[m.row+1].Line)
	m.lines = slices.Concat(m.lines[:m.row], []line{m.wrap(joined)}, m.lines[m.row+2:])
	return m
}

// left moves the caret one character back, to the end of the line above
// from a line's start.
func (m Model) left() Model {
	if l := m.lines[m.row]; l.Cursor() > 0 || m.row == 0 {
		return m.setLine(l.SetCursor(l.Cursor() - 1))
	}
	m.row--
	return m.setLine(m.lines[m.row].SetCursor(m.lines[m.row].Len()))
}

// right moves the caret one character on, to the start of the line below
// from a line's end.
func (m Model) right() Model {
	if l := m.lines[m.row]; l.Cursor() < l.Len() || m.row == len(m.lines)-1 {
		return m.setLine(l.SetCursor(l.Cursor() + 1))
	}
	m.row++
	return m.setLine(m.lines[m.row].SetCursor(0))
}

// textWidth is the cells a row of text gets: the width less the gutter, less
// one cell kept for a caret at the end of a full row.
func (m *Model) textWidth() int {
	if m.width == 0 {
		return 0
	}
	return max(m.width-m.lead()-1, 1)
}

// segments are where the wrapped rows of l start: after the last space that
// lets a row fit, or mid-word when a word is wider than a row.
func segments(l edit.Line, width int) []int {
	starts := []int{0}
	if width <= 0 {
		return starts
	}
	start, used, afterSpace := 0, 0, -1
	for i := range l.Len() {
		w, space := edit.Cells(l.At(i)), isSpace(l.At(i))
		if used+w > width && i > start {
			if space { // it hangs past the edge, in the cell kept for the caret
				starts = append(starts, i+1)
				start, used, afterSpace = i+1, 0, -1
				continue
			}
			brk := i
			if afterSpace > start {
				brk = afterSpace
			}
			starts = append(starts, brk)
			start, used, afterSpace = brk, l.Width(brk, i), -1
		}
		used += w
		if space {
			afterSpace = i + 1
		}
	}
	return starts
}

func isSpace(g string) bool {
	r, _ := utf8.DecodeRuneInString(g)
	return unicode.IsSpace(r)
}

// span is the graphemes [from, to) of segment k of starts on a line of n.
func span(starts []int, k, n int) (from, to int) {
	from, to = starts[k], n
	if k+1 < len(starts) {
		to = starts[k+1]
	}
	return from, to
}

// segmentOf is the segment of starts the caret at c shows on.
func segmentOf(starts []int, c int) int {
	k := 0
	for k+1 < len(starts) && starts[k+1] <= c {
		k++
	}
	return k
}

// caretRow is the wrapped row the caret shows on, counted from the first.
func (m Model) caretRow() int {
	rows := 0
	for _, l := range m.lines[:m.row] {
		rows += len(l.starts)
	}
	l := m.lines[m.row]
	return rows + segmentOf(l.starts, l.Cursor())
}

// Rows is how many rows the text takes at the width set, its wrapped lines
// counted: a form that grows the field sets its height from it.
func (m Model) Rows() int { return m.ready().rowCount() }

// rowCount is how many wrapped rows the text takes.
func (m Model) rowCount() int {
	n := 0
	for _, l := range m.lines {
		n += len(l.starts)
	}
	return n
}

// vertical moves the caret n wrapped rows down (or up), keeping the column
// it aimed for when the move began.
func (m Model) vertical(down bool, n int) Model {
	l := m.lines[m.row]
	k := segmentOf(l.starts, l.Cursor())
	if m.goal < 0 {
		m.goal = l.Width(l.starts[k], l.Cursor())
	}
	row := m.row
	for range n {
		switch {
		case down && k+1 < len(m.lines[row].starts):
			k++
		case down && row+1 < len(m.lines):
			row, k = row+1, 0
		case !down && k > 0:
			k--
		case !down && row > 0:
			row--
			k = len(m.lines[row].starts) - 1
		}
	}
	target := m.lines[row]
	segs := target.starts
	from, to := span(segs, k, target.Len())
	at := from + edit.NewLine(target.Slice(from, to)).IndexAt(m.goal)
	if k+1 < len(segs) {
		at = min(at, to-1) // at to the caret would show on the next row
	}
	m.row = row
	return m.setLine(target.SetCursor(at))
}

// window is the rows [m.top, end) the view shows, end past the text when the
// field is taller than it; every row when no height is set.
func (m Model) window() (start, end int) {
	if m.height == 0 {
		return 0, m.rowCount()
	}
	return m.top, m.top + m.height
}

// scroll moves the window to show the caret's row, and back up while rows
// below the text would show blank.
func (m Model) scroll() Model {
	if m.grow[0] > 0 {
		m.height = min(max(m.rowCount(), m.grow[0]), m.grow[1])
	}
	if m.height == 0 {
		m.top = 0
		return m
	}
	m.top = layout.ClampScroll(m.caretRow(), m.top, m.height)
	m.top = max(min(m.top, m.rowCount()-m.height), 0)
	return m
}

// View is the rows in the window, each a gutter and a wrapped piece of a
// line, exactly the width and height set; rows past the text show "~".
func (m Model) View() string {
	m = m.ready()
	caret := m.shows()
	text := lipgloss.NewStyle()
	if !m.focused {
		text = m.styles().Dim
	}
	start, end := m.window()
	var rows []string
	if len(m.lines) == 1 && m.lines[0].Len() == 0 && m.holder != "" && start == 0 && end > 0 {
		rows, start = []string{m.margin(0, 0) + m.placeholderRow()}, 1
	}
	n := 0
	for i, l := range m.lines {
		if n >= end {
			break
		}
		if n+len(l.starts) <= start { // wholly above the window
			n += len(l.starts)
			continue
		}
		for k := range l.starts {
			if n >= start && n < end {
				from, to := span(l.starts, k, l.Len())
				gutter := m.margin(i, k)
				onCaret := caret && i == m.row && segmentOf(l.starts, l.Cursor()) == k
				var spans []Span
				if m.focused && i < len(m.marks) {
					spans = m.marks[i]
				}
				rows = append(rows, gutter+m.piece(l.Line, spans, from, to, onCaret, m.selectedOn(i), text))
			}
			n++
		}
	}
	filler := strings.Repeat(" ", m.lead())
	if !m.noNumbers {
		filler = m.styles().Muted.Render(strings.Repeat(" ", draw.GutterWidth-4)+"~ │ ") + strings.Repeat(" ", m.promptW)
	}
	for len(rows) < m.height {
		rows = append(rows, filler+m.pad(""))
	}
	return strings.Join(rows, "\n")
}

// placeholderRow is the placeholder, muted, with the caret on its first
// character while it shows, padded to the text's width.
func (m Model) placeholderRow() string {
	text := m.holder
	if m.width > 0 {
		text = draw.Truncate(text, max(m.width-m.lead(), 0))
	}
	if !m.shows() {
		return m.pad(m.styles().Muted.Render(text))
	}
	g := edit.Graphemes(text)
	if len(g) == 0 {
		return m.pad(m.styles().Caret.Render(" "))
	}
	return m.pad(m.styles().Caret.Render(g[0]) + m.styles().Muted.Render(strings.Join(g[1:], "")))
}

// selectedOn is the characters [from, to) of line i the selection covers;
// from equals to when it covers none.
func (m Model) selectedOn(i int) [2]int {
	start, end := m.Selection()
	if !m.HasSelection() || i < start.Line || i > end.Line {
		return [2]int{}
	}
	sel := [2]int{0, m.lines[i].Len()}
	if i == start.Line {
		sel[0] = start.Column
	}
	if i == end.Line {
		sel[1] = end.Column
	}
	return sel
}

// piece is the graphemes [from, to) of l in style, the selection and the
// caret painted over it where this row holds them, padded to the text's width.
func (m *Model) piece(l edit.Line, spans []Span, from, to int, onCaret bool, sel [2]int, style lipgloss.Style) string {
	c := l.Cursor()
	cuts := []int{from, to, sel[0], sel[1]}
	if onCaret {
		cuts = append(cuts, c, c+1)
	}
	for i, at := range cuts {
		cuts[i] = min(max(at, from), to)
	}
	slices.Sort(cuts)
	cuts = slices.Compact(cuts)
	var b strings.Builder
	for i := 0; i+1 < len(cuts); i++ {
		a, z := cuts[i], cuts[i+1]
		switch {
		case onCaret && a == c:
			b.WriteString(m.styles().Caret.Render(l.Slice(a, z)))
		case a >= sel[0] && z <= sel[1]:
			b.WriteString(m.styles().Selected.Render(l.Slice(a, z)))
		default:
			b.WriteString(paint(l, spans, a, z, style))
		}
	}
	if onCaret && c >= to {
		b.WriteString(m.styles().Caret.Render(" "))
	}
	return m.pad(b.String())
}

// paint is the graphemes [from, to) of l, each run in the span holding its
// first byte, the rest in base.
func paint(l edit.Line, spans []Span, from, to int, base lipgloss.Style) string {
	if len(spans) == 0 || from >= to {
		return base.Render(l.Slice(from, to))
	}
	var b strings.Builder
	at := len(l.Slice(0, from))
	run, cur := from, spanAt(spans, at)
	flush := func(end int) {
		st := base
		if cur >= 0 {
			st = spans[cur].Style
		}
		b.WriteString(st.Render(l.Slice(run, end)))
	}
	for i := from; i < to; i++ {
		if s := spanAt(spans, at); s != cur {
			flush(i)
			run, cur = i, s
		}
		at += len(l.At(i))
	}
	flush(to)
	return b.String()
}

// spanAt is the first span holding byte at, -1 for none.
func spanAt(spans []Span, at int) int {
	for i, s := range spans {
		if at >= s.Start && at < s.End {
			return i
		}
	}
	return -1
}

// pad fills a row of text out to the width set, or leaves it be without one.
func (m *Model) pad(s string) string {
	if m.width == 0 {
		return s
	}
	return draw.Fit(s, max(m.width-m.lead(), 0))
}
