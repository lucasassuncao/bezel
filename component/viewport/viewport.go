// Package viewport is scrollable text the app sizes: wrapped at its width or
// panned sideways, with a gutter, a style per line and search highlights.
package viewport

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	defaultStep  = 6 // cells a pan moves
	defaultWheel = 3 // lines a wheel notch scrolls
	tabStop      = 4 // a tab runs to the next multiple of this many cells
)

// Keys are the bindings that scroll and pan. Down, Right, PageDown and the
// rest have no help of their own: Up's, Left's and PageUp's name them.
type Keys struct {
	Up, Down, Left, Right                      key.Binding
	PageUp, PageDown, HalfPageUp, HalfPageDown key.Binding
	Top, Bottom                                key.Binding
}

// DefaultKeys are the arrows, pgup/pgdown and home/end; half a page has no
// key until an app gives it one.
func DefaultKeys() Keys {
	return Keys{
		Up:       key.New("up").WithHelp("↑/↓", "scroll"),
		Down:     key.New("down"),
		Left:     key.New("left").WithHelp("←/→", "pan"),
		Right:    key.New("right"),
		PageUp:   key.New("pgup").WithHelp("pgup/pgdn", "page"),
		PageDown: key.New("pgdown"),
		Top:      key.New("home"),
		Bottom:   key.New("end"),
	}
}

// Styles are how search matches are painted.
type Styles struct {
	Match   lipgloss.Style // every match
	Current lipgloss.Style // the match the reader is on
}

// DefaultStyles paints matches in the theme's Match and MatchCurrent.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Match: th.Match, Current: th.MatchCurrent}
}

// Model is a value: every method returns the new viewport.
type Model struct {
	lines         []string
	idx           index
	width, height int
	wrap          bool
	y, x          int // as set; every read clamps them, so a resize loses no place
	step, wheel   int
	keys          Keys
	edit          func(Styles) Styles
	gutter        func(GutterRow) string
	gutterWidth   int // f(GutterRow{})'s cells, which every row's gutter keeps
	lineStyle     func(line int) lipgloss.Style
	fill          bool
	hl            highlights
	cur           int // the current match; -1 for none
}

// GutterRow is what a gutter function is asked to draw for one row.
type GutterRow struct {
	Line      int  // the content line, from 0; Lines or more on a row past the end
	Lines     int  // how many lines the content has
	Continued bool // a wrapped line's second row or later
}

// WithGutter draws f's text before every row, kept in place while panning:
// line numbers, marks. Every row's must be as wide as f(GutterRow{}); nil
// removes it.
func (m Model) WithGutter(f func(GutterRow) string) Model {
	return m.keepTop(func(m Model) Model {
		m.gutter, m.gutterWidth = f, 0
		if f != nil {
			m.gutterWidth = ansi.StringWidth(f(GutterRow{}))
		}
		return m
	})
}

// WithLineStyle paints each whole line in f's style for it, before the line
// is cut or wrapped.
func (m Model) WithLineStyle(f func(line int) lipgloss.Style) Model {
	m.lineStyle = f
	return m
}

// WithFill gives the rows past the end of the text the gutter too, as a
// column of "~" does; without it they are blank.
func (m Model) WithFill(on bool) Model {
	m.fill = on
	return m
}

// New is an empty viewport, 0 x 0 until sized, that pans six cells and
// scrolls three lines a wheel notch.
func New() Model {
	return Model{idx: newIndex(nil), step: defaultStep, wheel: defaultWheel, keys: DefaultKeys()}
}

// WithStep pans cells at a time; below 0 is 0, which stops the keys panning.
func (m Model) WithStep(cells int) Model {
	m.step = max(cells, 0)
	return m
}

// WithWheel scrolls lines a wheel notch; 0 or less turns the wheel off.
func (m Model) WithWheel(lines int) Model {
	m.wheel = max(lines, 0)
	return m
}

// WithKeys rebinds the scrolling keys.
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
	s := DefaultStyles(th)
	if m.edit != nil {
		s = m.edit(s)
	}
	return s
}

// SetContent shows s, one line per "\n"; see SetLines.
func (m Model) SetContent(s string) Model { return m.SetLines(strings.Split(s, "\n")) }

