package logger

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/viewport"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// DefaultLimit is the lines a Pane keeps until WithLimit says otherwise.
const DefaultLimit = 5000

var _ bezel.Model[Pane] = Pane{}

// Pane is a log panel that follows the end as lines arrive and holds still
// while the reader scrolls up. / filters it as list and table do, but enter
// keeps the filter: a log has nothing to choose. Its copies share one feed.
type Pane struct {
	vp     viewport.Model
	feed   *feed
	th     theme.Resolved
	edit   func(PaneStyles) PaneStyles
	keys   PaneKeys
	recs   []record
	n      int // lines in recs
	shown  int // lines the filter keeps
	find   filterMode
	match  func(query, text string) bool // nil is Contains
	limit  int
	height int
	follow bool
	unseen int // lines that arrived while it held still
}

// filterMode is the / filter: the text typed, and whether the prompt is
// still open and taking the keys.
type filterMode struct {
	on   bool
	text string
}

// record is what one Write brought: a log record's lines, and its plain
// text, also in lower case, for the filter.
type record struct {
	lines      []string
	plain, low string
}

func newRecord(lines []string) record {
	plain := ansi.Strip(strings.Join(lines, "\n"))
	return record{lines: lines, plain: plain, low: strings.ToLower(plain)}
}

// Contains is the filter's default match: the record's text holds the
// query, in any case.
func Contains(query, text string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(query))
}

// PaneStyles are how a Pane paints: the lines its Handler writes, and its
// own parts around them.
type PaneStyles struct {
	Lines  Styles         // what the Handler writes
	More   lipgloss.Style // the count of lines that arrived while it held still
	Prompt lipgloss.Style // the / prompt while it takes the keys
	Filter lipgloss.Style // the filter kept after enter
	Match  lipgloss.Style // every match of the filter
}

// DefaultPaneStyles paint the lines as the Handler does, the count and the
// kept filter dim, the prompt as the cursor and the matches as matches.
func DefaultPaneStyles(th theme.Resolved) PaneStyles {
	return PaneStyles{Lines: DefaultStyles(th), More: th.Dim, Prompt: th.Cursor, Filter: th.Dim, Match: th.Match}
}

// PaneKeys are a Pane's bindings: the viewport's to scroll, and the / filter's.
type PaneKeys struct {
	Scroll             viewport.Keys
	Filter, Esc, Enter key.Binding
	Erase              key.Binding // the filter's last character
}

// DefaultPaneKeys are the viewport's DefaultKeys, / to filter, enter to keep
// the filter and esc to clear it.
func DefaultPaneKeys() PaneKeys {
	return PaneKeys{
		Scroll: viewport.DefaultKeys(),
		Filter: key.New("/").WithHelp("/", "filter"),
		Esc:    key.New("esc").WithHelp("esc", "clear filter"),
		Enter:  key.New("enter").WithHelp("enter", "keep filter"),
		Erase:  key.New("backspace", "ctrl+h"),
	}
}

// VimPaneKeys are DefaultPaneKeys scrolling with the viewport's VimKeys.
func VimPaneKeys() PaneKeys {
	k := DefaultPaneKeys()
	k.Scroll = viewport.VimKeys()
	return k
}

// feed is where the Handler's goroutines leave records for the pane to take
// in Update. Writing never blocks: wake holds at most one pending signal.
type feed struct {
	mu      sync.Mutex
	pending [][]string
	wake    chan struct{}
}

// paneLines is the message that carries the records a feed gathered.
type paneLines struct {
	feed *feed
	recs [][]string
}

// NewPane is an empty pane that draws its lines in th's colours, keeps the
// last DefaultLimit of them and wraps the long ones.
func NewPane(th theme.Resolved) Pane {
	p := Pane{
		vp:   viewport.New().SetWrap(true),
		feed: &feed{wake: make(chan struct{}, 1)},
		th:   th, keys: DefaultPaneKeys(), limit: DefaultLimit, follow: true,
	}
	return p.restyle()
}

// WithLimit keeps the last n lines, dropping the oldest; n below 1 is 1.
func (p Pane) WithLimit(n int) Pane {
	p.limit = max(n, 1)
	return p.trim().show()
}

// WithStyles edits the pane's styles; the lines' apply to what the Handler
// writes from now on.
func (p Pane) WithStyles(edit func(PaneStyles) PaneStyles) Pane {
	p.edit = edit
	return p.restyle()
}

// WithMatch filters with match, which gets the typed text and each record's
// lines, colours stripped; nil is Contains again. Matches are painted only
// under Contains, the one match that says where it found the text.
func (p Pane) WithMatch(match func(query, text string) bool) Pane {
	p.match = match
	return p.show()
}

// WithKeys rebinds the pane; Scroll.Bottom is also what follows the end again.
func (p Pane) WithKeys(k PaneKeys) Pane {
	p.keys, p.vp = k, p.vp.WithKeys(k.Scroll)
	return p
}

// styles are the pane's styles under its theme, edited.
func (p Pane) styles() PaneStyles {
	st := DefaultPaneStyles(p.th)
	if p.edit != nil {
		st = p.edit(st)
	}
	return st
}

