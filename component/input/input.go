// Package input is a one-line text field: the caret moves and kills a shell's
// line editing offers, a placeholder, a mask for secrets, and a window that
// scrolls sideways to keep the caret in view.
package input

import (
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/cursor"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/clip"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Keys are the editing bindings. Any other key that types text inserts it.
type Keys struct {
	Left, Right, WordLeft, WordRight, Home, End key.Binding
	DeleteBack, DeleteForward                   key.Binding
	DeleteWordBack, DeleteWordForward           key.Binding
	KillToStart, KillToEnd                      key.Binding
	Paste                                       key.Binding // reads the clipboard
	Accept, NextSuggestion, PrevSuggestion      key.Binding // with suggestions only
}

// DefaultKeys are the arrows and the readline keys: ctrl+a/e, alt+b/f,
// ctrl+w, ctrl+k and ctrl+u, ctrl and alt arrows for words, ctrl+v to paste;
// tab takes a suggestion, ctrl+n and ctrl+p turn through them.
func DefaultKeys() Keys {
	return Keys{
		Left:              key.New("left", "ctrl+b"),
		Right:             key.New("right", "ctrl+f"),
		WordLeft:          key.New("alt+left", "ctrl+left", "alt+b"),
		WordRight:         key.New("alt+right", "ctrl+right", "alt+f"),
		Home:              key.New("home", "ctrl+a"),
		End:               key.New("end", "ctrl+e"),
		DeleteBack:        key.New("backspace", "ctrl+h"),
		DeleteForward:     key.New("delete", "ctrl+d"),
		DeleteWordBack:    key.New("alt+backspace", "ctrl+w", "ctrl+backspace"),
		DeleteWordForward: key.New("alt+delete", "alt+d", "ctrl+delete"),
		KillToStart:       key.New("ctrl+u"),
		KillToEnd:         key.New("ctrl+k"),
		Paste:             key.New("ctrl+v").WithHelp("ctrl+v", "paste"),
		Accept:            key.New("tab").WithHelp("tab", "complete"),
		NextSuggestion:    key.New("ctrl+n").WithHelp("ctrl+n/p", "suggestion"),
		PrevSuggestion:    key.New("ctrl+p"),
	}
}

// BlinkMsg flips the caret of the input that armed it. A host hands it on
// the way it hands on keys; an input ignores every other input's.
type BlinkMsg struct{ id, tag int64 }

var _ bezel.Model[Model] = Model{}

// Model is a value: every method returns the new field.
type Model struct {
	line        edit.Line
	offset      int // the first grapheme shown when the text is wider than width
	width       int
	prompt      string
	placeholder string
	mask        string
	hidden      bool // nothing of the text is drawn
	focused     bool
	caretOn     bool // false for the dark half of a blink
	mode        cursor.Mode
	term        bool  // the terminal's own cursor stands in for the caret
	id, tag     int64 // tag counts the blinks armed; only the newest flips
	keys        Keys
	clipboard   func() (string, error) // nil is the system's
	th          *theme.Resolved        // shared by every copy and never written
	edit        func(Styles) Styles
	st          *Styles // DefaultStyles of th, edited; rebuilt, never written
	suggestions []string
	lower       []string // suggestions lowercased once, for match
	matches     []int    // the suggestions the text starts, in order
	pick        int      // the match shown, an index into matches
	limit       int      // the most characters the field holds; 0 is no cap
	validate    func(string) error
	err         error // validate's answer for the text as it is
}

// Styles are how a field paints its parts.
type Styles struct {
	Prompt      lipgloss.Style
	Caret       lipgloss.Style // the cell under a lit caret
	Placeholder lipgloss.Style
	Suggestion  lipgloss.Style // the rest of a suggestion past the text
}

// DefaultStyles draw the prompt in the accent, the caret in the theme's
// Caret and the placeholder and suggestions muted.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Prompt: th.Accent, Caret: th.Caret, Placeholder: th.Muted, Suggestion: th.Muted}
}

// New is an empty field, not focused, as wide as its text.
func New(th theme.Resolved) Model {
	return Model{id: frames.NextID(), keys: DefaultKeys(), th: &th, caretOn: true}.restyle()
}

// noTheme and noStyles are what a zero Model works with; never written.
var (
	noTheme  theme.Resolved
	noStyles Styles
)

// resolved is the theme, held by pointer so a copy of the Model costs a word,
// not the 29KB a theme.Resolved takes; the zero one for a Model never built.
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

// SetWidth draws the text exactly w cells wide, scrolled to keep the caret
// in view; zero draws it whole.
func (m Model) SetWidth(w int) Model {
	m.width = max(w, 0)
	return m.scroll()
}

// SetTheme repaints the field after the terminal's background was learned.
func (m Model) SetTheme(th theme.Resolved) Model {
	m.th = &th
	return m.restyle()
}