// SetLines shows lines, copied. CRLF and lone CRs break lines too, and a tab
// becomes spaces to the next multiple of four cells, which a width can count.
func (m Model) SetLines(lines []string) Model {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.ReplaceAll(strings.ReplaceAll(l, "\r\n", "\n"), "\r", "\n")
		for _, part := range strings.Split(l, "\n") {
			out = append(out, clean(expandTabs(part)))
		}
	}
	m.lines = out
	m.idx = newIndex(out)
	m.hl, m.cur = highlights{}, -1
	if m.wrap {
		m.idx = m.idx.wrapped(m.textWidth())
	}
	m.y = min(max(m.y, 0), m.last())
	m.x = min(max(m.x, 0), m.maxX())
	return m
}

// expandTabs turns each tab into the spaces up to the next tab stop,
// counting the cells before it, escape sequences left out.
func expandTabs(s string) string {
	if !strings.Contains(s, "\t") {
		return s
	}
	parts := strings.Split(s, "\t")
	var b strings.Builder
	col := 0
	for i, p := range parts {
		b.WriteString(p)
		if i == len(parts)-1 {
			break
		}
		col += ansi.StringWidth(p)
		pad := tabStop - col%tabStop
		b.WriteString(strings.Repeat(" ", pad))
		col += pad
	}
	return b.String()
}

// clean drops a control character standing alone and an escape sequence the
// line cuts off, which would swallow what follows; bad UTF-8 becomes U+FFFD.
func clean(s string) string {
	if isNarrow(s) {
		return s
	}
	s = strings.ToValidUTF8(s, "�")
	var b strings.Builder
	var state byte
	for len(s) > 0 {
		seq, _, n, next := ansi.DecodeSequence(s, state, nil)
		r, size := utf8.DecodeRuneInString(seq)
		if lone := size == len(seq) && unicode.IsControl(r); !lone && next == byte(ansi.NormalState) && whole(seq) {
			b.WriteString(seq)
		}
		s, state = s[n:], next
	}
	return b.String()
}

// whole reports whether seq, decoded on its own, ends complete: the decoder
// hands back an escape cut short by a control character as if it were one.
func whole(seq string) bool {
	if !strings.HasPrefix(seq, "\x1b") {
		return true
	}
	_, _, n, st := ansi.DecodeSequence(seq, byte(ansi.NormalState), nil)
	return n == len(seq) && st == byte(ansi.NormalState)
}

// SetWidth is the viewport's width in cells, the gutter included; the line
// at the top stays there.
func (m Model) SetWidth(w int) Model {
	return m.keepTop(func(m Model) Model {
		m.width = w
		return m
	})
}

// SetHeight is how many rows the viewport shows.
func (m Model) SetHeight(h int) Model {
	m.height = h
	return m
}

// SetWrap folds long lines into rows, or cuts them for panning; the line at
// the top stays there.
func (m Model) SetWrap(on bool) Model {
	return m.keepTop(func(m Model) Model {
		m.wrap = on
		if on {
			m.x = 0
		}
		return m
	})
}

// keepTop applies change and rewraps, keeping the line at the top in place.
func (m Model) keepTop(change func(Model) Model) Model {
	line, _ := m.idx.at(m.YOffset())
	m = change(m)
	if m.wrap {
		m.idx = m.idx.wrapped(m.textWidth())
	} else {
		m.idx = m.idx.unwrapped()
	}
	m.y = m.idx.row(line)
	return m
}

// textWidth is the cells a row has for text, after the gutter.
func (m Model) textWidth() int { return max(0, m.width-m.gutterWidth) }

// last is the offset that shows the last row at the bottom.
func (m Model) last() int { return max(0, m.idx.total()-max(m.height, 0)) }

// maxX is how far a pan can go: the longest line's end at the right edge.
func (m Model) maxX() int {
	if m.wrap {
		return 0
	}
	return max(0, m.idx.longest-m.textWidth())
}

// SetYOffset makes row n the top, kept within the text.
func (m Model) SetYOffset(n int) Model {
	m.y = min(max(n, 0), m.last())
	return m
}

// SetXOffset pans to cell n, kept within the longest line; nothing while
// wrapping.
func (m Model) SetXOffset(n int) Model {
	m.x = min(max(n, 0), m.maxX())
	return m
}

// ScrollDown moves n rows down, stopping at the last; the current match
// becomes the first at or below the new top.
func (m Model) ScrollDown(n int) Model { return m.SetYOffset(m.YOffset() + n).renearest() }

// ScrollUp moves n rows up, stopping at the first; the current match
// becomes the first at or below the new top.
func (m Model) ScrollUp(n int) Model { return m.SetYOffset(m.YOffset() - n).renearest() }

// ScrollRight pans n cells right.
func (m Model) ScrollRight(n int) Model { return m.SetXOffset(m.XOffset() + n) }