// restyle hands the viewport the Match style: a filter has no match the
// reader is on, so every one is painted alike.
func (p Pane) restyle() Pane {
	match := p.styles().Match
	p.vp = p.vp.WithStyles(func(viewport.Styles) viewport.Styles {
		return viewport.Styles{Match: match, Current: match}
	})
	return p
}

// Handler writes into the pane, in colour and never wrapped, from any
// goroutine: slog.New(pane.Handler()) is a logger for it.
func (p Pane) Handler() Handler {
	lines := p.styles().Lines
	return New(p.feed, p.th).WithColor(true).WithWidth(-1).WithStyles(func(Styles) Styles { return lines })
}

// Writer is where any other handler writes into the pane, a pane line per
// line of text: slog.NewJSONHandler(pane.Writer(), nil) shows JSON lines.
func (p Pane) Writer() io.Writer { return p.feed }

// Write takes the formatted record and wakes the pane, without waiting.
func (f *feed) Write(b []byte) (int, error) {
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	f.mu.Lock()
	f.pending = append(f.pending, lines)
	f.mu.Unlock()
	select {
	case f.wake <- struct{}{}:
	default:
	}
	return len(b), nil
}

// Init is the command that waits for the Handler's first lines.
func (p Pane) Init() tea.Cmd { return p.listen() }

// listen waits for lines and hands every one gathered so far to Update.
func (p Pane) listen() tea.Cmd {
	f := p.feed
	return func() tea.Msg {
		for {
			<-f.wake
			f.mu.Lock()
			recs := f.pending
			f.pending = nil
			f.mu.Unlock()
			if len(recs) > 0 {
				return paneLines{feed: f, recs: recs}
			}
		}
	}
}

// Update takes the pane's own lines, the / filter's keys, and the viewport's
// keys and the mouse wheel to scroll. Scrolling up holds the pane still;
// reaching the end follows it. Lines arriving ask for the next in Result.Cmd.
func (p Pane) Update(msg tea.Msg) (Pane, bezel.Result) {
	if m, ok := msg.(paneLines); ok {
		if m.feed != p.feed {
			return p, bezel.Result{}
		}
		return p.add(m.recs), bezel.Result{Handled: true, Cmd: p.listen()}
	}
	if k, ok := msg.(tea.KeyPressMsg); ok {
		if next, done := p.filterKey(k); done {
			return next, bezel.Result{Handled: true}
		}
	}
	var r bezel.Result
	p.vp, r = p.vp.Update(msg)
	if r.Handled {
		p.follow = p.vp.AtBottom()
		if p.follow {
			p.unseen = 0
		}
		p = p.fit()
	}
	return p, r
}

// filterKey is the / filter: / opens the prompt, text edits it, enter keeps
// the filter, esc clears it. Other keys, the scrolling ones, are not its own.
func (p Pane) filterKey(k tea.KeyPressMsg) (Pane, bool) {
	ks := p.keys
	if !p.find.on {
		switch {
		case key.Matches(k, ks.Filter):
			p.find.on = true
			return p.fit(), true
		case p.find.text != "" && key.Matches(k, ks.Esc):
			return p.filter(filterMode{}), true
		}
		return p, false
	}
	switch {
	case key.Matches(k, ks.Esc):
		return p.filter(filterMode{}), true
	case key.Matches(k, ks.Enter):
		p.find.on = false
		return p.fit(), true
	case key.Matches(k, ks.Erase):
		return p.filter(filterMode{on: true, text: edit.Erase(p.find.text)}), true
	case edit.Typed(k) != "":
		return p.filter(filterMode{on: true, text: p.find.text + edit.Typed(k)}), true
	}
	return p, false
}

// filter shows the records f keeps and follows the end.
func (p Pane) filter(f filterMode) Pane {
	p.find, p.follow, p.unseen = f, true, 0
	p = p.show().fit()
	p.vp = p.vp.ToBottom()
	return p
}

// keeps reports whether the filter shows r.
func (p Pane) keeps(r record) bool {
	switch q := p.find.text; {
	case q == "":
		return true
	case p.match != nil:
		return p.match(q, r.plain)
	}
	return strings.Contains(r.low, strings.ToLower(p.find.text))
}

// show hands the viewport the lines the filter keeps, its matches painted.
func (p Pane) show() Pane {
	lines := make([]string, 0, p.n)
	for _, r := range p.recs {
		if p.keeps(r) {
			lines = append(lines, r.lines...)
		}
	}
	p.shown = len(lines)
	p.vp = p.vp.SetLines(lines)
	if p.find.text == "" || p.match != nil {
		p.vp = p.vp.ClearHighlights()
		return p
	}
	y := p.vp.YOffset()
	p.vp = p.vp.SetHighlights(matches(p.vp.Content(), p.find.text)).SetYOffset(y)
	return p
}