// WithMask draws every character as mask, for a password; word moves and
// kills then run to the ends, giving away no word boundary.
func (m Model) WithMask(mask rune) Model {
	m.mask = string(mask)
	return m.scroll()
}

// WithHidden draws none of the text, not even a mask, so not even its length
// shows; the caret stays on the first cell. Word keys run to the ends.
func (m Model) WithHidden(on bool) Model {
	m.hidden = on
	return m.scroll()
}

// WithCursorMode sets how the caret shows while focused: blinking at the
// theme's pace, steady, or hidden.
func (m Model) WithCursorMode(mode cursor.Mode) Model {
	m.mode = mode
	return m
}

// SetCursorMode changes how a live caret shows, lit at once; a focused field
// turned to Blinking starts its blink with the command returned.
func (m Model) SetCursorMode(mode cursor.Mode) (Model, tea.Cmd) {
	m.mode = mode
	if !m.focused {
		return m, nil
	}
	return m.lit()
}

// CursorMode is how the caret shows while focused.
func (m Model) CursorMode() cursor.Mode { return m.mode }

// shows reports whether the caret is painted now.
func (m *Model) shows() bool { return m.focused && m.caretOn && m.mode != cursor.Hidden && !m.term }

// blinks reports whether the field ticks its own caret.
func (m *Model) blinks() bool { return m.focused && m.mode == cursor.Blinking && !m.term }

// WithTerminalCursor leaves the caret to the terminal's own cursor: View
// paints none, and TerminalCursor says where to put the real one.
func (m Model) WithTerminalCursor(on bool) Model {
	m.term = on
	return m
}

// TerminalCursor is the terminal's cursor for a field made WithTerminalCursor,
// placed within the field's view: add the field's origin before handing it
// to tea.View. nil while blurred, hidden, or painting its own caret.
func (m Model) TerminalCursor() *tea.Cursor {
	if !m.term || !m.focused || m.mode == cursor.Hidden {
		return nil
	}
	x := ansi.StringWidth(m.prompt)
	if !m.hidden && (m.line.Len() > 0 || m.placeholder == "") {
		for i := m.offset; i < m.line.Cursor(); i++ {
			x += m.cell(i)
		}
	}
	c := tea.NewCursor(x, 0)
	c.Blink = m.mode == cursor.Blinking
	return c
}

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

// SetValue is the field holding s, the caret at its end, as a form opens on
// a saved value.
func (m Model) SetValue(s string) Model {
	m.line = edit.NewLine(m.capped(oneLine(s), 0))
	return m.scroll().match().check()
}

// WithCharLimit caps the field at n characters: typing stops there and a
// paste is cut to what fits. Zero or less is no cap.
func (m Model) WithCharLimit(n int) Model {
	m.limit = max(n, 0)
	return m.SetValue(m.Value())
}

// CharLimit is the most characters the field holds; zero is no cap.
func (m Model) CharLimit() int { return m.limit }

// capped is s cut to the characters left under the limit when the field
// already holds have.
func (m *Model) capped(s string, have int) string {
	if m.limit == 0 {
		return s
	}
	g := edit.Graphemes(s)
	return strings.Join(g[:min(len(g), max(m.limit-have, 0))], "")
}

// WithValidate runs fn on the text after every change and keeps its answer
// for Err; the text stays as typed either way. nil checks nothing.
func (m Model) WithValidate(fn func(string) error) Model {
	m.validate = fn
	return m.check()
}

// Err is what the validator said of the text as it is, nil when it took it.
func (m Model) Err() error { return m.err }

func (m Model) check() Model {
	m.err = nil
	if m.validate != nil {
		m.err = m.validate(m.line.String())
	}
	return m
}

// SetSuggestions offers to complete the text with any of list it starts,
// matched case-blind: the rest of one shows muted after the caret.
func (m Model) SetSuggestions(list []string) Model {
	m.suggestions = slices.Clone(list)
	m.lower = make([]string, len(list))
	for i, s := range list {
		m.lower[i] = strings.ToLower(s)
	}
	return m.match()
}

// SetSuggestions replaces the list after something happened, a directory
// read say.

// Suggestions is the whole list the field completes from.
func (m Model) Suggestions() []string { return slices.Clone(m.suggestions) }

// MatchedSuggestions are the suggestions the text starts, in list order.
func (m Model) MatchedSuggestions() []string {
	out := make([]string, len(m.matches))
	for i, j := range m.matches {
		out[i] = m.suggestions[j]
	}
	return out
}

// CurrentSuggestion is the match the field shows, "" when none does.
func (m Model) CurrentSuggestion() string {
	if len(m.matches) == 0 {
		return ""
	}
	return m.suggestions[m.matches[m.pick]]
}