// ScrollLeft pans n cells left.
func (m Model) ScrollLeft(n int) Model { return m.SetXOffset(m.XOffset() - n) }

// PageDown moves a page down.
func (m Model) PageDown() Model { return m.ScrollDown(m.height) }

// PageUp moves a page up.
func (m Model) PageUp() Model { return m.ScrollUp(m.height) }

// HalfPageDown moves half a page down.
func (m Model) HalfPageDown() Model { return m.ScrollDown(m.height / 2) }

// HalfPageUp moves half a page up.
func (m Model) HalfPageUp() Model { return m.ScrollUp(m.height / 2) }

// ToTop shows the first row, and makes the first match current.
func (m Model) ToTop() Model { return m.SetYOffset(0).renearest() }

// ToBottom shows the last row at the bottom; the current match becomes the
// first at or below the new top.
func (m Model) ToBottom() Model { return m.SetYOffset(m.last()).renearest() }

// EnsureVisible scrolls line to the top unless it shows already, and pans to
// a step left of colStart unless colEnd fits from the left edge. Before the
// viewport has a size nothing shows, so nothing moves.
func (m Model) EnsureVisible(line, colStart, colEnd int) Model {
	if m.width <= 0 || m.height <= 0 {
		return m
	}
	if colEnd <= m.textWidth() {
		m = m.SetXOffset(0)
	} else {
		m = m.SetXOffset(colStart - m.step)
	}
	if r, top := m.idx.row(line), m.YOffset(); r < top || r >= top+m.height {
		m = m.SetYOffset(r)
	}
	return m
}

// Content is the text, its lines joined by "\n".
func (m Model) Content() string { return strings.Join(m.lines, "\n") }

// Lines is a copy of the text's lines.
func (m Model) Lines() []string { return slices.Clone(m.lines) }

// Width is the viewport's width in cells.
func (m Model) Width() int { return m.width }

// Height is how many rows the viewport shows.
func (m Model) Height() int { return m.height }

// Wrap reports whether long lines fold into rows.
func (m Model) Wrap() bool { return m.wrap }

// YOffset is the first row shown.
func (m Model) YOffset() int { return min(max(m.y, 0), m.last()) }

// XOffset is the first cell of each line shown.
func (m Model) XOffset() int { return min(max(m.x, 0), m.maxX()) }

// AtTop reports whether the first row shows.
func (m Model) AtTop() bool { return m.YOffset() <= 0 }

// AtBottom reports whether the last row shows at the bottom.
func (m Model) AtBottom() bool { return m.YOffset() >= m.last() }

// PastBottom reports an offset set beyond the last row, as a resize leaves
// it; reads clamp it, so the first key back moves at once.
func (m Model) PastBottom() bool { return m.y > m.last() }

// ScrollPercent is how far down the view is, from 0 to 1; 1 when every row
// shows.
func (m Model) ScrollPercent() float64 {
	total := m.idx.total()
	if m.height >= total {
		return 1
	}
	return min(max(float64(m.YOffset())/float64(total-m.height), 0), 1)
}

// HorizontalScrollPercent is how far right the pan is, from 0 to 1.
func (m Model) HorizontalScrollPercent() float64 {
	span := m.idx.longest - m.textWidth()
	if m.XOffset() >= span {
		return 1
	}
	return min(max(float64(m.XOffset())/float64(span), 0), 1)
}

// TotalRows is how many rows the text takes, wrapped rows counted.
func (m Model) TotalRows() int { return m.idx.total() }

// VisibleRows is how many rows of text show, the blank ones after it left out.
func (m Model) VisibleRows() int {
	return max(0, min(m.height, m.idx.total()-m.YOffset()))
}

// Scroll is where the view sits in the text, for a pane's scrollbar.
func (m Model) Scroll() layout.Scroll {
	return layout.Scroll{Offset: m.YOffset(), Visible: m.VisibleRows(), Total: m.TotalRows()}
}

// Update scrolls and pans on its keys and the mouse wheel; the command is
// always nil.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if len(m.lines) > 0 {
			m = m.key(msg)
		}
	case tea.MouseWheelMsg:
		if m.wheel > 0 {
			m = m.wheelMoved(msg)
		}
	}
	return m, nil
}