// matches is where q falls in content's plain text, in any case, as byte
// ranges; none where lower case changes the text's length.
func matches(content, q string) [][]int {
	plain := ansi.Strip(content)
	low, q := strings.ToLower(plain), strings.ToLower(q)
	if len(low) != len(plain) || q == "" {
		return nil
	}
	var out [][]int
	for at := 0; ; {
		i := strings.Index(low[at:], q)
		if i < 0 {
			return out
		}
		out = append(out, []int{at + i, at + i + len(q)})
		at += i + len(q)
	}
}

// add appends records, dropping the oldest lines past the limit, and keeps
// to the end while following, else to the line the reader is on: wrapped
// lines take several rows, so the place is kept by line, not by row.
func (p Pane) add(batch [][]string) Pane {
	before, added := p.shown, 0
	top, part := p.vp.TopLine()
	recs := slices.Clip(p.recs)
	for _, lines := range batch {
		r := newRecord(lines)
		recs = append(recs, r)
		p.n += len(lines)
		if p.keeps(r) {
			added += len(lines)
		}
	}
	p.recs = recs
	p = p.trim().show()
	dropped := max(before+added-p.shown, 0)
	if !p.follow {
		p.unseen += added
	}
	p = p.fit()
	switch {
	case p.follow:
		p.vp = p.vp.ToBottom()
	case top < dropped:
		p.vp = p.vp.SetTopLine(0, 0)
	default:
		p.vp = p.vp.SetTopLine(top-dropped, part)
	}
	return p
}

// trim keeps the last limit lines, cutting into the oldest record kept.
func (p Pane) trim() Pane {
	over := p.n - p.limit
	if over <= 0 {
		return p
	}
	i := 0
	for over >= len(p.recs[i].lines) {
		over -= len(p.recs[i].lines)
		i++
	}
	recs := slices.Clone(p.recs[i:])
	if over > 0 {
		recs[0] = newRecord(recs[0].lines[over:])
	}
	p.recs, p.n = recs, p.limit
	return p
}

// prompt reports whether the / prompt's row is drawn: while typing, and
// while a kept filter applies.
func (p Pane) prompt() bool { return p.find.on || p.find.text != "" }

// fit gives the viewport the pane's height, less the rows that count what
// arrived while it held still and show the filter.
func (p Pane) fit() Pane {
	h := p.height
	if p.unseen > 0 {
		h--
	}
	if p.prompt() {
		h--
	}
	p.vp = p.vp.SetHeight(max(h, 0))
	return p
}

// SetWidth is the pane's width in cells.
func (p Pane) SetWidth(w int) Pane {
	p.vp = p.vp.SetWidth(w)
	return p
}

// SetHeight is the rows the pane takes; following, it stays on the end.
func (p Pane) SetHeight(h int) Pane {
	p.height = h
	p = p.fit()
	if p.follow {
		p.vp = p.vp.ToBottom()
	}
	return p
}

// Following reports whether the pane keeps to the end as lines arrive.
func (p Pane) Following() bool { return p.follow }

// Len is how many lines the pane holds, the filtered out ones included.
func (p Pane) Len() int { return p.n }

// Shown is how many lines the filter keeps; Len while there is none.
func (p Pane) Shown() int { return p.shown }

// Filtering reports whether the / prompt is open and taking the keys.
func (p Pane) Filtering() bool { return p.find.on }

// Filter is the text the shown records hold, typed at the / prompt.
func (p Pane) Filter() string { return p.find.text }

// Lines is a copy of the lines the pane holds, colours included.
func (p Pane) Lines() []string {
	lines := make([]string, 0, p.n)
	for _, r := range p.recs {
		lines = append(lines, r.lines...)
	}
	return lines
}

// Scroll is where the pane shows its lines from, for a pane's scrollbar.
func (p Pane) Scroll() layout.Scroll { return p.vp.Scroll() }

// Legend lists the scrolling keys and the filter's; the prompt's own while
// it is open.
func (p Pane) Legend() []legend.Entry {
	k := p.keys
	if p.find.on {
		return legend.FromBindings([]key.Binding{k.Enter, k.Esc})
	}
	keys := []key.Binding{k.Filter}
	if p.find.text != "" {
		keys = append(keys, k.Esc)
	}
	return append(p.vp.Legend(), legend.FromBindings(keys)...)
}

// View is the lines in the viewport; under them, while the pane holds still,
// how many arrived since, and the filter while there is one.
func (p Pane) View() string {
	w := p.vp.Width()
	if p.height <= 0 || w <= 0 {
		return ""
	}
	st := p.styles()
	var rows []string
	if v := p.vp.View(p.th); v != "" {
		rows = strings.Split(v, "\n")
	}
	if p.unseen > 0 {
		unit := "lines"
		if p.unseen == 1 {
			unit = "line"
		}
		rows = append(rows, st.More.Render(draw.Fit(fmt.Sprintf("↓ %d new %s", p.unseen, unit), w)))
	}
	switch {
	case p.find.on:
		rows = append(rows, st.Prompt.Render(draw.Fit("/"+p.find.text+"▋", w)))
	case p.find.text != "":
		rows = append(rows, st.Filter.Render(draw.Fit("/"+p.find.text, w)))
	}
	return strings.Join(rows[max(len(rows)-p.height, 0):], "\n")
}