// match finds the suggestions the text starts, keeping the one shown when
// it still matches.
func (m Model) match() Model {
	shown := m.CurrentSuggestion()
	m.matches, m.pick = nil, 0
	v := strings.ToLower(m.line.String())
	if v == "" {
		return m
	}
	for i, s := range m.lower {
		if len(s) > len(v) && strings.HasPrefix(s, v) {
			if m.suggestions[i] == shown {
				m.pick = len(m.matches)
			}
			m.matches = append(m.matches, i)
		}
	}
	return m
}

// ghost is what the current suggestion adds to the text, drawn after it.
func (m *Model) ghost() []string {
	s := m.CurrentSuggestion()
	if s == "" || m.mask != "" || m.line.Cursor() != m.line.Len() {
		return nil
	}
	return edit.Graphemes(s)[m.line.Len():]
}

// SetValue replaces the text after something happened, a completion say,
// with the caret at its end.

// Value is the text in the field.
func (m Model) Value() string { return m.line.String() }

// Placeholder is what an empty field shows.
func (m Model) Placeholder() string { return m.placeholder }

// Cursor is how many characters sit before the caret.
func (m Model) Cursor() int { return m.line.Cursor() }

// Legend is the keys worth a hint: completing while a suggestion shows, and
// pasting. The editing keys are the terminal's usual ones and go unlisted.
func (m Model) Legend() []legend.Entry {
	var bs []key.Binding
	if m.CurrentSuggestion() != "" {
		bs = append(bs, m.keys.Accept, m.keys.NextSuggestion)
	}
	return legend.FromBindings(append(bs, m.keys.Paste))
}

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
	if !m.blinks() {
		return nil
	}
	return m.tick()
}

// tick is the next flip under the current tag.
func (m Model) tick() tea.Cmd {
	msg := BlinkMsg{id: m.id, tag: m.tag}
	return tea.Tick(edit.Blink(m.resolved().CaretBlink), func(time.Time) tea.Msg { return msg })
}

// Update edits on a key or a paste while focused, and flips the caret on its
// own BlinkMsg. A key it has no use for leaves it as it was, with no command.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	next, cmd := m.update(msg)
	return next, bezel.Result{Handled: m.handles(msg), Cmd: cmd}
}

// handles reports whether msg is the field's: its own blink, or while focused
// a paste, a binding or typed text.
func (m Model) handles(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case BlinkMsg:
		return msg.id == m.id
	case tea.PasteMsg:
		return m.focused
	case tea.KeyPressMsg:
		return m.focused && (m.bound(msg) || (oneLine(msg.Key().Text) != "" && msg.Mod&typingMods == 0))
	}
	return false
}

// bound reports whether k fires a binding; the suggestion keys count only
// while there are suggestions, so tab stays the host's otherwise.
func (m Model) bound(k tea.KeyPressMsg) bool {
	km := m.keys
	if key.Matches(k, km.Left, km.Right, km.WordLeft, km.WordRight, km.Home, km.End, km.DeleteBack,
		km.DeleteForward, km.DeleteWordBack, km.DeleteWordForward, km.KillToStart, km.KillToEnd, km.Paste) {
		return true
	}
	return len(m.matches) > 0 && key.Matches(k, km.Accept, km.NextSuggestion, km.PrevSuggestion)
}

func (m Model) update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BlinkMsg:
		if !m.blinks() || msg.id != m.id || msg.tag != m.tag {
			return m, nil
		}
		m.caretOn = !m.caretOn
		return m.arm()
	case tea.PasteMsg:
		if m.focused {
			return m.edited(m.line.Insert(m.capped(oneLine(msg.Content), m.line.Len())))
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
	l, masked := m.line, m.mask != "" || m.hidden
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
	case key.Matches(k, m.keys.Paste):
		return m, clip.Paste(m.clipboard) // it comes back as a tea.PasteMsg
	case len(m.matches) > 0 && key.Matches(k, m.keys.Accept):
		l = edit.NewLine(m.capped(m.CurrentSuggestion(), 0))
	case len(m.matches) > 0 && key.Matches(k, m.keys.NextSuggestion):
		m.pick = (m.pick + 1) % len(m.matches)
		return m.lit()
	case len(m.matches) > 0 && key.Matches(k, m.keys.PrevSuggestion):
		m.pick = (m.pick + len(m.matches) - 1) % len(m.matches)
		return m.lit()
	default:
		text := oneLine(k.Key().Text)
		if text == "" || k.Mod&typingMods != 0 {
			return m, nil
		}
		if text = m.capped(text, l.Len()); text == "" {
			return m, nil
		}
		l = l.Insert(text)
	}
	return m.edited(l)
}

