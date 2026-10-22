// Package textbox is a multi-line text field: its lines numbered in draw's
// gutter, a long one wrapped at the field's width, the caret moved by the
// readline keys, and no cap on what it holds.
package textbox

import (
	"slices"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/clip"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/theme"
)

// Keys are the editing bindings. Any other key that types text inserts it.
type Keys struct {
	Left, Right, WordLeft, WordRight   key.Binding
	Up, Down, PageUp, PageDown         key.Binding
	LineStart, LineEnd, Top, Bottom    key.Binding
	Newline, DeleteBack, DeleteForward key.Binding
	DeleteWordBack, DeleteWordForward  key.Binding
	KillToStart, KillToEnd             key.Binding
	Paste                              key.Binding // reads the clipboard
}

// DefaultKeys are the arrows, pgup/pgdown, ctrl+home/end for the ends of the
// text, and the readline keys an input answers to, ctrl+v to paste.
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
	}
}

// ids tells fields apart, so a blink reaches only the field that armed it.
var ids atomic.Int64

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
	id, tag   int64
	keys      Keys
	clipboard func() (string, error) // nil is the system's
	th        theme.Resolved
}

// New is an empty field, not focused, sized by its text until SetWidth and
// SetHeight say otherwise.
func New(th theme.Resolved) Model {
	return Model{lines: []line{{starts: []int{0}}}, goal: -1, id: ids.Add(1), keys: DefaultKeys(), th: th, caretOn: true}
}

// wrap is l with its rows worked out at the field's width.
func (m Model) wrap(l edit.Line) line { return line{Line: l, starts: segments(l, m.textWidth())} }

// WithClipboard reads the clipboard through read on a paste key, for a test
// that must not touch the machine's own; nil is the system's.
func (m Model) WithClipboard(read func() (string, error)) Model {
	m.clipboard = read
	return m
}

// WithKeys rebinds the editing keys.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// SetTheme repaints the field after the terminal's background was learned.
func (m Model) SetTheme(th theme.Resolved) Model {
	m.th = th
	return m
}

// WithValue is the field holding s, the caret at its end. CRLF and lone CRs
// read as line breaks, so a Windows file comes back one line per line.
func (m Model) WithValue(s string) Model {
	parts := strings.Split(clean(s), "\n")
	m.lines = make([]line, len(parts))
	for i, p := range parts {
		m.lines[i] = m.wrap(edit.NewLine(p))
	}
	m.row, m.goal = len(m.lines)-1, -1
	return m.scroll()
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
func (m Model) LineCount() int { return len(m.lines) }

// Line is the line the caret is on, from 0.
func (m Model) Line() int { return m.row }

// Column is how many characters sit before the caret on its line.
func (m Model) Column() int { return m.lines[m.row].Cursor() }

// MoveTo puts the caret on line at column col, both kept within the text,
// and scrolls it into view.
func (m Model) MoveTo(line, col int) Model {
	m.row = min(max(line, 0), len(m.lines)-1)
	m = m.setLine(m.lines[m.row].SetCursor(col))
	m.goal = -1
	return m.scroll()
}

// SetWidth is the field's width in cells, the gutter included; zero draws
// each line whole.
func (m Model) SetWidth(w int) Model {
	if w = max(w, 0); w != m.width {
		m.width = w
		lines := make([]line, len(m.lines))
		for i, l := range m.lines {
			lines[i] = m.wrap(l.Line)
		}
		m.lines = lines
	}
	return m.scroll()
}

// SetHeight is how many rows the field shows; zero shows every one.
func (m Model) SetHeight(h int) Model {
	m.height = max(h, 0)
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
	if !m.focused {
		return nil
	}
	return m.tick()
}

// Update edits on a key or a paste while focused, and flips the caret on its
// own BlinkMsg. A key it has no use for leaves it as it was, with no command.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BlinkMsg:
		if !m.focused || msg.id != m.id || msg.tag != m.tag {
			return m, nil
		}
		m.caretOn = !m.caretOn
		return m.arm()
	case tea.PasteMsg:
		if m.focused {
			return m.insert(msg.Content).edited()
		}
	case tea.KeyPressMsg:
		if m.focused {
			return m.key(msg)
		}
	}
	return m, nil
}

