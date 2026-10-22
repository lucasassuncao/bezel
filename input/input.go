// Package input is a one-line text field: the caret moves and kills a shell's
// line editing offers, a placeholder, a mask for secrets, and a window that
// scrolls sideways to keep the caret in view.
package input

import (
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/theme"
)

// Keys are the editing bindings. Any other key that types text inserts it.
type Keys struct {
	Left, Right, WordLeft, WordRight, Home, End key.Binding
	DeleteBack, DeleteForward                   key.Binding
	DeleteWordBack, DeleteWordForward           key.Binding
	KillToStart, KillToEnd                      key.Binding
}

// DefaultKeys are the arrows and the readline keys: ctrl+a/e, alt+b/f,
// ctrl+w, ctrl+k and ctrl+u, with ctrl and alt arrows for words.
func DefaultKeys() Keys {
	return Keys{
		Left:              key.NewBinding(key.WithKeys("left", "ctrl+b")),
		Right:             key.NewBinding(key.WithKeys("right", "ctrl+f")),
		WordLeft:          key.NewBinding(key.WithKeys("alt+left", "ctrl+left", "alt+b")),
		WordRight:         key.NewBinding(key.WithKeys("alt+right", "ctrl+right", "alt+f")),
		Home:              key.NewBinding(key.WithKeys("home", "ctrl+a")),
		End:               key.NewBinding(key.WithKeys("end", "ctrl+e")),
		DeleteBack:        key.NewBinding(key.WithKeys("backspace", "ctrl+h")),
		DeleteForward:     key.NewBinding(key.WithKeys("delete", "ctrl+d")),
		DeleteWordBack:    key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w", "ctrl+backspace")),
		DeleteWordForward: key.NewBinding(key.WithKeys("alt+delete", "alt+d", "ctrl+delete")),
		KillToStart:       key.NewBinding(key.WithKeys("ctrl+u")),
		KillToEnd:         key.NewBinding(key.WithKeys("ctrl+k")),
	}
}

// ids tells inputs apart, so a blink reaches only the input that armed it.
var ids atomic.Int64

// BlinkMsg flips the caret of the input that armed it. A host hands it on
// the way it hands on keys; an input ignores every other input's.
type BlinkMsg struct{ id, tag int64 }

// Model is a value: every method returns the new field.
type Model struct {
	line        edit.Line
	offset      int // the first grapheme shown when the text is wider than width
	width       int
	prompt      string
	placeholder string
	mask        string
	focused     bool
	caretOn     bool  // false for the dark half of a blink
	id, tag     int64 // tag counts the blinks armed; only the newest flips
	keys        Keys
	th          theme.Resolved
}

// New is an empty field, not focused, as wide as its text.
func New(th theme.Resolved) Model {
	return Model{id: ids.Add(1), keys: DefaultKeys(), th: th, caretOn: true}
}

// WithPrompt is drawn before the text in the accent, as ":" or "> ".
func (m Model) WithPrompt(p string) Model {
	m.prompt = p
	return m
}

// WithPlaceholder shows s, muted, while the field is empty.
func (m Model) WithPlaceholder(s string) Model {
	m.placeholder = s
	return m
}

// WithWidth draws the text exactly w cells wide, scrolled to keep the caret
// in view; zero draws it whole.
func (m Model) WithWidth(w int) Model {
	m.width = max(w, 0)
	return m.scroll()
}

// SetWidth changes the width after a resize, keeping the caret in view.
func (m Model) SetWidth(w int) Model { return m.WithWidth(w) }

// SetTheme repaints the field after the terminal's background was learned.
func (m Model) SetTheme(th theme.Resolved) Model {
	m.th = th
	return m
}

// WithMask draws every character as mask, for a password; word moves and
// kills then run to the ends, giving away no word boundary.
func (m Model) WithMask(mask rune) Model {
	m.mask = string(mask)
	return m.scroll()
}

// WithKeys rebinds the editing keys.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// WithValue is the field holding s, the caret at its end, as a form opens on
// a saved value.
func (m Model) WithValue(s string) Model {
	m.line = edit.NewLine(oneLine(s))
	return m.scroll()
}

// SetValue replaces the text after something happened, a completion say,
// with the caret at its end.
func (m Model) SetValue(s string) Model { return m.WithValue(s) }

// Value is the text in the field.
func (m Model) Value() string { return m.line.String() }

// Placeholder is what an empty field shows.
func (m Model) Placeholder() string { return m.placeholder }

// Cursor is how many characters sit before the caret.
func (m Model) Cursor() int { return m.line.Cursor() }

// Focused reports whether the field takes keys and shows its caret.
func (m Model) Focused() bool { return m.focused }

// Focus gives the field the keys and starts its caret blinking.
func (m Model) Focus() (Model, tea.Cmd) {
	m.focused = true
	return m.lit()
}

// Blur takes the keys away and hides the caret.
func (m Model) Blur() Model {
	m.focused = false
	return m
}