// edited takes l, keeps the caret in view, and lights it for a full blink:
// a caret that went dark while typing would seem to lag.
func (m Model) edited(l edit.Line) (Model, tea.Cmd) {
	changed := l.String() != m.line.String() // a caret move leaves both alone
	m.line = l
	if m = m.scroll(); changed {
		m = m.match().check()
	}
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
	if !m.blinks() {
		return m, nil
	}
	return m, m.tick()
}

// spaced turns line breaks and tabs into spaces; a Replacer is safe to share.
var spaced = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "\t", " ")

// oneLine is s fit for one line: a line break or a tab is a space, and what
// would draw nothing goes.
func oneLine(s string) string { return edit.Printable(spaced.Replace(s)) }

// show is how grapheme i is drawn: itself, or the mask.
func (m *Model) show(i int) string {
	if m.mask != "" {
		return m.mask
	}
	return m.line.At(i)
}

// cell is how wide grapheme i is drawn. The helpers a loop calls take a
// pointer, which stays cheaper than a copy of the whole Model.
func (m *Model) cell(i int) int { return edit.Cells(m.show(i)) }

// caretCells is the width of the cell the caret covers: its grapheme, or one
// past the end.
func (m *Model) caretCells() int {
	if c := m.line.Cursor(); c < m.line.Len() {
		return m.cell(c)
	}
	return 1
}

// fits is the first grapheme from which the ones up to end, and used cells
// more, still fit the width: found walking back from end, never past what
// the width holds, so a long value costs no more than a short one.
func (m *Model) fits(end, used int) int {
	start := end
	for start > 0 && used+m.cell(start-1) <= m.width {
		start--
		used += m.cell(start)
	}
	return start
}

// scroll moves the window to show the caret, and back left while the text
// from one grapheme earlier still fits, so a deletion never leaves a gap.
func (m Model) scroll() Model {
	if m.width <= 0 {
		m.offset = 0
		return m
	}
	c, n := m.line.Cursor(), m.line.Len()
	m.offset = max(min(m.offset, c), m.fits(c, m.caretCells()))
	end := 0
	if c == n {
		end = 1 // the caret's own cell
	}
	m.offset = min(m.offset, m.fits(n, end))
	return m
}

// View is the prompt and the text, the caret on the cell it covers, exactly
// the width set (or as wide as the text, and the caret, without one).
func (m Model) View() string {
	prompt := ""
	if m.prompt != "" {
		prompt = m.styles().Prompt.Render(m.prompt)
	}
	if m.line.Len() == 0 && m.placeholder != "" {
		return prompt + m.placeholderView()
	}
	if m.hidden {
		return prompt + m.hiddenView()
	}
	var b strings.Builder
	used, c, caret := 0, m.line.Cursor(), m.shows()
	for i := m.offset; i < m.line.Len(); i++ {
		g := m.show(i)
		w := ansi.StringWidth(g)
		if m.width > 0 && used+w > m.width {
			break
		}
		if i == c && caret {
			g = m.styles().Caret.Render(g)
		}
		b.WriteString(g)
		used += w
	}
	if c == m.line.Len() && m.focused && (m.width == 0 || used < m.width) {
		used += m.end(&b, used, caret)
	}
	if m.width > used {
		b.WriteString(strings.Repeat(" ", m.width-used))
	}
	return prompt + b.String()
}

// end draws the cell after the text, and the rest of a suggestion from it,
// into what the width leaves; it returns the cells it took.
func (m *Model) end(b *strings.Builder, used int, caret bool) int {
	ghost, rest := m.ghost(), m.styles().Suggestion
	if len(ghost) == 0 {
		ghost, rest = []string{" "}, lipgloss.NewStyle()
	}
	took := 0
	for i, g := range ghost {
		w := edit.Cells(g)
		if m.width > 0 && used+took+w > m.width {
			break
		}
		if i == 0 && caret {
			b.WriteString(m.styles().Caret.Render(g))
		} else {
			b.WriteString(rest.Render(g))
		}
		took += w
	}
	return took
}

// hiddenView is a blank field, the caret on its first cell while focused.
func (m Model) hiddenView() string {
	var b strings.Builder
	used := 0
	if m.focused {
		end := " "
		if m.shows() {
			end = m.styles().Caret.Render(end)
		}
		b.WriteString(end)
		used = 1
	}
	if m.width > used {
		b.WriteString(strings.Repeat(" ", m.width-used))
	}
	return b.String()
}

// placeholderView is the placeholder, muted, with the caret on its first
// character while focused.
func (m Model) placeholderView() string {
	text := m.placeholder
	if m.width > 0 {
		text = draw.Fit(text, m.width)
	}
	if !m.shows() {
		return m.styles().Placeholder.Render(text)
	}
	g := edit.Graphemes(text)
	return m.styles().Caret.Render(g[0]) + m.styles().Placeholder.Render(strings.Join(g[1:], ""))
}