// typingMods are the modifiers that make a key a command, never text.
const typingMods = tea.ModCtrl | tea.ModAlt | tea.ModMeta | tea.ModSuper | tea.ModHyper

func (m Model) key(k tea.KeyPressMsg) (Model, tea.Cmd) {
	l, keys := m.lines[m.row], m.keys
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

// edited keeps the caret in view and lights it for a full blink: a caret
// that went dark while typing would seem to lag.
func (m Model) edited() (Model, tea.Cmd) { return m.scroll().lit() }

func (m Model) lit() (Model, tea.Cmd) {
	m.caretOn = true
	return m.arm()
}

// arm schedules the next flip under a new tag, so every tick already on its
// way, this one's twin included, lands spent.
func (m Model) arm() (Model, tea.Cmd) {
	m.tag++
	return m, m.tick()
}

func (m Model) tick() tea.Cmd {
	msg := BlinkMsg{id: m.id, tag: m.tag}
	return tea.Tick(draw.Blink(m.th.CaretBlink), func(time.Time) tea.Msg { return msg })
}

// clean folds every line break to "\n", turns tabs into spaces and drops
// what would draw nothing, line by line.
func clean(s string) string {
	s = strings.NewReplacer("\r\n", "\n", "\r", "\n", "\t", strings.Repeat(" ", tabWidth)).Replace(s)
	parts := strings.Split(s, "\n")
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
	parts := strings.Split(clean(s), "\n")
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
func (m Model) textWidth() int {
	if m.width == 0 {
		return 0
	}
	return max(m.width-draw.GutterWidth-1, 1)
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
		w, space := ansi.StringWidth(l.At(i)), isSpace(l.At(i))
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
	r := []rune(g)
	return len(r) > 0 && unicode.IsSpace(r[0])
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
func (m Model) Rows() int { return m.rowCount() }

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
	if m.height == 0 {
		m.top = 0
		return m
	}
	r := m.caretRow()
	m.top = min(m.top, r)
	if r >= m.top+m.height {
		m.top = r - m.height + 1
	}
	m.top = max(min(m.top, m.rowCount()-m.height), 0)
	return m
}

// View is the rows in the window, each a gutter and a wrapped piece of a
// line, exactly the width and height set; rows past the text show "~".
func (m Model) View() string {
	caret := m.focused && m.caretOn
	text := lipgloss.NewStyle()
	if !m.focused {
		text = m.th.Dim
	}
	start, end := m.window()
	var rows []string
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
				gutter := draw.Gutter(i+1, m.th.Muted)
				if k > 0 {
					gutter = m.th.Muted.Render(strings.Repeat(" ", draw.GutterWidth-2) + "│ ")
				}
				onCaret := caret && i == m.row && segmentOf(l.starts, l.Cursor()) == k
				rows = append(rows, gutter+m.piece(l.Line, from, to, onCaret, text))
			}
			n++
		}
	}
	for len(rows) < m.height {
		rows = append(rows, m.th.Muted.Render(strings.Repeat(" ", draw.GutterWidth-4)+"~ │ ")+m.pad(""))
	}
	return strings.Join(rows, "\n")
}

// piece is the graphemes [from, to) of l in style, the caret painted on its
// cell when this row holds it, padded to the text's width.
func (m Model) piece(l edit.Line, from, to int, onCaret bool, style lipgloss.Style) string {
	var b strings.Builder
	c := l.Cursor()
	if !onCaret {
		return m.pad(style.Render(l.Slice(from, to)))
	}
	b.WriteString(style.Render(l.Slice(from, c)))
	if c < to {
		b.WriteString(m.th.Caret.Render(l.At(c)))
		b.WriteString(style.Render(l.Slice(c+1, to)))
	} else {
		b.WriteString(m.th.Caret.Render(" "))
	}
	return m.pad(b.String())
}

// pad fills a row of text out to the width set, or leaves it be without one.
func (m Model) pad(s string) string {
	if m.width == 0 {
		return s
	}
	return draw.Fit(s, max(m.width-draw.GutterWidth, 0))
}