// Blink is the command that keeps a focused caret blinking, for a host that
// focused the field before it could hand on a command; nil while blurred.
func (m Model) Blink() tea.Cmd {
	if !m.focused {
		return nil
	}
	msg := BlinkMsg{id: m.id, tag: m.tag}
	return tea.Tick(draw.Blink(m.th.CaretBlink), func(time.Time) tea.Msg { return msg })
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
			return m.edited(m.line.Insert(oneLine(msg.Content)))
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
	l, masked := m.line, m.mask != ""
	switch {
	case key.Matches(k, m.keys.Left):
		l = l.SetCursor(l.Cursor() - 1)
	case key.Matches(k, m.keys.Right):
		l = l.SetCursor(l.Cursor() + 1)
	case key.Matches(k, m.keys.Home), masked && key.Matches(k, m.keys.WordLeft):
		l = l.SetCursor(0)
	case key.Matches(k, m.keys.End), masked && key.Matches(k, m.keys.WordRight):
		l = l.SetCursor(l.Len())
	case key.Matches(k, m.keys.WordLeft):
		l = l.WordLeft()
	case key.Matches(k, m.keys.WordRight):
		l = l.WordRight()
	case key.Matches(k, m.keys.DeleteBack):
		l = l.DeleteBack()
	case key.Matches(k, m.keys.DeleteForward):
		l = l.DeleteForward()
	case key.Matches(k, m.keys.KillToStart), masked && key.Matches(k, m.keys.DeleteWordBack):
		l = l.KillToStart()
	case key.Matches(k, m.keys.KillToEnd), masked && key.Matches(k, m.keys.DeleteWordForward):
		l = l.KillToEnd()
	case key.Matches(k, m.keys.DeleteWordBack):
		l = l.DeleteWordBack()
	case key.Matches(k, m.keys.DeleteWordForward):
		l = l.DeleteWordForward()
	default:
		text := oneLine(k.Key().Text)
		if text == "" || k.Mod&typingMods != 0 {
			return m, nil
		}
		l = l.Insert(text)
	}
	return m.edited(l)
}

// edited takes l, keeps the caret in view, and lights it for a full blink:
// a caret that went dark while typing would seem to lag.
func (m Model) edited(l edit.Line) (Model, tea.Cmd) {
	m.line = l
	m = m.scroll()
	return m.lit()
}

func (m Model) lit() (Model, tea.Cmd) {
	m.caretOn = true
	return m.arm()
}

// arm schedules the next flip under a new tag, so every tick already on its
// way, this one's twin included, lands spent.
func (m Model) arm() (Model, tea.Cmd) {
	m.tag++
	msg := BlinkMsg{id: m.id, tag: m.tag}
	return m, tea.Tick(draw.Blink(m.th.CaretBlink), func(time.Time) tea.Msg { return msg })
}

// oneLine is s fit for one line: a line break or a tab is a space, and what
// would draw nothing goes.
func oneLine(s string) string {
	return edit.Printable(strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "\t", " ").Replace(s))
}

// show is how grapheme i is drawn: itself, or the mask.
func (m Model) show(i int) string {
	if m.mask != "" {
		return m.mask
	}
	return m.line.At(i)
}

// cells is how wide the graphemes [from, to) are drawn.
func (m Model) cells(from, to int) int {
	if m.mask != "" {
		return max(to-from, 0) * ansi.StringWidth(m.mask)
	}
	return m.line.Width(from, to)
}

// caretCells is the width of the cell the caret covers: its grapheme, or one
// past the end.
func (m Model) caretCells() int {
	if c := m.line.Cursor(); c < m.line.Len() {
		return m.cells(c, c+1)
	}
	return 1
}

// scroll moves the window to show the caret, and back left while the text
// from one grapheme earlier still fits, so a deletion never leaves a gap.
func (m Model) scroll() Model {
	if m.width <= 0 {
		m.offset = 0
		return m
	}
	c, n := m.line.Cursor(), m.line.Len()
	m.offset = min(m.offset, c)
	for m.offset < c && m.cells(m.offset, c)+m.caretCells() > m.width {
		m.offset++
	}
	end := 0
	if c == n {
		end = 1 // the caret's own cell
	}
	for m.offset > 0 && m.cells(m.offset-1, n)+end <= m.width {
		m.offset--
	}
	return m
}

// View is the prompt and the text, the caret on the cell it covers, exactly
// the width set (or as wide as the text, and the caret, without one).
func (m Model) View() string {
	prompt := ""
	if m.prompt != "" {
		prompt = m.th.Accent.Render(m.prompt)
	}
	if m.line.Len() == 0 && m.placeholder != "" {
		return prompt + m.placeholderView()
	}
	var b strings.Builder
	used, c, caret := 0, m.line.Cursor(), m.focused && m.caretOn
	for i := m.offset; i < m.line.Len(); i++ {
		g := m.show(i)
		w := ansi.StringWidth(g)
		if m.width > 0 && used+w > m.width {
			break
		}
		if i == c && caret {
			g = m.th.Caret.Render(g)
		}
		b.WriteString(g)
		used += w
	}
	if c == m.line.Len() && m.focused && (m.width == 0 || used < m.width) {
		end := " "
		if caret {
			end = m.th.Caret.Render(end)
		}
		b.WriteString(end)
		used++
	}
	if m.width > used {
		b.WriteString(strings.Repeat(" ", m.width-used))
	}
	return prompt + b.String()
}

// placeholderView is the placeholder, muted, with the caret on its first
// character while focused.
func (m Model) placeholderView() string {
	text := m.placeholder
	if m.width > 0 {
		text = draw.Fit(text, m.width)
	}
	if !m.focused || !m.caretOn {
		return m.th.Muted.Render(text)
	}
	g := edit.Graphemes(text)
	return m.th.Caret.Render(g[0]) + m.th.Muted.Render(strings.Join(g[1:], ""))
}