func (m Model) key(k tea.KeyPressMsg) Model {
	km := m.keys
	switch {
	case key.Matches(k, km.Up):
		return m.ScrollUp(1)
	case key.Matches(k, km.Down):
		return m.ScrollDown(1)
	case key.Matches(k, km.Left):
		return m.ScrollLeft(m.step)
	case key.Matches(k, km.Right):
		return m.ScrollRight(m.step)
	case key.Matches(k, km.PageUp):
		return m.PageUp()
	case key.Matches(k, km.PageDown):
		return m.PageDown()
	case key.Matches(k, km.HalfPageUp):
		return m.HalfPageUp()
	case key.Matches(k, km.HalfPageDown):
		return m.HalfPageDown()
	case key.Matches(k, km.Top):
		return m.ToTop()
	case key.Matches(k, km.Bottom):
		return m.ToBottom()
	}
	return m
}

// wheelMoved scrolls a notch, or pans with shift held or a sideways wheel.
func (m Model) wheelMoved(w tea.MouseWheelMsg) Model {
	shift := w.Mod.Contains(tea.ModShift)
	switch w.Button {
	case tea.MouseWheelDown:
		if shift {
			return m.ScrollRight(m.step)
		}
		return m.ScrollDown(m.wheel)
	case tea.MouseWheelUp:
		if shift {
			return m.ScrollLeft(m.step)
		}
		return m.ScrollUp(m.wheel)
	case tea.MouseWheelRight:
		return m.ScrollRight(m.step)
	case tea.MouseWheelLeft:
		return m.ScrollLeft(m.step)
	}
	return m
}

// Legend lists the scrolling keys; panning drops out while wrapping.
func (m Model) Legend() []legend.Entry {
	km := m.keys
	bs := []key.Binding{km.Up, km.Down}
	if !m.wrap {
		bs = append(bs, km.Left, km.Right)
	}
	bs = append(bs, km.PageUp, km.PageDown, km.HalfPageUp, km.HalfPageDown, km.Top, km.Bottom)
	return legend.FromBindings(bs)
}

// View is the rows from the offset, exactly Width cells by Height rows; ""
// before both are above zero.
func (m Model) View(th theme.Resolved) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	tw, n := m.textWidth(), len(m.lines)
	var st Styles
	if len(m.hl.first) > 0 {
		st = m.styles(th)
	}
	out := make([]string, 0, m.height)
	line, part := m.idx.at(m.YOffset())
	for len(out) < m.height && line < n {
		text := m.lines[line]
		if m.lineStyle != nil {
			text = m.lineStyle(line).Render(text)
		}
		text = m.paint(text, line, st)
		if !m.wrap {
			out = append(out, m.gutterFor(line, false)+fit(m.cut(text, m.XOffset(), tw), tw))
			line++
			continue
		}
		br := breaks(m.lines[line], tw)
		for ; part < len(br) && len(out) < m.height; part++ {
			end := br[part] + tw
			if part+1 < len(br) {
				end = br[part+1]
			}
			out = append(out, m.gutterFor(line, part > 0)+fit(ansi.Cut(text, br[part], end), tw))
		}
		line, part = line+1, 0
	}
	for row := n; len(out) < m.height; row++ {
		past := ""
		if m.fill {
			past = m.gutterFor(row, false)
		}
		out = append(out, past)
	}
	for i, row := range out {
		out[i] = fit(row, m.width)
	}
	return strings.Join(out, "\n")
}

// paint styles line's matches in one pass, the current one apart.
func (m Model) paint(text string, line int, st Styles) string {
	spans := m.hl.byLine[line]
	if len(spans) == 0 {
		return text
	}
	ranges := make([]lipgloss.Range, len(spans))
	for i, s := range spans {
		style := st.Match
		if s.match == m.cur {
			style = st.Current
		}
		ranges[i] = lipgloss.NewRange(s.start, s.end, style)
	}
	return lipgloss.StyleRanges(text, ranges...)
}

// gutterFor is the gutter of a row of line, "" without one.
func (m Model) gutterFor(line int, continued bool) string {
	if m.gutter == nil {
		return ""
	}
	return m.gutter(GutterRow{Line: line, Lines: len(m.lines), Continued: continued})
}

// cut is the cells [from, from+width) of text; the whole line when it fits.
func (m Model) cut(text string, from, width int) string {
	if from == 0 && ansi.StringWidth(text) <= width {
		return text
	}
	return ansi.Cut(text, from, from+width)
}

// fit is s exactly width cells: cut where it runs over, padded where short.
func fit(s string, width int) string {
	w := ansi.StringWidth(s)
	if w > width {
		s = ansi.Truncate(s, width, "")
		w = ansi.StringWidth(s)
	}
	return s + strings.Repeat(" ", max(0, width-w))
}
