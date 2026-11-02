package logger

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/viewport"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// DefaultLimit is the lines a Pane keeps until WithLimit says otherwise.
const DefaultLimit = 5000

var _ bezel.Model[Pane] = Pane{}

// Pane is a log panel: the lines its Handler writes, in a viewport that
// follows the end as they arrive and holds still while the reader scrolls
// up. A value: every method returns the new pane; its copies share one feed.
type Pane struct {
	vp     viewport.Model
	feed   *feed
	th     theme.Resolved
	edit   func(Styles) Styles
	lines  []string
	limit  int
	height int
	follow bool
	unseen int // lines that arrived while it held still
}

// feed is where the Handler's goroutines leave lines for the pane to take
// in Update. Writing never blocks: wake holds at most one pending signal.
type feed struct {
	mu      sync.Mutex
	pending []string
	wake    chan struct{}
}

// paneLines is the message that carries the lines a feed gathered.
type paneLines struct {
	feed  *feed
	lines []string
}

// NewPane is an empty pane that draws its lines in th's colours, keeps the
// last DefaultLimit of them and wraps the long ones.
func NewPane(th theme.Resolved) Pane {
	return Pane{
		vp:   viewport.New().SetWrap(true),
		feed: &feed{wake: make(chan struct{}, 1)},
		th:   th, limit: DefaultLimit, follow: true,
	}
}

// WithLimit keeps the last n lines, dropping the oldest; n below 1 is 1.
func (p Pane) WithLimit(n int) Pane {
	p.limit = max(n, 1)
	return p.trim()
}

// WithStyles edits the styles of the lines Handler writes from now on.
func (p Pane) WithStyles(edit func(Styles) Styles) Pane {
	p.edit = edit
	return p
}

// WithKeys rebinds the scrolling; Bottom is also what follows the end again.
func (p Pane) WithKeys(k viewport.Keys) Pane {
	p.vp = p.vp.WithKeys(k)
	return p
}

// Handler writes into the pane, in colour and never wrapped, from any
// goroutine: slog.New(pane.Handler()) is a logger for it.
func (p Pane) Handler() Handler {
	h := New(p.feed, p.th).WithColor(true).WithWidth(-1)
	if p.edit != nil {
		h = h.WithStyles(p.edit)
	}
	return h
}

// Writer is where any other handler writes into the pane, a pane line per
// line of text: slog.NewJSONHandler(pane.Writer(), nil) shows JSON lines.
func (p Pane) Writer() io.Writer { return p.feed }

// Write takes the formatted record and wakes the pane, without waiting.
func (f *feed) Write(b []byte) (int, error) {
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	f.mu.Lock()
	f.pending = append(f.pending, lines...)
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
			lines := f.pending
			f.pending = nil
			f.mu.Unlock()
			if len(lines) > 0 {
				return paneLines{feed: f, lines: lines}
			}
		}
	}
}

// Update takes the pane's own lines, and the viewport's keys and the mouse
// wheel to scroll. Scrolling up holds the pane still; reaching the end again
// follows it. Lines arriving ask for the next ones in Result.Cmd.
func (p Pane) Update(msg tea.Msg) (Pane, bezel.Result) {
	if m, ok := msg.(paneLines); ok {
		if m.feed != p.feed {
			return p, bezel.Result{}
		}
		return p.add(m.lines), bezel.Result{Handled: true, Cmd: p.listen()}
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

// add appends lines, dropping the oldest past the limit, and keeps to the
// end while following, else to the lines the reader is on.
func (p Pane) add(lines []string) Pane {
	p.lines = append(slices.Clip(p.lines), lines...)
	dropped := max(len(p.lines)-p.limit, 0)
	p = p.trim()
	if !p.follow {
		p.unseen += len(lines)
	}
	p = p.fit()
	if p.follow {
		p.vp = p.vp.ToBottom()
	} else {
		p.vp = p.vp.SetYOffset(p.vp.YOffset() - dropped)
	}
	return p
}

// trim keeps the last limit lines and hands them to the viewport.
func (p Pane) trim() Pane {
	if n := len(p.lines) - p.limit; n > 0 {
		p.lines = slices.Clone(p.lines[n:])
	}
	p.vp = p.vp.SetLines(p.lines)
	return p
}

// fit gives the viewport the pane's height, less the line that counts what
// arrived while it held still.
func (p Pane) fit() Pane {
	h := p.height
	if p.unseen > 0 {
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

// Len is how many lines the pane holds.
func (p Pane) Len() int { return len(p.lines) }

// Lines is a copy of the lines the pane holds, colours included.
func (p Pane) Lines() []string { return slices.Clone(p.lines) }

// Scroll is where the pane shows its lines from, for a pane's scrollbar.
func (p Pane) Scroll() layout.Scroll { return p.vp.Scroll() }

// Legend lists the scrolling keys.
func (p Pane) Legend() []legend.Entry { return p.vp.Legend() }

// View is the lines in the viewport, and under them, while the pane holds
// still, how many arrived since.
func (p Pane) View() string {
	v := p.vp.View(p.th)
	if p.unseen == 0 || p.height <= 0 || p.vp.Width() <= 0 {
		return v
	}
	unit := "lines"
	if p.unseen == 1 {
		unit = "line"
	}
	more := p.th.Dim.Render(draw.Fit(fmt.Sprintf("↓ %d new %s", p.unseen, unit), p.vp.Width()))
	if v == "" {
		return more
	}
	return v + "\n" + more
}
